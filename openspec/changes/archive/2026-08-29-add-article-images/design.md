## Context

The article reader converts HTML body content to markdown (`convert.Convert`)
and renders it to styled terminal text with glamour at the content width
(`internal/ui/markdown.go`), wrapping inline `<img>` elements as
`[alt]`/`[image]` placeholders. The rendered text feeds a
`bubbles/viewport.Model` that scrolls by line; the viewport is framed by
`renderArticleBorder` (`internal/ui/border.go`). Article data flows from
`gofeed` → `itemToArticle` (`internal/feed/fetch.go`) → SQLite
(`internal/store`), with the schema created by `migrate()` using `CREATE TABLE
IF NOT EXISTS` only (no additive migrations yet). Config lives in
`internal/config` with a `Display` struct and `-a`/`--ascii` flag forcing
Unicode-fallback glyphs.

`blacktop/go-termimg` (MIT) provides multi-protocol terminal image rendering
with a Bubbletea-aware `StatefulImageWidget`/`AsyncRenderWorker` and a
halfblocks renderer backed by `charmbracelet/x/mosaic`. Its README warns that
Kitty Unicode-placeholder/relative-placement features are under construction
and not recommended for production — this directly shapes the protocol
strategy below.

## Goals / Non-Goals

**Goals:**
- Render publisher-attached (enclosure) lead images inline in the scrolling
  article body, scrolling with the text.
- Ship a universally-correct halfblocks path that works on every
  Unicode-capable terminal.
- Attempt a native-protocol (Kitty placeholder) upgrade within the same change,
  gated on a decisive spike, without making the feature depend on it.
- Keep the change privacy-safe: only publisher-hosted URLs are fetched, never
  inline content `<img>` URLs.

**Non-Goals:**
- Inline body-image (`<img>` in content) rendering — separate future change.
- Image cache pruning/retention — stored image bytes grow with articles;
  eviction and article-retention pruning are future work.
- Image formats beyond Go's `image` package (PNG/JPEG/GIF); WebP when
  go-termimg ships it.
- Native-protocol as a guaranteed deliverable — it is an attempted upgrade,
  descoped if the spike fails.

## Decisions

### D1: Inline placement, not fixed banner

The lead image is composed into the article viewport content as a block of text
lines (halfblocks) spliced above the glamour-rendered header+body, then handed
to the existing `bubbles/viewport` unchanged. The image scrolls with the body
because it IS body content.

**Alternative considered:** a fixed-height banner band above the viewport
(which would allow native-protocol painting at a stable position). Rejected
because it does not match the web-article model the user wants, and it would
require restructuring `renderArticleBorder` and splitting the viewport. Inline
halfblocks requires zero viewport/border surgery — the image is just more
content lines.

### D2: Halfblocks is the committed render path; native is a gated spike

Halfblocks output is real text-grid content (Unicode half-block characters +
ANSI SGR color). `bubbles/viewport` slices it, `lipgloss` width-measures it,
and bubbletea's frame-diff handles it — all natively, with no escape-sequence
lifecycle concerns. It works on every Unicode-capable terminal at ~800µs/
render. The cost is fidelity: 2 vertical pixels per cell.

Native protocol (Kitty/Sixel/iTerm2) paints images at a cursor position via
escape sequences that the viewport's text-slice model cannot carry — except
Kitty's Unicode-placeholder mode, which binds an image to a placeholder cell
so it scrolls with text. go-termimg exposes this via `UseUnicode(true)`+
`Virtual(true)`, but marks it under construction.

**Spike pass criteria** (all must hold, else defer native and ship
halfblocks-only):
1. Image displays inline within the viewport via Kitty placeholder.
2. Scrolling moves the image with text — no ghost at the old position.
3. Unchanged-frame re-render (bubbletea diff) does not duplicate/flicker.
4. Resize re-renders cleanly, no artifacts.
5. Exiting the article view clears the image (no ghost on list view).
6. go-termimg's API supports separating **transmission** (upload once) from
   **placement** (per-frame) without re-uploading image data each tick — or a
   documented workaround exists. (This is the suspected real gate: a one-shot
   `Render()` likely bundles transmission+placement.)

If the spike passes, native (Kitty) is preferred when supported with halfblocks
as the automatic fallback; protocol is detected once via
`termimg.DetectProtocol()`/`QueryTerminalFeatures()` and cached on the Model.
If it fails, the findings are documented here and halfblocks ships alone.

**Spike outcome (documented at implementation time):** the native-protocol
spike did not pass and native-inline is descoped to future work. Findings:

- **Criterion 6 fails on the API surface.** go-termimg's widget model is
  one-shot: `ImageWidget.Render()` re-renders (and for placeholder placement,
  re-transmits) the image every frame. The README documents no
  transmit-once/place-per-frame split, and the Kitty Unicode-placeholder /
  relative-placement features are explicitly marked "under construction and
  not recommended for production use". A one-shot `Render()` inside
  `bubbles/viewport` would re-upload image data on every frame, defeating the
  transmission/placement separation the spike required.
- **Criteria 1-5 could not be exercised end-to-end in this environment** (an
  interactive Kitty terminal with the running TUI is required). Given
  criterion 6 already fails decisively on the documented API, the spike is
  concluded as failed rather than deferred.
- **Halfblocks ships as the sole render path** (the committed default), which
  is the design's escape hatch: it loses only fidelity (2 vertical pixels per
  cell), not the feature, and requires no protocol state, so scrolling and
  view-exit leave no ghost.

### D3: Atomic scroll via a pure snap function

The image block occupies a contiguous range `[imgStart, imgEnd]` of content
lines. `bubbles/viewport` has no concept of atomic blocks, so a pure function
snaps the offset out of the interior after every scroll mutation:

```
snapYOffset(yOffset, vpH, imgStart, imgEnd, direction int) int
  down (direction>0) into interior → imgEnd + 1   (skip past)
  up   (direction<0) into interior → imgStart      (reveal full)
  otherwise                            → yOffset     (no change)
```

Called after `ScrollDown/Up`, `PageDown/Up`, `HalfPage*`, `GotoBottom` in
`updateArticle`, then `viewport.SetYOffset(snapped)`. Direction is the sign of
the offset delta. Fully unit-testable as a table-driven pure function.

**Prerequisite:** the image block height is capped so the full image plus one
line of body fits in `viewportH` — otherwise the interior is unescapable
(image taller than viewport can never be fully visible). The cap is
`viewportH - headerLines - 1`.

### D4: DB-backed image cache, async fetch, message-driven compose

Images are remote; the render path never blocks on the network. The cache
hierarchy is three tiers:

```
open article with ImageURL
  │
  ├─ in-memory decoded cache (map[URL]image.Image)  ◄── session-local, fastest
  │     hit → decode done, render
  │     miss ↓
  ├─ DB image_data BLOB                              ◄── survives restart
  │     hit → image.Decode(bytes), cache in memory, render
  │     miss ↓
  └─ network fetch (tea.Cmd)                         ◄── contacted ONCE, ever
        success → decode, persist raw bytes to DB, cache in memory, imageLoadedMsg
        failure → imageFailedMsg (no persist)
```

Raw compressed bytes (JPEG/PNG as fetched) are stored in `articles.image_data`,
not decoded pixels — re-decode on DB hit is cheap and keeps storage small.
`openArticle` fires the load command; on success `imageLoadedMsg{key, img}`
triggers re-compose + re-render; on failure `imageFailedMsg{key}` renders
without a block. The in-memory decoded cache means resizes reuse the decoded
image (resize only re-renders the halfblocks output, ~800µs) without re-decoding
from the DB. No eviction policy in v1; DB growth is bounded per-image by the
fetch max-bytes cap and grows with article count (see Risks).

Fetch guards: HTTP timeout, max-bytes cap (reject oversized responses),
`image/*` content-type allowlist, `image.Decode` for format validation.

### D5: Composition splice above the glamour-rendered content

`newArticleState` composes the viewport content from two independently produced
parts. The image block is a pre-rendered string of halfblock lines sized to
exactly `contentW`; the article text is the glamour-rendered markdown
(`renderArticleMarkdown` header + body). The composition order is:

```
[image block lines]        ← pre-sized to contentW, aspect-preserved, capped
[glamour-rendered content] ← header (title/author/link) + wrapped body
```

The image block is never passed through the markdown renderer or re-wrapped —
it is already line-broken to the width. glamour renders and wraps the body to
`contentW`. On `WindowSizeMsg`, `contentW` changes; the block is re-rendered
from the cached image at the new width and the content is re-composed.

### D6: Scroll-position preservation on late load

When `imageLoadedMsg` arrives and the user has scrolled past the insertion
point (`YOffset > imgStart`), the inserted lines push their reading position
down. The handler bumps `YOffset` by the inserted line count before
re-composing so the visible text is unchanged. If the user is at or above the
insertion point, no adjustment is needed — the image simply appears above.

### D7: Guarded additive migration

SQLite lacks `ADD COLUMN IF NOT EXISTS`. The migration queries
`PRAGMA table_info(articles)`, checks for `image_url` and `image_data`, and
issues `ALTER TABLE articles ADD COLUMN ...` for each absent column
(`image_url TEXT`, `image_data BLOB`). This is the project's first additive
migration; `migrate()` currently only does `CREATE TABLE IF NOT EXISTS`, so
the guard pattern is new but localized.

### D8: Renderer interface seam for testability

Image output is terminal-dependent and not assertable in unit tests. The
renderer is behind an interface so a fake returns deterministic strings; tests
cover URL extraction, cache hit/miss, config/ascii gating, composition
ordering, `snapYOffset` table cases, scroll preservation, fetch-failure
fallback, and resize re-render — all without rendering real images.

## Risks / Trade-offs

- **[Halfblocks fidelity is blocky]** → Acceptable for lead thumbnails; native
  is the gated upgrade path. A 76-wide banner is 76×(2×rows)px — recognizable.
- **[Native spike may fail (criterion 6)]** → Halfblocks ships regardless; the
  spike is scoped so failure only drops fidelity, not the feature. Findings get
  documented and native-inline becomes a future change.
- **[Image fetch reveals reading activity to image host]** → Mitigated on two
  fronts: enclosure-only source selection (host is the publisher yerss already
  contacted for the feed; no third-party CDN/tracker URLs), and DB persistence
  means each image host is contacted exactly once per article, ever — no
  repeated requests on every restart. `auto` default is safe under this
  constraint.
- **[DB storage growth from image bytes]** → Bounded per-image by the fetch
  max-bytes cap; grows with article count since there is no retention pruning
  today (~900MB/month for an active user). Acceptable for v1; pruning/retention
  is future work, likely tied to an article-retention feature.
- **[Late image load shifts content]** → Mitigated by D6 scroll-position
  preservation. If the user is at the top, the image appears above with no
  shift.
- **[go-termimg dependency churn]** → Pinned version in go.mod; the halfblocks
  path depends only on `charmbracelet/x/mosaic` which is stable. Native path
  is the only consumer of the WIP placeholder API.

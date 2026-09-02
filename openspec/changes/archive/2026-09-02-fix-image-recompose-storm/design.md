## Context

The article view re-composes from scratch on every image-load message. Each
recompose calls `newArticleState`, which re-runs `convert.ConvertImages` (whole
HTML → marked markdown), glamour `renderMarkdown` over every segment, link
harvesting, dedupe passes, and viewport rebuild — all synchronously on the
bubbletea update goroutine. A photo-dense article fires ~2 messages per image
(Block + Photo/Native), so the UI is repeatedly frozen while images load, and a
Back keypress queues behind the storm.

See proposal.md — Why for the observed impact. Current relevant state:

- `articleState` holds the fully-composed `lines`, `imageBlocks`, `imageURLs`,
  `inlineImages`, `inlineCaps`, `links`, and a `viewport` built from `lines`.
- `newArticleState` derives everything from `a.Content` + geometry each call.
- `recomposeArticle` (image.go) rebuilds via `newArticleState` and re-applies
  the reading offset with `SnapAfterRecompose`.
- `computeNativeClear` and the image-snap invariants depend on `imageBlocks`
  line indices staying in sync with `lines` — any caching layer must keep block
  indices valid.
- `loadList` runs `ListArticles` (selects full `content`) + `ArticleCount` +
  `DBSize` on every back-out, refresh, filter change, and startup.

## Goals / Non-Goals

**Goals:**

- Eliminate full-body derivation re-runs on image-load recompose.
- Bound per-image-message UI work; coalesce a batch of landings into one
  recompose.
- Stop the list query from copying article bodies.
- Preserve all current rendering, scrolling, snap, native-clear, and read-state
  behavior byte-for-byte for the article's displayed output.

**Non-Goals:**

- Off-thread rendering of the markdown/glamour pipeline (out of scope; the
  cache removes the repeated cost without a threading redesign).
- Image-cache eviction or memory-bounding of the new derivation cache.
- Changing the article's visual output, keybindings, or scroll-snap rules.

## Decisions

### D1: Derive once per (article, geometry), recompose by re-joining

Split `newArticleState` into a **derivation** stage and a **composition**
stage. The derivation converts `Content` → marked markdown, discovers inline
images, harvests links, and renders each markdown segment through glamour —
the expensive, content-only work. It is cached on the model keyed by
`(articleID, contentW, vpH)` (geometry matters: glamour wraps at `contentW` and
the viewport height feeds the image-height fit). The composition stage renders
image blocks (cheap: served from `ImgNatives`/`ImgBlocks`/`ImgCache`), splices
them into the cached rendered segments at the sentinel positions, and builds
the viewport — the part that must run per recompose because image blocks
change.

**Alternatives considered:**
- *Cache the final composed `lines` string and re-splice on change.* Too coarse:
  image block insertion shifts every subsequent line index, so splicing into a
  flat string requires re-walking it anyway and loses the segment structure the
  inline sentinels provide.
- *Incremental delta of the existing `lines`.* Fragile: block insertions move
  indices across the whole tail; the snap machinery and `imageBlocks`
  coordinates would need an offset fix-up pass, which is exactly the bug-prone
  surface the segment approach avoids.
- *Off-thread full recompose per message.* Removes the freeze but reintroduces
  ordering/races with the single-threaded tea loop and still wastes CPU; the
  cache makes recompose cheap enough to stay synchronous.

### D2: Cache the derivation on the Model, not in Session

The derivation is tied to the open article and the current palette/ascii/theme
(the glamour renderer and styles are model-owned). Holding it on `Model`
(alongside `mdRenderer`) keeps the cache lifetime bounded to the article and
geometry, and a `WindowSizeMsg` / `openArticle` naturally replaces it. No
cross-session lifetime is needed.

Key: `articleID` + `contentW` + `vpH`. A cache hit is only valid when the
article id matches the open article, so a stale entry from a previous article
at the same geometry can never be spliced into the wrong article.

### D3: Recompose once per image message with the cached derivation

bubbletea (v1.3.10) delivers every message in a `tea.Batch` as a separate
`Update` invocation with a render between, so "one recompose per batch" is not
achievable by per-message Update logic. Instead, in `Update`, tag each
image-load message (`image.BlockMsg`, `PhotoMsg`, `NativeMsg`, `FailedMsg`)
with a "recompose pending" flag on the model rather than recomposing inline.
After the message switch, if the flag is set and the view is still the article
view, run `recomposeArticle` once and clear the flag. The per-message work is
therefore bounded to cache updates plus marking; the recompose runs at most
once per message, after it, and — because it reuses the cached derivation from
D1/D2 — never re-runs HTML-to-markdown conversion, markdown rendering, or link
harvesting on the UI thread. The reading position is preserved exactly as
today (`SnapAfterRecompose`).

**Trade-off:** a recompose now runs after the message's cache work rather than
inline mid-handler; this is not observable because the recompose lands in the
same update round and the terminal coalesces the frame writes. The message
handlers stay free of direct recompose calls so future coalescing can be
layered on without touching them.

**Why not aggregate loads into one message:** a single aggregated image-load
message (one recompose per whole open) would make every image of a photo-dense
article appear only when the last load lands, regressing the progressive
appearance each image gets today. Per-message recompose with the cached
derivation keeps progressive appearance and the reading-position behavior.

### D4: Slim the list-load query with a dedicated lite method

Add `ListArticlesLite` (store.go + session) that selects only
`id, feed_url, guid, title, link, author, published_at, read, fetched_at,
image_url` — omitting `content` and `description`, which the list never renders.
The UI's `loadList` uses it, so returning to the list view stops copying every
article body. `ListArticles` keeps its full SELECT: `cmd/yerss/json.go` (the
`--json` export) reads `Content`/`Description` from its rows, and
`TestJSONDumpArticleFieldsAndCategories` asserts the body is present, so the
export path must keep full rows. `scanListArticle` reads the slim column list
(leaving `Content`/`Description` empty on the returned rows) while
`scanArticle` stays the full-row scan shared with `GetArticle`. Both queries
share `listQuery` so their ordering and tag-filter shapes cannot drift.

## Risks / Trade-offs

- [Cached derivation goes stale if the model's renderer, palette, or ascii mode
  changes while an article is open] → Key the cache on geometry AND invalidate
  it whenever `mdRenderer` is rebuilt (width/theme change) or `ascii`/palette
  changes; treat those as a "derivation epoch" that clears the cache.
- [Inline sentinel positions or dedupe interactions make the spliced segments
  diverge from today's `newArticleState` output] → Keep the segment structure
  byte-identical to what `inlineBodyParts` walks today; add a test that compares
  the cached-derivation composition to the pre-change `newArticleState` output
  for a fixture with lead + inline + duplicate-sentinel cases.
- [An image message lands together with the Back keypress, so the recompose
  could run after the view switched to list] → The flag is cleared and skipped
  when `m.view != viewArticle`; image caches are updated regardless (as today),
  so re-opening the article renders the block.
- [List query slimming breaks a caller that expects `Content`] → The JSON
  export reads body content from `ListArticles`, so `ListArticles` keeps full
  rows and only the new `ListArticlesLite` path (the UI list) is slim; the
  article view loads the full row via `GetArticle`.
- [Cache holds one full article's derived segments in memory] → Single-entry,
  replaced on article/geometry change; bounded like the existing `mdRenderer`.

## Migration Plan

- Implementation order: D4 first (small, independent), then D1+D2 (derivation
  cache), then D3 (per-message recompose via the cached derivation). Each lands
  with its tests green; no schema or external dependency changes, so no data
  migration.

## Open Questions

None that change the specs or task breakdown. Whether to also cache the
harvested links across recomposes is subsumed by D1 (links are part of the
derivation).
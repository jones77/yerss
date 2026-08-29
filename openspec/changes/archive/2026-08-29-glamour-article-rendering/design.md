## Context

The article reader currently renders HTML to markdown with hand-rolled ANSI
escapes (`convert.Convert` embeds `\x1b[1m` bold and OSC 8 `[text](url)`
links) and then link-aware soft-wraps via `convert.WrapText` (see
`internal/ui/article.go`, `convert/`). The spec (`openspec/specs/reader-ui`)
mandates that exact visible format. See proposal.md for motivation.

Constraints shaping the approach:

- `convert.Convert` and `convert.WrapText` have a single UI call site
  (`internal/ui/article.go`); the `convert` package also holds unit tests
  asserting the current escape-embedded output.
- `Display.Theme` (`auto`/`dark`/`light`) already drives the TUI palette via
  `resolvePalette` + `detectLightBackground()` in `internal/ui/theme.go`.
- The article viewport (`bubbles/viewport`) renders pre-wrapped lines; mouse
  selection strips styling with `ansi.Strip`/`Cut`/`Truncate`, which already
  handle ANSI/OSC 8 output, so styled glamour output is compatible.
- glamour v2 (`charm.land/glamour/v2`) always emits OSC 8 hyperlinks and
  ANSI styling regardless of TTY (link.go `makeHyperlink`, baseelement.go
  `renderText`), so output is deterministic enough for tests to assert on
  visible text after `ansi.Strip`.

## Goals / Non-Goals

**Goals:**
- Render article markdown through glamour at the content width.
- Drive glamour's theme from `display.theme`, defaulting to glamour's built-in
  dark/light styles.
- Keep existing reader chrome (border, scrollbar thumb, mouse selection,
  read-on-scroll) working on glamour output.
- Minimal blast radius on `convert`: keep `WrapText` exported/tested, unused by
  the article path (the in-progress `add-article-images` change will rebase
  onto this pipeline later).

**Non-Goals:**
- Custom glamour stylesheets (built-in dark/light are the defaults; `auto`
  resolves via the existing light-background detection).
- Preserving the old `[text](url)` visible link format or URL ellipsis
  truncation (superseded by glamour's native rendering — see specs).
- Wiring glamour into the list view, popups, or any non-article surface.

## Decisions

**D1. `convert.Convert` emits plain commonmark.**
Remove the custom `<strong>`/`<b>`/`<a>` renderers that embed ANSI (convert.go
`renderBold`, `renderLink`) so the converter output is plain markdown
(`**bold**`, `[text](url)`) ready for glamour's goldmark parser. Keep the
`<img>` → `[alt]`/`[image]` renderer (glamour shows it as literal text, matching
the spec) and the `collapseBlankLines` post-renderer. Alternative considered:
keeping escapes and letting glamour parse raw ANSI — rejected because embedded
SGR/OSC 8 inside markdown would be styled/tokenized unpredictably and the
converter tests would diverge from what glamour actually receives.

**D2. Theme resolution via a small `markdown.go` seam.**
A helper on `Model` builds a `*glamour.TermRenderer` and caches it keyed by
(width, resolved style). Style mapping: `dark` → `"dark"`, `light` → `"light"`,
`auto` → `detectLightBackground() ? "light" : "dark"` (same detector as the
palette, so the whole UI agrees). Renderer options: `WithStandardStyle(style)` +
`WithWordWrap(contentW)`. Rebuild when width or style changes (resize path in
`newArticleState` re-renders anyway). Alternative considered: glamour's own
`WithAutoStyle()` / `"auto"` string — rejected for consistency: the palette and
markdown should resolve `auto` from the same signal.

**D3. Header rendered through glamour as markdown.**
`renderArticleContent` becomes `renderArticleMarkdown`, emitting
`**Title**`, `*by Author*`, `[url](url)` (when present), a blank line, then the
converted body. glamour styles the title/author/link consistently and the link
stays OSC 8 clickable. Alternative considered: keeping the plain header block
above the glamour body — rejected for visual consistency per the decided
direction.

**D4. glamour owns wrapping; `convert.WrapText` stays unused by the article path.**
`WithWordWrap(contentW)` makes glamour wrap to the content width. Trim
glamour's trailing newline before `viewport.SetContent`. `WrapText` and its
tests remain in the package (exported, still valid for other callers and the
upcoming image change). Alternative considered: deleting `WrapText` — rejected
to keep the diff surgical and avoid churning the image change's assumptions.

**D5. Copy Article Text copies visible rendered text.**
The `CopyArticleText` action strips ANSI/OSC 8 from the rendered viewport
content (what the user sees), not the markdown source. Links copy as
`text url`. Alternative: copying markdown source — rejected; users expect the
visible text.

## Risks / Trade-offs

- **Glamour's `document.margin: 2`** adds a 2-column left margin inside the
  padX=2 content area, shifting column math and several render tests.
  → Recalibrate affected tests to the new content layout; margin is part of the
  default theme we chose to keep.
- **Exact styling bytes differ from the old hand-rolled escapes**; tests that
  assert on `\x1b[1m…\x1b[22m` break.
  → Update converter tests to assert plain markdown and UI tests to assert
  visible (stripped) text plus presence of OSC 8 where relevant.
- **glamour's color output depends on the x/ansi color profile**; in a no-TTY
  test environment colors may be downsampled.
  → Assert on `ansi.Strip`ped visible text and on OSC 8 presence/absence
  conditionally; avoid asserting exact SGR bytes.
- **glamour re-renders on every resize/open** (goldmark parse per render).
  → Content is article-sized; measured cost is acceptable. The renderer (style
  parse) is cached; only width changes force a rebuild.
- **Future `add-article-images` interaction**: that change assumed
  `wrapText`/`renderArticleContent`. It is deliberately sequenced after this
  change and its section 5 will rebase onto the glamour pipeline.
  → No action now; documented in tasks.

## Migration Plan

No data migration. `display.theme` semantics are unchanged; existing configs
keep working. Rollback is a source revert — the change is self-contained to
the article rendering path and the `convert` package.

## Open Questions

None.
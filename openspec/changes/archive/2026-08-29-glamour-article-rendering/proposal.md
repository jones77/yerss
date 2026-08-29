## Why

The article reader renders HTML as markdown with hand-rolled ANSI escapes
(bold, OSC 8 links) and a custom link-aware soft wrapper. Charm's glamour
library produces polished, themeable terminal markdown for free; wiring it in
lets yerss match the look of tools like Glow and delegate link/emphasis/wrapping
handling to a maintained renderer.

## What Changes

- **BREAKING**: Article HTML content is converted to plain markdown (no
  embedded ANSI) and rendered to styled terminal text by glamour at the
  content width. The theme is driven by the existing `display.theme`
  (`auto`/`dark`/`light`): glamour's built-in dark and light themes are the
  defaults, and `auto` resolves via the existing light-background detection.
- **BREAKING**: Links render in glamour's native form: styled link text
  followed by the URL (`text url`), OSC 8 clickable when the terminal supports
  it. Overlong URLs wrap across lines instead of being truncated with the
  project's ellipsis glyph. The custom `[text](url)` + OSC 8 wrap and the URL
  ellipsis-truncation behavior are removed.
- **BREAKING**: `<strong>`/`<b>` no longer emit raw `\x1b[1m` escapes from the
  converter; bold (and emphasis, headings, lists, blockquotes, code, tables) is
  styled by glamour's theme.
- The reader header (title, author, link) is rendered through glamour as
  markdown so it is styled consistently with the body; the header URL remains
  OSC 8 clickable.
- glamour handles word wrapping; the article body no longer goes through
  `convert.WrapText`. `convert.WrapText` stays in the package (still exported
  and tested) but is unused by the article path.
- `Copy Article Text` copies the visible rendered text (ANSI/OSC 8 stripped),
  not markdown source.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Article content rendering" requirement is rewritten so
  markdown is rendered to styled terminal text by glamour (theme-configurable
  dark/light/auto), with links in glamour's `text url` form and overlong URLs
  wrapping across lines instead of being ellipsis-truncated. The "Clickable
  links in article content" requirement is rewritten to reflect that OSC 8
  wrapping is glamour's responsibility and the visible link form is
  `text url`. The "Clickable article URL in reader header" requirement is
  amended so the header link is a markdown link rendered by glamour, still OSC
  8 clickable.

## Impact

- **Dependencies**: Adds `charm.land/glamour/v2` (v2.0.1; repository
  `github.com/charmbracelet/glamour`). Pulls in glamour's transitive deps
  (charmbracelet/x/wrap, goldmark, etc.).
- **Converter**: `internal/convert` — `Convert` emits plain commonmark; the
  custom `<strong>`/`<b>`/`<a>` ANSI renderers are removed; the `<img>`
  `[alt]`/`[image]` renderer is kept. `WrapText` is retained but unused by the
  article path.
- **UI layer**: `internal/ui` — new `markdown.go` renderer seam (cached
  `glamour.TermRenderer` per width/theme); `article.go` renders the header +
  body markdown through glamour and drops `WrapText`; `Copy Article Text` uses
  the rendered, stripped text.
- **Tests**: `convert/convert_test.go` asserts plain markdown; UI tests
  recalibrate for glamour output (margins, link form, conditional OSC 8).
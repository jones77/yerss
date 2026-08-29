## Why

Article HTML is converted to plain text by `jaytaylor/html2text`, a library
with no extension seam: it fixes how every tag renders and offers no per-tag
override. Six regex pre-passes in `internal/ui/html.go` work around its
limitations — flattening `<li><p>`, emboldening `<strong>` via ANSI escapes,
stripping `<span>` to kill spurious spaces, hand-rolling list indentation with
a space-counting trick that reasons about html2text's internal `\n `→`\n` trim,
and pre-wrapping `<a>` tags in OSC 8 because html2text loses the href→text
association. The result is fragile (each hack is coupled to html2text's exact
output behavior), incomplete (blockquotes, code blocks, and tables are not
handled at all), and the conversion, width-wrapping, and display-width
primitive layers are tangled together in one 327-line file.

`JohannesKaufmann/html-to-markdown` is built around the exact seam html2text
lacks — `Register.RendererFor(tag, type, fn, priority)` lets each tag's output
be controlled directly — so the hack stack dissolves into three small custom
renderers while native commonmark handling gives correct lists, nesting,
blockquotes, code blocks, and tables for free.

## What Changes

- Replace `jaytaylor/html2text` with `JohannesKaufmann/html-to-markdown/v2`
  (base + commonmark plugins) as the HTML conversion engine.
- Register three custom renderers that emit terminal escapes directly into the
  markdown output: `<a>` → OSC 8 hyperlink-wrapped `[text](url)`, `<strong>` /
  `<b>` → ANSI bold (`\x1b[1m…\x1b[22m`), `<img>` → `[alt]` / `[image]`
  placeholder. Native commonmark handles lists, blockquotes, code blocks, and
  tables.
- Set `EscapeModeDisabled` so literal `*`, `[`, `]` in article prose pass
  through raw (the markdown is never re-parsed; smart escaping would only add
  backslash noise).
- Remove the six regex pre-passes (`replaceImages`, `flattenListParagraphs`,
  `emboldenStrong`, `stripSpans`, `renderLists`, `wrapLinks`) and their
  compiled regexps — all superseded by the converter and its renderers.
- Split the conversion and terminal-formatting layers out of the `internal/ui`
  god-package into a new repo-root `convert` package: `convert/convert.go`
  (HTML→markdown with embedded terminal escapes) and `convert/wrap.go`
  (width-aware soft wrapping, moved verbatim from `html.go`).
  `internal/ui` retains the bubbletea model/view code and the generic
  display-width primitives (`truncate`, `padRight`) used by borders and lists.
- Preserve the existing `wrapText` link-aware wrapping logic unchanged — it
  already keys on the OSC 8 open sequence (`\x1b]8;`), not bare `](`, so it is
  stable across the converter swap.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Article content rendering" requirement is modified so
  `<strong>` / `<b>` render as ANSI bold (not literal markdown asterisks), and
  structured HTML (lists, blockquotes, code blocks, tables) renders as
  readable markdown syntax with correct nesting and indentation. The
  image-placeholder, link-as-markdown, URL-break, and URL-ellipsis behaviors
  are preserved.

## Impact

- **Dependencies**: Removes `github.com/jaytaylor/html2text`; adds
  `github.com/JohannesKaufmann/html-to-markdown/v2` (MIT, 3.8k stars, golden-
  file tested). The converter pulls in `JohannesKaufmann/dom` and uses
  `golang.org/x/net/html` (already an indirect dependency via gofeed).
- **New package**: `convert/` at repository root — `convert.go` (converter
  construction + three custom renderers) and `wrap.go` (`wrapText`,
  `linkTokens`, `truncateURL` moved from `internal/ui/html.go`).
- **UI layer**: `internal/ui/html.go` is deleted; its conversion logic moves
  to `convert/convert.go`. `internal/ui/article.go` imports `yerss/convert`
  instead of calling local `HTMLToText` / `wrapText`. Generic display-width
  primitives (`truncate`, `padRight`) stay in `internal/ui` (used by
  `border.go`, `list.go`).
- **Tests**: `internal/ui/html_test.go` and `osc8_test.go` assertions move to
  `convert/` and are updated where the converter changes output shape (e.g.
  list nesting now native; bold now via the custom renderer). `polish_test.go`
  and `render_safety_test.go` (which exercise `wrapText` / `truncate`) update
  import paths.
- **Spec**: `openspec/specs/reader-ui/spec.md` "Article content rendering"
  requirement gains bold-rendering and structured-content scenarios.
- **Sequencing**: Orthogonal to the in-progress `add-article-images` change,
  which splices an image block into `renderArticleContent` and is coupled to
  the viewport, not the converter. Doing this swap first gives
  `add-article-images` a cleaner base and prevents merge churn in
  `html.go` / `article.go`.

## Why

Two error paths in the article rendering pipeline silently feed the *wrong*
input to the next stage, and one placeholder path fails to escape its text.
`convert.Convert` returns the raw HTML on conversion error, which glamour then
parses as markdown and renders `<p>`/`<a>` tags as literal text; `renderMarkdown`
returns the raw markdown on render error, so `**bold**` and `[text](url)` appear
verbatim; and `renderImage` writes `[alt]` without escaping, so alt text
containing `*`, `_`, or `[` becomes live markdown instead of literal text. Each
is a rare path, but each produces garbled reader output rather than degrading
cleanly.

## What Changes

- **BREAKING**: On HTML-to-markdown conversion failure, `convert.Convert`
  SHALL NOT return the raw HTML. It SHALL degrade to the article's plain text
  (HTML markup removed), so the reader shows readable text without tags.
- **BREAKING**: On markdown-render failure, `renderMarkdown` SHALL NOT return
  the raw markdown. It SHALL degrade to plain text with markdown formatting
  markers neutralized, so no literal `**` or `[text](url)` syntax is displayed.
- Image alt text SHALL be escaped before it is written as the `[alt]`
  placeholder, so alt text containing markdown special characters renders
  literally instead of being styled or broken.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Article content rendering" requirement is amended so image
  alt text renders literally (escaped) and so conversion/render failure degrades
  to plain text rather than displaying raw HTML or markdown syntax.

## Impact

- **Converter**: `convert/convert.go` — `Convert` gains a plain-text fallback
  path (reusing `golang.org/x/net/html`, already a dependency).
- **UI layer**: `internal/ui/markdown.go` (`renderMarkdown`, `renderImage`).
- **Tests**: `convert/convert_test.go` and `internal/ui/markdown_test.go` gain
  failure-path and alt-escaping cases; existing rendering tests must stay green.

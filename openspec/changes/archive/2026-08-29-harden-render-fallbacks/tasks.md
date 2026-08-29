## 1. Convert plain-text fallback

- [x] 1.1 Add a `plainText(html string) string` helper to `convert` that walks the `x/net/html` tree and returns text-node content
- [x] 1.2 Replace the `return html` error path in `Convert` with `return plainText(html)`
- [x] 1.3 Add a `convert` test asserting a failing conversion returns plain text with no `<p>`/`<a>` tag text

## 2. Markdown render plain-text fallback

- [x] 2.1 Add a `markdownToPlainText(md string) string` helper to `internal/ui` that strips block/inline markdown markers (headings, lists, blockquotes, emphasis, code, links, images)
- [x] 2.2 Replace the `return md` error path in `renderMarkdown` with `return markdownToPlainText(md)`
- [x] 2.3 Add a test asserting the fallback contains no literal `**` or `[text](url)` syntax and keeps the visible text and URL

## 3. Image alt-text escaping

- [x] 3.1 Add an `escapeAlt` helper to `convert` and apply it to the alt text in `renderImage` before writing `[alt]`
- [x] 3.2 Add a `convert` test asserting alt text with `*`/`_`/`[`/`]` renders as literal `[alt]` with no styled/link interpretation

## 4. Verify

- [x] 4.1 Run `go build ./...` and `go vet ./...` and confirm a clean result
- [x] 4.2 Run `go test ./convert/... ./internal/ui/...` and confirm all tests pass

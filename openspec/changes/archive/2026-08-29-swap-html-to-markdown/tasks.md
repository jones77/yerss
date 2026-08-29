## 1. Spike — de-risk OSC 8 / ANSI escape passthrough

- [x] 1.1 Add `github.com/JohannesKaufmann/html-to-markdown/v2` to go.mod (`go get`)
- [x] 1.2 Write a throwaway spike test (can live in `convert/` or a temp file) that constructs a converter with base + commonmark plugins, `EscapeModeDisabled`, and custom renderers for `<a>`, `<strong>`/`<b>`, and `<img>`, then feeds `<p>See <a href="https://example.com/x"><strong>bold link</strong></a> and <img src="c.png" alt="a cat"></p>` and asserts the output contains: OSC-8-open + `\x1b[1mbold link\x1b[22m` + OSC-8-close + `(https://example.com/x)` with no stray backslashes or mangled escapes, plus `[a cat]` for the image
- [x] 1.3 Confirm the renderer child-walk API (e.g. `ctx.RenderChildren(w, n)` or manual `n.FirstChild` walk) composes bold-inside-link correctly; record the exact API used
- [x] 1.4 Feed a nested list (`<ul><li>top<ul><li>nested</li></ul></li></ul>`) and a blockquote through the spike converter and eyeball that nesting/indentation and `> ` prefixes render correctly
- [x] 1.5 If the spike passes, delete the throwaway test and proceed. If it fails, stop and update design.md with the post-render injection pivot before continuing.

## 2. Create `convert` package

- [x] 2.1 Create `convert/convert.go` with a `Convert(html string) string` function that constructs the html-to-markdown converter (base + commonmark plugins, `EscapeModeDisabled`, commonmark bullet marker `*`) and runs `ConvertString`
- [x] 2.2 Register the `<img>` renderer (`RendererFor("img", Inline, ...)`): extract `alt` attribute, write `[alt]` or `[image]` to `w`
- [x] 2.3 Register the `<strong>` / `<b>` renderer (`RendererFor` for both tags, `Inline`): write `\x1b[1m`, render children, write `\x1b[22m`
- [x] 2.4 Register the `<a>` renderer (`RendererFor("a", Inline, ...)`): extract `href`, write OSC-8-open + `[` + rendered children + `]` + OSC-8-close + `(` + url + `)`
- [x] 2.5 Verify `go build ./convert/...` passes

## 3. Move wrapping logic to `convert/wrap.go`

- [x] 3.1 Move `wrapText`, `linkTokens`, `truncateURL`, `isTrailingPunct`, and the `linkOpenSeq` constant from `internal/ui/html.go` to `convert/wrap.go` verbatim (no logic changes)
- [x] 3.2 Export `WrapText` (and any helpers it needs) so `internal/ui` can call it
- [x] 3.3 Verify `go build ./convert/...` passes

## 4. Move display-width primitives within `internal/ui`

- [x] 4.1 Move `truncate` and `padRight` from `internal/ui/html.go` to `internal/ui/width.go` (keep unexported; only used within `internal/ui`)
- [x] 4.2 Verify `go build ./internal/ui/...` passes (border.go, list.go still compile)

## 5. Wire `internal/ui` to the `convert` package

- [x] 5.1 Update `internal/ui/article.go`: import `yerss/convert`, replace `HTMLToText(a.Content)` with `convert.Convert(a.Content)` and `wrapText(...)` with `convert.WrapText(...)`
- [x] 5.2 Update `internal/ui/article.go` `CopyArticleText` path (uses `renderArticleContent`) to use `convert.Convert` if it references the old function
- [x] 5.3 Delete `internal/ui/html.go`
- [x] 5.4 Remove `github.com/jaytaylor/html2text` from go.mod (`go mod tidy`)
- [x] 5.5 Verify `go build ./...` passes

## 6. Update tests

- [x] 6.1 Move the conversion assertions from `internal/ui/html_test.go` and `osc8_test.go` into `convert/convert_test.go`, updating expected outputs where the converter's native output differs from the old html2text+hacks output (list nesting, bullet spacing, bold via renderer)
- [x] 6.2 Move `wrapText`-focused tests (`TestWrapText`, `TestWrapTextKeepsBoldEscapes`) into `convert/wrap_test.go`, updating the call to `WrapText`
- [x] 6.3 Update `internal/ui/polish_test.go` and `internal/ui/render_safety_test.go` import paths / call sites for `wrapText` → `convert.WrapText` and `truncate` (now local to `internal/ui`)
- [x] 6.4 Run `go test ./...` and fix any remaining failures
- [x] 6.5 Run `golangci-lt run` (or the project's linter) and fix any issues

## 7. Visual verification

- [x] 7.1 Run yerss against real feeds (including a Substack feed for `<li><p>` lists) and confirm bullets render as bullets, bold renders bold, links are clickable with URL fallback, blockquotes show `> ` prefixes, and code blocks are readable
- [x] 7.2 Test ASCII fallback mode (`-a` / `--ascii`) — confirm OSC 8 links still work and bold renders (bold is ANSI, not Unicode) and no Unicode-only glyphs leak
- [x] 7.3 Test a narrow terminal width to confirm URL-break and ellipsis truncation still work via the unchanged `WrapText`

## 8. Bug fix — WrapText preserves nested-list indentation

- [x] 8.1 Fix `WrapText` to preserve leading indentation across wrapped lines: nested list items rendered at column 0 because `strings.Fields` (via `linkTokens`) collapsed leading spaces. Each paragraph now keeps its leading indent, the effective width is reduced by the indent width, and every wrapped line is re-prefixed with it
- [x] 8.2 Add regression tests (`TestWrapTextPreservesNestedListIndentation`, `TestWrapTextPreservesIndentOnWrappedLines`) asserting nested bullets and their wrapped continuation lines stay indented through the full `Convert` + `WrapText` pipeline
- [x] 8.3 Verify against the real Drop Site article: nested `<li><p>…</p><ul><li>…</li></ul></li>` bullets render indented (2 spaces per level, 3 levels confirmed)

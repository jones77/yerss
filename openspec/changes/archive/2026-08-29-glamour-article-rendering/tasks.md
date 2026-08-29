## 1. Dependency

- [x] 1.1 Add `charm.land/glamour/v2` (v2.0.1) to `go.mod` via `go get`; tidy and confirm the build resolves

## 2. Converter emits plain markdown

- [x] 2.1 Remove the `<strong>`/`<b>` and `<a>` ANSI renderers (`renderBold`, `renderLink`) and their registrations in `convert/convert.go` so commonmark emits `**bold**` and `[text](url)`; keep the `<img>` `[alt]`/`[image]` renderer and `collapseBlankLines`
- [x] 2.2 Update `convert/convert_test.go`: bold/link tests assert plain markdown instead of ANSI escapes (rename `TestConvertWrapsLinksInOSC8` to reflect markdown links); keep list/paragraph/image assertions

## 3. Glamour renderer seam

- [x] 3.1 Add `internal/ui/markdown.go`: `markdownRenderer(contentW int)` builds/caches a `*glamour.TermRenderer` on `Model` (fields for renderer, width, style); rebuild when width or resolved style changes; map `dark`/`light`/`auto` via `detectLightBackground()`; `WithStandardStyle` + `WithWordWrap(contentW)`
- [x] 3.2 Test: theme mapping (dark→dark, light→light, auto→light/dark by detection), renderer rebuilt on width change, plain markdown rendered to styled output

## 4. Article rendering through glamour

- [x] 4.1 Replace `renderArticleContent` with `renderArticleMarkdown`: `**Title**` / `*by Author*` / `[url](url)` header plus `convert.Convert(a.Content)` body
- [x] 4.2 In `newArticleState`, render the markdown via `markdownRenderer(contentW)`, trim the trailing newline, `viewport.SetContent(rendered)`; drop `convert.WrapText` for the body; fall back to raw markdown on render error
- [x] 4.3 Update `CopyArticleText` to strip ANSI/OSC 8 from the rendered article text (not the markdown source)

## 5. Test calibration

- [x] 5.1 Update `internal/ui/osc8_test.go` and `internal/ui/mouse_selection_test.go` for the new link form and content function
- [x] 5.2 Recalibrate `internal/ui/render_test.go`, `internal/ui/polish_test.go`, `internal/ui/render_safety_test.go` for glamour output (margins, styling, line counts)

## 6. Validation

- [x] 6.1 Run `go test ./...` and `golangci-lint`; fix any failures
- [x] 6.2 `openspec validate --change glamour-article-rendering --strict` passes
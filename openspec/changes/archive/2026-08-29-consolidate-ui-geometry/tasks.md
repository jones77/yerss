## 1. Shared geometry

- [x] 1.1 Add `contentGeom(w, h, padX, padY int) (textW, viewportH, effPadY int)` (and the `wrapIndex`/`dayKey` pure helpers) in a new `internal/ui/geometry.go`
- [x] 1.2 Route `newArticleState` (article.go) through `contentGeom`, replacing its inline `contentW`/`vpH` math
- [x] 1.3 Route `contentRect` (article.go) through `contentGeom`, keeping its `x0/y0` origin offsets
- [x] 1.4 Route `renderArticleBorder` (border.go) through `contentGeom` for `textW`/`viewportH`/`effPadY`
- [x] 1.5 Test: `contentGeom` returns the clamped values; `newArticleState` and `contentRect` agree for `height < 2*padY+3`

## 2. Cursor / key helpers

- [x] 2.1 Add `wrapIndex(i, delta, n int) int` and use it in `moveListCursor` (list.go)
- [x] 2.2 Use `wrapIndex` in `movePopupCursor` (popup.go)
- [x] 2.3 Add `dayKey(t time.Time) string` and use it in `bucketDayGroups` (list.go), `persistSelection` (selection.go), `restoreSelection` (selection.go)
- [x] 2.4 Add `articleAtCursor() (*articleItem, bool)` and use it in `openArticle`, `toggleReadAtCursor` (article.go/list.go)
- [x] 2.5 Use `articleAtCursor` where applicable in `handleListClick` (list.go); leave header-row branches using `visibleRows()` as-is
- [x] 2.6 Test: `wrapIndex` wraps and fixes negative modulo; `dayKey` zero-time and boundary cases

## 3. Width helpers and clamp sweep

- [x] 3.1 Merge `padToWidth` into `padRight` (width.go); delete `padToWidth`; update its caller in `spliceStyled` (model.go)
- [x] 3.2 Replace `if x < 1 { x = 1 }` with `max(1, x)` in list.go, article.go, border.go, image.go where idiomatic
- [x] 3.3 Reuse `pageSize()` in `renderList` and `handleListClick` instead of inline `m.height - 1` clamps

## 4. Validation

- [x] 4.1 Run `go test ./...` and `go vet ./...`; fix any failures
- [x] 4.2 `openspec validate --changes consolidate-ui-geometry --strict` passes
- [x] 4.3 Manual smoke: list and article views render identically at a normal size and at a tiny (2-line) terminal height (covered by `render_safety_test.go` degenerate-size tests; worth an eyeball in a real terminal)

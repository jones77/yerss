# Tasks

## 1. Restore the original geometry and rendering

- [x] 1.1 Restore `ContentGeom` in `internal/ui/render/geometry.go` to `(w, h, padX, padY) -> (textW, viewportH, effPadY)` with `viewportH = h - 2*effPadY - 2` (symmetric top and bottom margin).
- [x] 1.2 Restore `ClampPadY`, the `padY` parameter on `RenderArticleBorder`, and the below-content blank-row loop in `internal/ui/render/border.go`; render `effPadY` blank rows below the top border as the top margin; update doc comments.
- [x] 1.3 Restore the `ContentGeom`/`RenderArticleBorder` callers in `internal/ui/article.go` (removing the trailing-padding append), `internal/ui/image.go`, and `internal/ui/article_selection.go` (content origin at `row 1 + effPadY`).

## 2. Tests

- [x] 2.1 Restore `internal/ui/geometry_test.go`, `internal/ui/render_safety_test.go`, `internal/ui/render_test.go`, `internal/ui/mouse_selection_test.go`, and `internal/ui/article_image_test.go` to the pre-article-padding-snap-flush state (removing the trailing-padding tests), then update them for the symmetric margin (smaller viewport, content origin shifted down).

## 3. Verification

- [x] 3.1 Run `go test ./...` and `go vet ./...`; confirm the full suite passes.
- [x] 3.2 Reproduce the real ProPublica article at 80x24 and confirm the top margin below the top border, the bottom-snapped caption's last line as the viewport's last line, and the bottom margin above the bottom border.
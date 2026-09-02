# Tasks

## 1. Geometry and rendering

- [x] 1.1 Change `ContentGeom` in `internal/ui/render/geometry.go` to drop the `padY`/`effPadY` outputs: `viewportH = h - 2`.
- [x] 1.2 Remove `ClampPadY` and the below-content blank-row loop from `internal/ui/render/border.go`, and drop the `padY` parameter from `RenderArticleBorder`; update doc comments.
- [x] 1.3 Update the `ContentGeom`/`RenderArticleBorder` callers in `internal/ui/article.go`, `internal/ui/image.go`, and `internal/ui/article_selection.go`.

## 2. Padding as trailing content

- [x] 2.1 In `internal/ui/article.go` `composeArticle`, append `padding_y` blank rows to the end of the article's scrollable content (after the body, so image-block positions are unchanged).
- [x] 2.2 Confirm config keeps both `PaddingX` (default 2) and `PaddingY` (default 1) as independent, overridable options (no config changes needed beyond the geometry cleanup).

## 3. Tests

- [x] 3.1 Update `internal/ui/geometry_test.go` (`TestContentGeom` cases and `TestContentGeomStateAndRectAgree`).
- [x] 3.2 Update `internal/ui/render_safety_test.go` (`TestRenderArticleBorderHeightClamp` signature).
- [x] 3.3 Recompute the scrollbar thumb tracks and `N/M` labels in `internal/ui/render_test.go` for the larger viewport.
- [x] 3.4 Add a test that a bottom-snapped inline image caption's last line is the last visible row of the rendered article frame, modeled on the ProPublica staircase figure.
- [x] 3.5 Add a test that the article's scrollable content ends with `padding_y` blank rows (GotoBottom shows the final text line above the trailing padding).

## 4. Verification

- [x] 4.1 Run `go test ./...` and `go vet ./...`; confirm the full suite passes.
- [x] 4.2 Reproduce the real ProPublica article at 80x24 and confirm the bottom-snapped caption's last line is the viewport's last visible row (flush at the bottom border).
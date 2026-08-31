## 1. Caption width

- [x] 1.1 Compute a standard caption width once per geometry (a fixed fraction of the content width, e.g. 70%) and make `fitBlock`/`composeBlock` in `internal/ui/image.go` wrap the attribution at it instead of the image's `imgW`, centering the wrapped lines beneath the block
- [x] 1.2 Make `fitBlock`'s iteration reserve the wrapped attribution line count computed from the caption width, so the image shrinks to fit the image plus the wrapped caption within the viewport budget
- [x] 1.3 UI tests: a long credit wraps at the caption width (not the photo width) and is centered; a long caption scales the image down so the block fits the viewport

## 2. Fit uses the composed caption

- [x] 2.1 Track the composed caption (including the inline link text) per inline image and pass it to `nativeRenderCmd`/`inlineAttrFor` so the native render's height fit reserves the real wrapped caption line count instead of the shortened per-URL fallback
- [x] 2.2 Verify a native-terminal inline image whose long caption previously filled the viewport now fits with a distinct bottom stage (no full-viewport collapse)
- [x] 2.3 Tests: the native fit and the composed block agree on the caption, so a long inline caption cannot overflow the block past the viewport

## 3. Regression on inline-image-rendering

- [x] 3.1 Re-run the `inline-image-rendering` snap tests with the shorter caption geometry and update any expected offsets that assumed the overflowed (full-viewport) block heights
- [x] 3.2 Manual smoke test on the hurricane article: the two gallery photos' identical long caption wraps compactly at the standard width, each block fits with room to spare, and scroll-j/scroll-k keep distinct two-stage snapping for both photos

## 4. Verification

- [x] 4.1 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [x] 4.2 Run `openspec validate standard-caption-width`
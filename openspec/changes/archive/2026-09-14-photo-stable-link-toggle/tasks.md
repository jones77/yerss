# Tasks: photo-stable-link-toggle

## 1. Tests

- [x] 1.1 Model test: toggling link mode keeps a visible photo on screen and re-positions it at the same relative photo row (offset inside a photo block's rows; assert the same URL's block spans the viewport top after `x`, and the photo did not leave the viewport)
- [x] 1.2 Model test: toggling link mode with the attribution row first visible keeps the photo on screen (offset inside `[CapStart, ImgEnd]`; same URL's block stays in view)
- [x] 1.3 Model test: photo does not change size when toggling modes (re-fit path: block not width-matched; assert identical photo row counts in compact and expanded)
- [x] 1.4 Unit test: `fitHeaderLines` equals the max of compact and expanded header line counts and is used by the fit when the two modes' headers differ

## 2. Implementation

- [x] 2.1 Add `fitHeaderLines` to `articleDerivation` and `articleState`; compute it in `computeDerivation` as `max(compactHeaderLines, expandedHeaderLines)` by rendering the header in both link modes
- [x] 2.2 Switch the height fit to `fitHeaderLines`: `composeImageBlock`/`fitBlock` call path (article.go) and `nativeRenderCmd` (image.go) pass the mode-independent budget; keep `headerLines` for actual-header position logic
- [x] 2.3 Add photo-block anchoring to `toggleLinkExpand` (image.go): when the first visible line lies inside an image block span `[ImgStart, ImgEnd]`, flip mode, recompose, then re-anchor at the same URL's block with the same relative row (clamped), falling back to the text anchor when the URL is gone; keep `atBottom`/`GotoBottom` short-circuit

## 3. Verification

- [x] 3.1 `go test ./internal/ui/...`
- [x] 3.2 `go vet ./...`
- [x] 3.3 `openspec validate --all --strict`
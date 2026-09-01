## 1. Boundary primitives

- [x] 1.1 Define the four canonical snap positions for an image block in `internal/ui/image.go` (computed per snap call from the current `vpH`): `BOTTOM(b) = b.imgEnd - vpH + 1`, `TOP(b) = b.imgStart`, `EXIT(b) = b.capStart`, `OFF(b) = b.imgStart - vpH`
- [x] 1.2 Update the `imageBlock` doc comment in `internal/ui/article.go` to name the two snap boundaries (top = image's first line at viewport top; bottom = wrapped caption's last line at viewport bottom) and the four positions

## 2. Snap rule restructure

- [x] 2.1 Key the bottom-snap and scroll-off conditions on the block's last line: replace `b.capStart > yOffset+vpH` with `b.imgEnd > yOffset+vpH` in the down bottom-snap and up scroll-off branches of `snapYOffset`
- [x] 2.2 Restructure the down branch of `snapYOffset` around the four positions, preserving: the lead block's pre-consumed stages (photo-range landing and fully-visible skip go to `EXIT`), the `before == BOTTOM(b)` rise-to-top stage, the reveal-to-top on entering the photo range from above, and the `best` (lowest) selection for multiple bottom-clipped candidates
- [x] 2.3 Restructure the up branch of `snapYOffset` around the four positions, preserving: the photo-range reveal to `TOP`, the `before == TOP(b)` sink-to-bottom stage, the entry-from-above top snap, the fills-viewport collapse (single position: next move scrolls off rather than looping), and the reveal-previous fallback when an `OFF` target lands in a preceding block's photo range
- [x] 2.4 Keep the `snap` loop guard ("never return the pre-move offset") intact across both branches

## 3. Tests and verification

- [x] 3.1 Add a test for the wrapped-caption bottom snap: an inline image whose caption wraps to three lines with only its first line visible snaps to the bottom boundary so the caption's last line sits at the viewport bottom (new scenario in `article_image_test.go`)
- [x] 3.2 Confirm the existing snap tests pass unmodified (two-stage bottom-then-top, full-viewport no-loop, never-leaves-inline-image-blank down/up, skip-leaves-next-partial, lead skip scenarios)
- [x] 3.3 Run the full ui test suite, the linter, and a build; verify `git status` shows only the intended files changed
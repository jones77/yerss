## 1. Gallery captions become centered captions

- [x] 1.1 Fix `figureCaption` (internal/convert/credit.go) to attribute the figcaption of the figure that directly contains the image (the innermost figure), so each photo in a gallery with per-photo figcaptions resolves its own caption; preserve the photo-grid rule (a single figcaption over multiple images still belongs to the last image).
- [x] 1.2 Update `ResolveImageAttribution` (internal/ui/compose/image.go) precedence: directly-containing figure's caption first, then alt text, then link text, then `photo: <source>`.
- [x] 1.3 In `ConvertImages` (internal/convert/convert.go), suppress a single-image figure's `figcaption` from the body markdown (emit only the image sentinel), keeping a multi-image figure's shared figcaption (e.g. a gallery credit) as body text, so a per-photo caption is not duplicated.
- [x] 1.4 Add regression tests: a gallery fixture of three nested figures with per-photo figcaptions composes each figcaption centered beneath its photo and exactly once (no left-aligned body-text copy); existing photo-grid and linked-promo attribution scenarios still pass.

## 2. Downward scroll snaps short images

- [x] 2.1 Extend `SnapYOffset`'s downward branch (internal/ui/compose/image.go): a single-line down move that brings a short inline image's top into the window from below snaps to its top boundary (flush to the viewport top); the next down move skips it onto its caption. Keep the partial-visibility bottom snap and the `before`-gating (a snap never returns the pre-move offset).
- [x] 2.2 Add scroll tests: with the Helene gallery geometry (short images entered from above), assert the down scroll snaps flush to the top boundary then skips to the caption, up scroll still snaps, and no snap loops.
- [x] 2.3 Run the full test suite (`go test ./...`) and the linter; confirm no regressions in scroll/snap, native-clear, or attribution tests.

## 3. Verification

- [x] 3.1 Manual check on the Helene article: the three gallery photos each show a centered caption with no duplicated left-aligned body text, and scroll-j from the "Caring for others" paragraph snaps the photos flush instead of scrolling line-by-line. (Verified against the stored article via the DB-reproduction path: real block geometry rendered with the real height fit and scrolls simulated; 80×24 and 100×40 both confirm each caption renders exactly once and the down scroll snaps flush.)

## 4. Follow-up: the caption is part of the image snap unit

Feedback on the reader at 124×37: the second gallery photo was skipped entirely when scrolling up (the off-reveal chose a higher image), a multiline caption's last line showed alone before the image when scrolling up, and a down scroll from a top-aligned image landed on the caption's first line instead of past the block.

- [x] 4.1 Update the `reader-ui` delta spec: the down-skip lands on the line immediately after the wrapped caption (the image and caption are one snap unit); the upward entry snap fires when the block's last line enters the window from above even at its first row, so the whole block appears at once; the up scroll-off reveals the immediately preceding image so consecutive images show in sequence.
- [x] 4.2 Change the inline down-exit target from the caption's first line to the line after the caption (`snapPositions` uses `ImgEnd+1` for inline blocks; the lead keeps landing on its caption); a down scroll from a top-aligned image now skips the whole block.
- [x] 4.3 Make the upward entry snap fire when the block's last line is at the window's first row (`b.ImgEnd >= yOffset`), so an up scroll never rests on a caption-only frame; a multiline caption enters with its image.
- [x] 4.4 Make the upward scroll-off reveal the nearest (immediately preceding) image (`offRevealTop` scans downward), so the second gallery photo is no longer skipped on the way up.
- [x] 4.5 Update the affected scroll/snap tests and add `TestSnapMultilineCaptionUpEntryShowsWholeBlock`; run the full suite and linter.
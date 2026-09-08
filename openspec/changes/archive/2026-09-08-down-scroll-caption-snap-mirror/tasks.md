# Tasks

## 1. Implementation

- [x] 1.1 Change `SnapYOffset`'s down-exit target in `internal/ui/compose/image.go` to the first line of the next image or paragraph (one line past the block's blank separator) so the space below an image scrolls off with it.
- [x] 1.2 Change the fully visible lead photo's down behavior to rise flush to the viewport top first, with the following press skipping the whole block past its caption.
- [x] 1.3 For an adjacent short image, make the down-skip land on the next image's BOTTOM boundary (its caption at the viewport bottom), mirroring the upward sink; run the rise stage before the photo-range skip so the adjacent bottom stage rises instead of looping.
- [x] 1.4 Add the up-direction settle-to-bottom snap for a short inline block left fully visible mid-window by a non-snapping move.
- [x] 1.5 Update the `SnapYOffset` doc comment and the inline comments to describe the mirror and the adjacent-image bottom reveal.

## 2. Tests

- [x] 2.1 Update `TestSnapYOffset` and `TestSnapYOffsetBlockList` to the new exit (first line past the block).
- [x] 2.2 Update `TestScrollSnapsOntoCaption`, `TestScrollSkipsFullyVisiblePhoto`, and `TestLeadPhotoRisesToTopThenSkipsOntoCaption` to the lead rise-then-skip flow.
- [x] 2.3 Update the adjacent-photo tests (`TestScrollSnapsAcrossAdjacentPhotosSkipsGap`, `TestSnapYOffsetAdjacentDoublePhotoNotSkipped`, `TestSnapYOffsetCaptionlessGridNextImageFlush`, `TestSnapHeleneGalleryShortImagesDownFlush`) to the next-photo bottom-boundary reveal and the rise that follows.
- [x] 2.4 Update `TestSnapHeleneStaircaseFigureBottomThenTop`, `TestSnapMultilineCaptionUpEntryShowsWholeBlock`, `TestSnapYOffsetSkipLeavesNextImagePartial`, `TestSnapYOffsetFullViewportBlockDoesNotLoop`, `TestScrollSnapsShortInlineImageBottomThenTop`, and `TestScrollSnapsThroughInlineImage` to the new exit offsets.
- [x] 2.5 Add `TestSnapUpFromArticleEndSettlesPhotoAtViewportBottom` for the up-direction settle.

## 3. Verification

- [x] 3.1 Run `go test ./...` and `go vet ./internal/ui/...`; confirm the full suite passes.
- [x] 3.2 Verify no snap loop: the down-then-up walks in the adjacent-photo and gallery tests advance on every press in both directions.
- [x] 3.3 Reproduce the reported ProPublica Helene article geometry from the local SQLite DB and confirm the down sequence through the three Sandra Rogers photos mirrors the up sequence (each photo shows a caption-at-bottom stage in both directions).
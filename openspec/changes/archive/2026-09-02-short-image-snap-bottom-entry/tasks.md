# Tasks

## 1. Implementation

- [x] 1.1 Change `SnapYOffset`'s short-image down-entry snap in `internal/ui/compose/image.go` to target the block's BOTTOM boundary instead of its TOP, and update its doc comments (the function doc and the inline comment) to describe the two-stage entry.

## 2. Tests

- [x] 2.1 Update `TestSnapYOffsetBlockList`'s short-entry case to expect the bottom boundary.
- [x] 2.2 Update `TestSnapHeleneGalleryShortImagesDownFlush`'s down table to the two-stage entry (bottom, then top, then skip for each photo) and confirm the up table and no-loop walks still pass.
- [x] 2.3 Update `TestScrollSnapsShortInlineImageFlushThenCaption` to assert the bottom → top → skip flow.
- [x] 2.4 Add a test modeling the reported ProPublica staircase-figure geometry (standalone short inline block, long wrapped caption) asserting the entry snaps to the bottom boundary, the next down move rises to the top boundary, and the following move skips past the caption.

## 3. Verification

- [x] 3.1 Run `go test ./internal/ui/...` and `go vet ./internal/ui/...`; confirm the full suite passes.
- [x] 3.2 Verify no snap loop: a down-then-up walk over a short inline block advances on every press in both directions.
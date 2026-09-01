## Why

The two-stage image snap is "a lot better but still buggy": `snapYOffset` is a
pile of overlapping special cases that chase two boundary positions only
implicitly, so interactions between them are hard to reason about and keep
producing edge-case bugs. The clean model is simple: every image block has
exactly two snap boundaries — the image's first line at the viewport top, and
the last line of the wrapped caption at the viewport bottom. Making these
boundaries first-class removes the confusion about how to snap.

## What Changes

- Define two snap boundaries per image block, used everywhere snapping happens:
  - `TOP(b) = imgStart` — the image's first line at the viewport's first row.
  - `BOTTOM(b) = imgEnd - vpH + 1` — the **last line of the wrapped caption** at
    the viewport's last row.
- The bottom-snap/scroll-off conditions key on `imgEnd` (the wrapped caption's
  last line) instead of `capStart` (the caption's first line), so a wrapped
  caption whose tail hangs below the fold snaps correctly instead of staying
  partially cut.
- Restructure `snapYOffset` around the four canonical positions per block so the
  stage transitions fall out of the boundaries rather than being bolted on:
  - down: enter at `BOTTOM(b)` → rise to `TOP(b)` → exit onto the caption
    (`capStart`)
  - up: enter at `TOP(b)` → sink to `BOTTOM(b)` → scroll off below the fold
    (`imgStart - vpH`)
- Keep the genuinely special cases as named exceptions of the boundary model,
  not ad-hoc patches: the lead block (its stages are pre-consumed — it was shown
  on open, so a down press skips straight to its caption), the fills-viewport
  collapse (`TOP(b) == BOTTOM(b)`, one position, so rise/sink collapse and the
  next press exits or scrolls off), the native-terminal lead top-skip
  optimization, and adjacent-block collisions (an exit/scroll-off target landing
  in a neighbor's photo range reveals that neighbor instead).
- The down-exit target stays `capStart` — after the image rises to the top, the
  next down press lands on the caption, keeping it at the top of the screen.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `reader-ui`: the "Atomic image scroll behavior" requirement changes so that
  snapping is defined by two explicit per-block boundaries (image top at
  viewport top; last wrapped-caption line at viewport bottom), the
  bottom-snap/scroll-off conditions key on the block's last line, and the
  two-stage transitions are expressed as boundary-to-boundary moves.

## Impact

- `internal/ui/image.go`: `snapYOffset`, `scrollArticle`, and the boundary
  conditions in the snap rules; `nativeImageClear` /
  `suppressClippedNativeTransmits` / `previewClippedImage` keep their
  fully-contained checks keyed on the photo range (unchanged, but re-verified
  against the new boundary definitions).
- `internal/ui/article.go`: `imageBlock` geometry documentation (`imgStart`,
  `capStart`, `imgEnd`) — no field changes expected.
- `internal/ui/article_image_test.go`: snap tests updated/extended for the
  boundary model (wrapped-caption bottom snap, boundary-based stage
  transitions).
- `openspec/specs/reader-ui/spec.md`: the atomic scroll behavior requirement and
  its scenarios updated.
## Why

Photo captions wrap to the photo's own width, so a long caption on a narrow
photo balloons into many lines that overflow the reserved fit budget — the
intercept hurricane article's two gallery photos share a long caption that
inflates each block to the full viewport height (which also produced the
scroll-k snap loop fixed in `inline-image-rendering`). Captions at a standard
width look uniform across photos and wrap predictably regardless of photo
width, so the block height fits the viewport budget by construction.

## What Changes

- Render each image's attribution at a **standard caption width** (a fixed
  fraction of the content width, e.g. 70%), centered beneath the photo, instead
  of wrapping to the photo's own width. The attribution SHALL be constrained to
  the caption width for both the lead image and every inline image.
- Reserve the **real wrapped caption line count** in the image fit budget:
  the native render's height fit currently uses a shorter caption than the one
  actually composed (the inline link text is omitted), so a long caption
  overflows the reserved space and the composed block can exceed the viewport
  budget. The fit SHALL use the same caption the block composes with, so the
  block always fits.
- The two behaviors together keep an image-plus-caption block within the
  viewport budget even for long, identical gallery captions, so the two-stage
  scroll snap keeps a distinct bottom stage and the full-viewport collapse
  added in `inline-image-rendering` stops being hit by ordinary long captions.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `reader-ui`: the lead image attribution requirement changes from "constrained
  to the photo's own width" to a standard caption width, and the fit-budget
  requirement now reserves the composed caption's real line count.

## Impact

- `internal/ui/image.go`: `composeBlock`/`fitBlock` wrap the attribution at the
  standard caption width; `composeImageBlock` centers it accordingly;
  `inlineAttrFor`/`nativeRenderCmd` pass the composed caption (including link
  text) so the native fit reserves the real line count.
- `internal/ui/article.go`: inline attribution composition (the link-text path
  already knows the full caption) supplies the caption the fit must use.
- `convert/credit.go` is unchanged (caption extraction is independent of width).
- Block geometry changes ripple through the snap tests in
  `inline-image-rendering` (narrower captions shrink blocks); those tests are
  updated when this change is implemented.
## Open Questions

Resolved for this change. The attribution width is a fixed fraction of the
content width (70%) independent of the photo's width, applied to the lead and
every inline image. The height fit reserves the wrapped line count of the
caption the block actually composes with — including the inline link text that
`inlineAttrFor` currently omits — so a long caption can no longer overflow the
budget and inflate a block to the full viewport height (the condition that
forced the full-viewport scroll collapse in `inline-image-rendering`). The
specific fraction (70%) is a first guess; tune it against a real terminal in
the smoke test.

## Options

- Standard caption width as a fixed fraction of the content width (chosen) vs
  a fixed cell count. A fraction scales with the terminal; a fixed count is
  more predictable but does not track wide/narrow terminals.
- Reserve the composed caption for the fit (chosen) vs continuing to fit
  against a shortened caption and clamping the block. The composed-caption fit
  removes the overflow at the source; clamping hides the symptom and can shrink
  an image more than the budget requires.

## Problems

- The caption width interacts with the block geometry (`imgStart`, `capStart`,
  `imgEnd`) and every snap target in `inline-image-rendering`; shrinking
  captions changes those numbers, so the snap tests there must be re-verified
  (most will simply exercise a shorter caption).
- The native render fit (`nativeRenderCmd`) and the composition must agree on
  the caption, or the composed block still overflows. The link text is only
  known during composition, so the fit must receive the composed caption, not
  the shortened per-URL fallback.

## Details

### 1. Caption width

Introduce a caption width for the article content area, computed once per
geometry as a fraction of the content width (e.g.
`captionW = max(1, contentW*7/10)`). `composeBlock`/`fitBlock` (in
`internal/ui/image.go`) wrap the attribution at `captionW` instead of the
image's `imgW`, and center the wrapped lines beneath the block:

- `image.LayoutWrapLines(attr, captionW)` replaces `image.LayoutWrapLines(attr,
  imgW)` for the attribution.
- Each attribution line is centered within the content width (as today) and
  within the caption width (so the lines align under the caption, not the photo
  edges). The block's `imgEnd`/`capStart` accounting is unchanged: the caption
  lines still follow the image rows.
- `fitBlock`'s iteration reserves `attrLines` from `captionW` wrapping, so the
  image shrinks to fit the image plus the wrapped caption.

### 2. Fit uses the composed caption

`nativeRenderCmd` (and `inlineAttrFor`) currently fit the native render against
a caption without the inline link text, while `composeImageBlock` composes with
the full caption (alt → figure credit → link text → source). Make the native
fit use the same caption the block composes with:

- Track the composed caption per inline image (the `inlineImages` entries or a
  per-URL map built during `inlineBodyParts`), and have `nativeRenderCmd` fit
  against it.
- `articleAttribution`/`inlineAttribution` remain the single source of the
  caption text; only the fit input changes.

### 3. Gallery identical captions

Each image keeps its own attribution (per-image, as `inline-image-rendering`
requires); the standard width is what makes two identical long captions render
compactly instead of each inflating its block. No grouping or dedup is done for
identical captions in this change.

### Tests

- `internal/ui`: a long credit wraps at the caption width (not the photo width)
  and is centered; a long caption scales the image down so the block fits the
  viewport; an inline gallery of two narrow photos with an identical long
  caption yields two blocks that each fit the viewport budget.
- `inline-image-rendering` snap tests re-run: with shorter captions the block
  heights drop, so the two-stage snap keeps a distinct bottom stage for images
  that previously filled the viewport.
## Why

A short inline image (its block shorter than the viewport, as fitted images
almost always are) whose top a downward single-line scroll brings into the
window from below snaps straight to its TOP boundary, so the image appears
flush at the top of the screen and the next scroll skips it past its caption —
it is never shown at the viewport bottom first. The two-stage entry the inline
snap contract promises (bottom, then top, then skip) is bypassed for every
short image, which reads as a skipped photo when scrolling through an article
like the ProPublica Helene piece.

## What Changes

- A short inline image (block shorter than the viewport) whose top enters the
  window from below now snaps to its BOTTOM boundary instead of its TOP: the
  image and its full wrapped caption appear at the viewport bottom on entry.
- The next downward scroll rises the image to its TOP boundary (flush at the
  viewport top), and the following scroll skips the whole image-and-caption
  block onto the line after its caption, matching the two-stage entry already
  used for taller inline images.
- Upward scrolling is unchanged: entry still snaps the whole block flush to the
  viewport top, the next up scroll sinks it to the bottom boundary, and the
  following one scrolls it off — so the two directions stay mirrors of each
  other.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: the atomic image scroll behavior requirement changes how a short
  inline image's downward entry snaps — from flush-to-top to the bottom
  boundary (two-stage entry), matching the existing tall-image entry contract.

## Impact

- `internal/ui/compose/image.go`: `SnapYOffset`'s short-image down-entry snap
  targets the block's BOTTOM boundary instead of its TOP.
- `internal/ui/article_image_test.go`: the snap tests that assert the
  flush-to-top short-image entry (block-list, Helene gallery, and the
  scroll-through short-image test) update to the two-stage entry, plus a new
  test modeling the reported staircase-figure geometry.
- `openspec/specs/reader-ui/spec.md`: the "Atomic image scroll behavior"
  requirement's short-image entry sentence and the "Short inline image entered
  from above snaps flush to the viewport top" scenario are revised.
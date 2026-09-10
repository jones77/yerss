## Context

See proposal.md - Why. The article snap model in `internal/ui/compose/image.go`
treats a "shown on open" block (a genuine lead, or a suppressed lead whose URL
opens the body as its first inline image, e.g. ProPublica features) with
pre-consumed entry stages: a down press rises it flush to the viewport top, the
next skips the whole photo-and-caption block. Mid-article inline images snap in
two stages (bottom, then top, then skip). Before this change, `leadShown` was
set only when a genuine, non-suppressed lead block was composed, so a
suppressed-lead top-of-article photo fell through to the inline behavior even
though it occupies the first content rows and is fully visible on open.

## Goals / Non-Goals

**Goals:**
- Treat any first block composed directly below the header as shown on open for
  snapping (rise-to-top then skip on the way down).
- Snap the whole photo-and-caption block flush to the viewport top on an
  up-scroll as soon as its bottom enters the window, for every block including
  the shown-on-open top image.

**Non-Goals:**
- No change to multi-line (page/half-page/goto-bottom) scroll behavior.
- No change to the two-stage inline entry/exit transitions for mid-article
  images or to the promo-banner reuse case (still inline behavior).

## Decisions

**Derive `leadShown` from geometry, not from the genuine-lead path.** A block
is shown on open exactly when it is composed one row past the header's blank
line: `blocks[0].ImgStart == headerLines+1`. `composeArticle` now computes
`leadShown` from that condition after all blocks are built, replacing the
`leadShown = true` set only inside the genuine-lead branch. This covers the
suppressed-lead-at-top case (the block lands at `headerLines+1` through the
body splice) without touching the promo-banner case, whose first block sits
after body text (ImgStart > headerLines+1) and keeps its inline snapping.
Alternative considered: content-based detection in `computeDerivation` (whether
`bodyMD` begins with the image sentinel). Rejected as redundant: the composed
geometry already answers the exact question "was blocks[0] shown on open," and
works uniformly whether the block is a genuine lead, a suppressed-lead-at-top,
or a leadless article whose first content element is a photo.

**Apply the upward top-snap to every block.** The up-scroll "top snap" (block
whose bottom just entered the window from above snaps to its top boundary) was
gated `(i > 0 || !lead)`, excluding the shown-on-open top image and making its
caption scroll up line by line. Removing the gate lets the whole block appear
flush at the viewport top on the first up press that brings its bottom into the
window. The prior "narrowed upward reveal" (line-by-line caption) was a
deliberate lead behavior; the user's report makes the flush reveal the desired
behavior. The gate on the down-side rules (`(i > 0 || !lead)`) is unchanged —
the shown-on-open top image keeps its pre-consumed down entry.

## Risks / Trade-offs

- [Geometry-derived `leadShown` fires for a leadless article whose first
  element is a photo] → This is the desired semantics: that photo is shown on
  open, so pre-consumed entry is correct.
- [The upward top-snap applies to a block taller than the viewport, snapping to
  its top when its bottom enters] → The lead block is height-capped to fit the
  viewport during composition (fitBlock), so the block is at most viewport
  height; snapping to its top shows the whole block.
- [A genuine-lead test asserted the old line-by-line caption reveal] → Updated
  to assert the flush reveal; the snap no longer rests on a caption-only frame.
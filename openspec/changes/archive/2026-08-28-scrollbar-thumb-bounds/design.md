## Context

See `proposal.md` for motivation. Today the right-edge scroll indicator is
computed in `internal/ui/border.go` as a top-anchored fill:

```
interiorH := viewportH + 2*effPadY   // the right border's track height
fillH     := interiorH * percent / 100
right(row) = accent fill if row < fillH else dim unfill
```

`percent` is a 0–100 integer cached on `articleState` by `updateArticlePercent`
from `viewport.ScrollPercent()`. Because the fill starts at row 0 and its height
*is* the percentage, it is invisible at 0% and consumes the whole track at 100%.
The `bubbles/viewport` model already exposes everything a real thumb needs:
`TotalLineCount()`, `Height`, and `YOffset` (the raw scroll offset, `0..
max(0, totalH - viewportH)`).

## Goals / Non-Goals

**Goals:**
- Replace the growing fill with a movable, height-bounded thumb positioned by
  the live scroll offset, visible from the first frame of a scrollable article.
- Derive both the thumb geometry and the bottom-border percentage from one live
  scroll-state input so they can never disagree.
- Keep the short-article fully-filled / "100% scrolled" behavior and the
  degenerate height clamp unchanged.

**Non-Goals:**
- Click/drag-to-scroll on the thumb (mouse interaction). The viewport already
  supports the mouse wheel; thumb dragging is a separate change.
- Changing the border glyphs, colors, padding, or the "N% scrolled" label
  format. The thumb reuses the existing `fill`/`unfill` glyphs and accent/dim
  palette.
- Re-architecting the article resize path. Reading live viewport state is no
  worse than the cached percent there and is strictly more correct; the resize
  wiring itself is out of scope.

## Decisions

**D1 — Thumb geometry from live viewport state, not the cached percent.**
Replace the `percent int` input to `renderArticleBorder` with a small scroll
state carrying `{totalH, viewportH, offset}` read from the viewport at render
time (`TotalLineCount()`, `Height`, `YOffset`). The cached `articleState.percent`
field and `updateArticlePercent` are removed. *Rationale:* a 0–100 int is too
coarse to position a thumb when the scrollable range is small, and a cached
value can go stale on resize. Raw offset/total/height give exact geometry and
handle resize for free. *Alternative:* keep `percent` and add `totalH` alongside
— rejected; percent rounding loses precision and adds a second source of truth.

**D2 — Thumb proportional to the visible fraction, clamped to `[1, trackH-1]`.**
For a scrollable article (`totalH > viewportH`), with `trackH = interiorH`
(the right border's full extent between the top and bottom borders):

```
scrollableRange = totalH - viewportH
thumbH    = trackH * viewportH / totalH          // integer; visible fraction
thumbH    = clamp(thumbH, 1, max(1, trackH - 1)) // min 1 row, max trackH-1
thumbTop  = (trackH - thumbH) * offset / scrollableRange
```

A row `r` of the track is accent when `thumbTop <= r < thumbTop + thumbH`,
otherwise dim. *Rationale:* standard scrollbar semantics — thumb size encodes
"how much of the content is visible", position encodes "where you are". The min
of 1 guarantees the thumb is present on open (offset 0 → `thumbTop = 0`); the
max of `trackH-1` leaves at least one dim row so the thumb always reads as
movable, matching the user's stated bounds. *Alternative:* a fixed 1-row thumb —
rejected, it loses the "how much content" signal. *Alternative:* allow `trackH`
(full fill) on scrollable articles — rejected per the explicit
"maximum height of the text height - 1" requirement.

**D3 — Preserve the short-article fully-filled case.**
When `totalH <= viewportH` there is no scrollable range, so a thumb is
meaningless: the entire track is rendered accent and the bottom label reads
"100% scrolled". *Rationale:* keeps the existing tested scenario intact and
gives a clear "everything visible" signal. *Alternative:* render a `trackH-1`
thumb even for short articles — rejected as inconsistent with the existing
"100% / fully filled" contract and visually odd with nothing to scroll.

**D4 — `renderArticleBorder` derives the percentage label internally.**
The bottom-border "N% scrolled" text is computed inside the renderer from the
same scroll state: `100` for the fits-viewport case, else
`round(100 * offset / scrollableRange)`. This equals what
`updateArticlePercent` produced (`ScrollPercent()*100`), so the label is
unchanged. *Rationale:* one input drives both the thumb and the label, so they
can never diverge. The two direct callers — `renderArticle` and the
height-clamp test — are updated in lockstep; the height-clamp assertion (line
count `<= h`) is independent of scroll state.

## Risks / Trade-offs

- **[Integer rounding at mid-scroll]** → At `offset = 0`, `thumbTop = 0`
  exactly; at max offset, `thumbTop = trackH - thumbH` exactly. Only mid-range
  positions round, which is acceptable for a scrollbar. Tests assert exact
  placement at 0% and 100% and rounded placement at 42%.
- **[Thumb shrinks to 1 row on very tall articles]** → Expected and acceptable
  (huge content, tiny viewport → tiny thumb); the min-1 clamp keeps it visible.
- **[Signature change touches a direct test caller]** →
  `TestRenderArticleBorderHeightClamp` calls `renderArticleBorder` directly with
  `percent = 0`; it is updated to pass a scroll-state value. The assertion is
  unaffected.
- **[Stale viewport `Height` after resize]** → The article viewport is built
  with the current `vpH` along the same path the cached percent used, so live
  reading is no worse and is strictly more correct; no new resize wiring is
  introduced by this change.

## Migration Plan

No data, config, or API migration. The change is internal rendering behavior.
Rollback is a plain revert; no persisted state depends on the new geometry.

## Open Questions

- None. The thumb reuses the existing `fill`/`unfill` glyphs (in Unicode mode
  both are `║`, distinguished by accent vs dim color; in ASCII, `#` vs `:`), so
  no glyph decision is needed.

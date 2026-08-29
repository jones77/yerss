## Why

The article reader's right-edge scroll indicator is a *fill that grows from the
top*: it renders `interiorH * percent / 100` accent rows starting at row 0. On
open the percent is 0, so the indicator is invisible and only creeps in as the
user scrolls, eventually consuming the entire right border at 100%. This reads
as a progress bar, not a scrollbar — it gives no immediate sense of how much
content exists or where the viewport sits, and it is absent exactly when the
user first opens a long article.

## What Changes

- **Movable thumb instead of growing fill:** the right double-line border SHALL
  render a bounded *thumb* (a contiguous accent run) whose position reflects the
  current scroll offset, rather than a fill whose height equals the scroll
  percentage. The thumb SHALL be visible immediately on open (at the top of the
  track when scroll offset is 0), so the scrollbar is present from the first
  frame.
- **Thumb height bounds:** the thumb height SHALL be proportional to the visible
  fraction of the content (`trackH * viewportH / totalH`), clamped to a minimum
  of 1 row and a maximum of `trackH - 1` rows, where `trackH` is the right
  border's interior height. The thumb never fills the whole track on a
  scrollable article, so it always reads as a movable indicator.
- **Short-article exception preserved:** when the article content fits entirely
  within the viewport (nothing to scroll), the entire right border SHALL remain
  accent-filled and the bottom border SHALL still read "100% scrolled", matching
  the existing fully-visible behavior.
- **Single source of scroll state:** the thumb geometry and the bottom-border
  percentage SHALL both be derived at render time from the live viewport state
  (total lines, viewport height, scroll offset), replacing the cached
  `articleState.percent` field.
- **Tests:** add a thumb-position test (open a long article at 0% → a 1-row
  accent thumb sits at the top with the rest dim; scroll to ~50% → the thumb
  moves near the middle; scroll to 100% → the thumb sits at the bottom), a
  thumb-bounds test (thumb height stays within `[1, trackH-1]` across a range of
  content lengths), and keep the short-article "100% scrolled" / fully-filled
  assertion and the degenerate height-clamp assertion intact.

## Capabilities

### New Capabilities
<!-- None — the affected capability already exists. -->

### Modified Capabilities
- `reader-ui`: the "Article reader border rendering" requirement changes the
  right-edge scroll indicator from a top-anchored fill proportional to scroll
  percentage into a movable, height-bounded thumb positioned by scroll offset,
  with the short-article fully-filled case preserved.

## Impact

- **Code:** `internal/ui/border.go` (thumb geometry computation and rendering
  replacing `fillH`; `renderArticleBorder` signature gains scroll-state input
  and derives the bottom-border percent internally), `internal/ui/article.go`
  (build the scroll-state input from the viewport at render time; remove the
  cached `percent` field and `updateArticlePercent`), `internal/ui/render_test.go`
  and `internal/ui/render_safety_test.go` (update the direct
  `renderArticleBorder` caller and add thumb assertions).
- **APIs:** No public API changes. Internal: `renderArticleBorder`'s parameter
  list changes (the `percent int` argument is replaced by a scroll-state value),
  so direct test callers are updated in lockstep.
- **Dependencies:** None added. Reuses the existing `bubbles/viewport` model
  (`TotalLineCount`, `Height`, `YOffset`) and `lipgloss` styling already in use.
- **Behavior:** The scrollbar is visible from the moment a long article opens,
  moves as the user scrolls, and no longer grows to fill the right border. The
  "N% scrolled" bottom label and the short-article fully-filled indicator are
  unchanged. No storage, config, or first-run behavior changes.

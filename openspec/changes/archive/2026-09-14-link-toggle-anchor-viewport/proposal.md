## Why

Toggling a link's visibility with `x` (compact ↔ expanded) changes the number of
lines the link occupies, which shifts every line below it. Since URLs are longer
than their names, expanding has the reader jump while the toggle lands on a
different line than the one that was visible before.

## What Changes

- Toggling compact/expanded link mode preserves the reading position: whatever
  the first visible line was before the toggle stays the first visible line
  afterwards, instead of drifting to the offset predicted by the total line
  delta (which is only correct for insertions at the top of the article).

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `compact-link-mode`: the x-toggle requirement gains anchor-preserving scroll
  behavior while toggling between compact and expanded link display.

## Impact

- **Code:** `internal/ui/article_update.go` (the `LinkExpandToggle` handler),
  `internal/ui/image.go` (`recomposeArticle`), and tests in
  `internal/ui/link_expand_test.go`.
- **Behavior:** the x-toggle in the article reader no longer changes which line
  leads the viewport, except when the line count shrinks enough to force a scroll
  toward the bottom.
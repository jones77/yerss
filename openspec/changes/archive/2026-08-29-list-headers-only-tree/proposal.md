## Why

Every article row in the list view carries a `├─` tree-rail prefix, so the left
edge reads as noise with no hierarchy. The tree rail should appear only on the
collapsible day-group headers, letting article rows start directly at their
publication time for a cleaner list.

## What Changes

- Article rows lose the tree rail (`corner + dash + space`) entirely and render
  as `HH:MM Title ... source`, with the time flush at column 0.
- Day headers keep the rail, with the corner based on the day group's position:
  first group `┌`, last group `└`, interior groups `├` (ASCII `+`).
- The three columns freed by removing the rail expand the title width.
- Selection highlighting is unchanged; article rows simply begin at the time
  instead of at a glyph column.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: The "Article list rendering" requirement changes so the tree
  rail appears only on day-group headers (with `┌`/`└` bookends on the first
  and last groups) and article rows render without a rail, starting at the
  publication time; the associated rail scenarios and the selection clause are
  updated accordingly.

## Impact

- `internal/ui/list.go` — `renderArticleRow` (drop rail, widen title),
  `railGlyph` (group-based corners), `renderList` (corner only on headers).
- `internal/ui/polish_test.go` — tests that pass a corner to
  `renderArticleRow`.
- Spec: `openspec/specs/reader-ui/spec.md` (delta).
## Why

The list view mixes formatting concerns: a cursor gutter, middle-dot bullets,
and the publication time crowding the right edge next to the source. A tree-rail
layout with time on the left and the news org right-aligned reads cleaner, and
whole-row selection removes the redundant gutter marker. Day labels also read
better with an explicit "today/yesterday" cue.

## What Changes

- **Tree rail**: the list renders as a continuous tree — `┌` prefixes the first
  row of the whole list, `└` the last row, `├` every other row; article rows
  carry a dash after the corner (`├─`), day headers a space. Day headers lose
  their fold markers. ASCII fallbacks (`+`, `-`) per the glyph-sync rule.
- **Row format**: `HH:MM` follows the tree glyph on the left, then the title;
  the news org is right-aligned at the content edge and is the only field on
  the right (no bullets, no time). A truncated title ends with an ellipsis and
  always keeps at least one space before the org.
- **Day labels**: the existing long-date format is kept; the current local day
  is prefixed `today, ` and the previous local day `yesterday, `. `Undated` is
  unchanged.
- **Selection**: the `> ` cursor gutter is removed; selection is shown as a
  full-row background highlight (the style day headers already use), extended
  to article rows. Fold keys and selection wrap-around are unchanged.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `reader-ui`: article list rendering — tree rail glyphs, row field layout,
  day header labels, and selection indication.

## Impact

- `internal/ui/list.go` — `bucketDayGroups` (labels, with an injected reference
  time), `renderDayHeader`, `renderArticleRow` (rail, row layout, selection).
- `internal/ui/border.go` — `glyphsFor` gains the tee glyph (`├` / `+`); `┌`,
  `└`, `─` reuse existing fields. ASCII/Unicode pairs stay in sync per
  AGENTS.md.
- Tests: `daygroups_test`, `render_test`, `polish_test` updated; new tests for
  rail glyphs (Unicode + ASCII), label prefixes, and truncation gap.
- Independent of `article-view-rework`; either may land first (both touch
  `glyphsFor` and list rendering tests, so rebase order matters, not scope).

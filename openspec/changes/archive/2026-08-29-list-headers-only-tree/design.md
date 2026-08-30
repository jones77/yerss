## Context

See `proposal.md` — Why. In `internal/ui/list.go`, `renderArticleRow`
(currently line 284) prefixes every article row with a tree rail
(`corner + "─" + " "`), and `railGlyph` (line 265) picks the corner from the
flat visible-row index, so every row — header or article — carries a glyph.
The day header renders `corner + " " + label` (line 281). The rail and header
corners are rendered in the grey role (`palette.Dim`).

## Goals / Non-Goals

**Goals:**
- Article rows render without any tree glyph, starting at the publication time.
- Day headers keep the rail with `┌`/`└`/`├` bookends chosen per day group.
- Selection highlight still covers the full row on both row kinds.

**Non-Goals:**
- No change to day-group bucketing, collapse behavior, or the source-identifier
  derivation.
- No new glyphs; the `glyphsFor` Unicode/ASCII table is untouched (article rows
  simply stop using `g.h` and the corner).

## Decisions

### 1. Article rows drop the rail entirely

`renderArticleRow` loses its `corner` parameter and the `rail := corner + g.h +
" "` prefix. The row becomes `ts + " " + title ... src`, rendered in the same
style order as today (dim time, bar separator, styled title, dim source). The
three columns freed by the rail go to the title: `railW` in the width math
drops from `railW + ts + 1` to `ts + 1`.

Rationale: the user asked for flush-left article rows (time at column 0). An
alternative — indenting rows to align under the header text — was considered
and rejected by the user.

### 2. Corners are chosen per day group, not per flat row

`railGlyph(n, i)` is replaced by a group-based corner selection: group index 0
→ `┌` (`g.tl`), last group → `└` (`g.bl`), else `├` (`g.tee`). `renderList`
already iterates groups; the header row passes its group's index and the group
count. Because bookends now derive from group position, the "glyphs are stable
while scrolling" behavior is automatic: they never track the window.

`renderArticleRow` no longer needs a corner; `renderList` stops passing one.
The article row's leading element becomes the dim time.

### 3. Selection and tests

The selected-row styles are unchanged; article rows simply begin with the dim
time, so `dim.Render(ts)` already carries the grey foreground plus the
selection background when selected. In `polish_test.go`, the four tests that
call `renderArticleRow(item, corner, selected)` drop the corner argument, and
`TestArticleRowSelectionBackground` keeps asserting the first escape is the
grey+background prefix (now from the time instead of the rail). No other
renderers change.

## Risks / Trade-offs

- [A list with article rows but no visible glyph may read as "flat"] → That is
  the requested design; day headers still branch the tree.
- [Removing the corner parameter touches four tests and the render call sites]
  → Mechanical; the compiler catches every reference.
- [Single-group lists collapse to `┌`/`└` on one header] → `railGlyph` treats a
  lone group as both first and last; it renders `┌` (first wins), matching the
  existing first-row-wins behavior.

## Migration Plan

None — code-only change; the reader UI spec delta covers the requirement
update. Rollback is a revert of `internal/ui/list.go`, `polish_test.go`, and
the `reader-ui` spec.

## Open Questions

None.
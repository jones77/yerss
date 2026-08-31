## Context

See `proposal.md` — Why for motivation. Current state (from code):

- Day-group folds live on `dayGroup.collapsed`; `h` collapses, `l` expands, and
  Enter/Space/Tab toggle the group under the list cursor (`updateList` in
  `internal/ui/list.go`).
- Actions are declared once in the catalog (`internal/config/keys.go`); the
  help popup, seeded config template, and runtime key resolution all derive
  from it.

## Goals / Non-Goals

**Goals:**
- One `x` key in the list view that switches between all-expanded and
  all-collapsed, defaulting to all-expanded in a mixed state.
- The action is a remappable catalog action, listed in the help popup under
  List view.

**Non-Goals:**
- Any change to single-group fold keys, the article view, or the popups.

## Decisions

### Toggle semantics

Pressing `x` collapses every day group when all are already expanded, and
expands every day group otherwise — so a fully collapsed list expands and a
mixed list expands. This matches the requirement that a mixed state defaults
to expand all.

### Implementation

`expandToggle` scans `m.list.groups` for any collapsed group, then sets every
group's `collapsed` flag to the inverse of that finding and clamps the cursor
(the visible-row list may shrink or grow). The action is dispatched in
`updateList` like the other list actions, independent of the cursor row.
## Why

Day groups can be folded one at a time with `h`/`l`/Enter/Space/Tab, but there
is no way to fold or unfold the whole list at once. A single `x` key that
expands all day groups or collapses them all makes a long feed list scannable
at a glance and restorable in one keystroke.

## What Changes

- A new `expand_toggle` action bound to `x` in the list view toggles the day-group
  folds between all-expanded and all-collapsed.
- When the day groups are in a mixed state (some expanded, some collapsed), the
  toggle defaults to expanding all.
- The action appears in the help popup under List view and in the seeded config
  template like every other binding.
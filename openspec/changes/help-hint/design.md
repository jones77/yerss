## Context

See `proposal.md` — Why for motivation. Current state (from code):

- `renderStatusBar` (`internal/ui/list.go`) builds the bottom line as
  `left + strings.Repeat(" ", pad) + right` where `pad` is the leftover width,
  then truncates and styles the whole line.
- `bottomBorder` (`internal/ui/border.go`) builds the article frame's bottom
  edge as `└─ <hint> <dashes> <indicator> ─┘`, computing the dash fill from
  the leftover interior width (`fill = inner - 6 - hintW - indicatorW`).
- `?` is bound to `Help` in both `ViewList` and `ViewArticle` (`internal/
  config/keys.go`).

## Goals / Non-Goals

**Goals:**
- A literal `?: help` hint as the first right-aligned element of the list view
  status bar and the article view bottom border, separated from the following
  element by the bullet.
- Identical behavior in Unicode and ASCII fallback modes (the hint is plain
  ASCII text and uses the existing bullet glyph pair).

**Non-Goals:**
- Making the hint reflect the configured keybinding (hardcoded `?`).
- Changes to the help popup, tag popup bottom border, or keymap.

## Decisions

### First right-aligned element in the list view status bar

`renderStatusBar` prepends the help hint to the existing right label, separated
by the bullet in the dim role: `?: help · <size> · <percent>% · <n>/<total>`.
The gap between the left refresh/status text and the right group stays a plain
space pad, so the hint reads as the first right-aligned element.

### First right-aligned element in the article bottom border

`bottomBorder` builds the position indicator as
`?: help · <percent>% · <bottomLine>/<totalLines>` with every bullet in the dim
role. Width accounting is unchanged: the left hint is truncated first, then the
indicator, then the dash fill is computed from the remaining interior width.
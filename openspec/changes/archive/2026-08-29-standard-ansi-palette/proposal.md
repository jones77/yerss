## Why

The TUI currently uses bespoke hex palettes (Tokyo-Night-style dark, a
hand-picked light) under arbitrary names (`Accent`, `Dim`, `Bold`), with
comments calling colors "baby blue". Colors should come from the standard
terminal palette so the reader matches the user's terminal theme and the code
names match the standard terminology.

## What Changes

- Replace the hex palette values with standard ANSI color numbers:
  - **8** dark grey — article border, rails, bullets, scrollbar track, and all
    muted/metadata text (read titles, timestamps, sources, day headers, image
    attribution, popup URLs).
  - **12** bright blue — status bar, border inline text (date/title/hint/percent/
    ratio), scrollbar thumb, and popup titles/borders.
  - **7** white — article body text (forced via glamour base style override).
  - **15** bright white — unread articles in the list view and bold tag counts.
- Rename palette fields to standard roles (`Accent`/`StatusBar` merge into the
  blue role; `Bold` becomes `Bright`; `Border` folds into the grey `Dim` role) and
  update the "baby blue" comments to standard names.
- Light theme keeps its `auto`/`light`/`dark` behavior but uses dark-variant
  ANSI values (body = 0 black, unread = bold black, blue = 4, grey = 8).
- The selected-row highlight keeps its `#333333` background (unchanged; a
  background, not covered by ANSI foreground numbering).

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: The color requirements change from arbitrary hex palettes to
  standard ANSI color numbers (8/12/7/15 in dark mode, dark-variant values in
  light mode); article body text is explicitly white (7) / black (0); the
  dim/grey role covers both the border and muted text; the bright role covers
  unread titles.

## Impact

- `internal/ui/theme.go` — palette definitions and field renames.
- `internal/ui/border.go` — field references and "baby blue" comments.
- `internal/ui/list.go` — read/unread title styling and field references.
- `internal/ui/popup.go` — Accent→blue role for titles/borders.
- `internal/ui/image.go` — attribution styling (grey role).
- `internal/ui/markdown.go` — glamour base document foreground override.
- `internal/ui/polish_test.go` — test references to renamed fields.
- Spec: `openspec/specs/reader-ui/spec.md` (delta).
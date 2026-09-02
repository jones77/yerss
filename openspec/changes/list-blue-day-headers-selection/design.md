## Context

The list view's day-group headers are rendered by `renderDayHeader` in
`internal/ui/list.go` as `{corner} {label}` using `m.styles.dim`/`selDim`
(grey role). The palette in `internal/ui/render/theme.go` sets the selection
background to a fixed `#707070` in both modes, with the selected-text
foregrounds (`selText`/`selBright`/`selDim`) derived from the role. The
precomputed `modelStyles` in `internal/ui/model.go` cache all render styles per
frame. See proposal.md — Why for the motivation.

## Goals / Non-Goals

**Goals:**
- Day headers render as `{corner}─── {date}` (ASCII `+---`) entirely in the
  blue role, matching the status bar.
- The selection highlight becomes theme-aware (`#242424` dark / `#d3d3d3`
  light) without changing any selected-text foregrounds.
- Keep the Unicode/ASCII glyph pair in sync and avoid per-frame style
  construction.

**Non-Goals:**
- No change to article-row layout, fold markers, or selection semantics.
- No new palette roles or config options; no changes to the `Dim` role.

## Decisions

### D1 — Reuse the `H` dash glyph for the connector

`renderDayHeader` builds the prefix as `corner + strings.Repeat(g.H, 4) + " "`.
`g.H` is already `─` in Unicode mode and `-` in ASCII mode (via
`render.GlyphsFor`/`m.glyphs()`), so the connector stays in sync with the
existing ASCII fallback without a new literal pair. Four dashes place the
header's date at the same column as the status bar's date (after the `HH:MM `
time prefix), aligning the two date reads.

*Alternatives:* hardcoding `"───"`/`"---"` literals — rejected because it
would duplicate the glyph mapping AGENTS.md requires the repo to keep in one
place (`internal/ui/border.go` + `render.GlyphsFor`).

### D2 — Blue role via `status` and a new `selStatus` style

Unselected headers render with `m.styles.status` (Foreground `StatusBar`),
the same precomputed style the help bar uses. A new `selStatus` style —
`Foreground(StatusBar).Background(Selection)` — is added to `modelStyles` and
built in `buildModelStyles`, so a selected header renders blue-on-selection in
one escape sequence, mirroring how `selText`/`selDim` already work.

*Alternatives:* composing a fresh `lipgloss.NewStyle()` in `renderDayHeader`
for the selected case — rejected because the codebase deliberately precomputes
styles to avoid per-frame construction; a one-off style would also let the
selected case drift out of sync with the palette roles.

### D3 — Theme-aware selection background via the `Selection` role

`DarkPalette` sets `Selection: lipgloss.Color("#242424")` and `LightPalette`
sets `Selection: lipgloss.Color("#d3d3d3")`. No selected-text foregrounds
change: white on `#242424` and black on `#d3d3d3` both keep adequate contrast,
so `selText`/`selBright`/`selDim` stay valid in both themes. Because every
selection site already consumes `p.Selection` through `buildModelStyles` (list
rows, day headers, tag popup cells), a single palette change propagates to all
of them.

*Alternatives:* keeping a fixed `#707070` — rejected by the user's requirement.
Per-theme selected-text variants (`selTextDark`/`selTextLight`) — rejected as
unnecessary churn given the two themes pair black or white foregrounds with
the new backgrounds and both combinations are readable.

## Risks / Trade-offs

- [Style cache staleness] → `modelStyles` is rebuilt in `New` from the
  resolved palette; the palette is resolved once at startup (a theme switch
  requires a restart), so the new `selStatus` style cannot go stale. If a
  future theme-hot-swap feature lands, `buildModelStyles` must be re-run where
  `palette` is reassigned (already documented in the archived
  `palette-selection-role` change).
- [Test churn from the color change] → Tests that hardcode `#707070`
  (`polish_test.go`, `tags_test.go`) assert the dark-theme selection; they
  become `#242424`. Render tests run against the dark palette, so the change
  is localized to those literals.
- [Light-theme contrast] → The light palette already uses ANSI 0 (black)
  foregrounds, so `#d3d3d3` keeps the selected row readable; no new
  foreground logic is introduced.
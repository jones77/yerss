## Context

`Palette` has four roles; selection background duplicates the dim value literally. Styles like `lipgloss.NewStyle().Foreground(m.palette.Dim)` are rebuilt ~25 times per frame. See proposal.md.

## Goals / Non-Goals

**Goals:** one selection role; styles built once; pixel-identical output.

**Non-Goals:** no color changes.

## Decisions

### D1 — `Selection` role on `Palette`

`darkPalette`/`lightPalette` both set `Selection: lipgloss.Color("#707070")`. Rationale: mirrors the existing role pattern and the spec's fixed-grey selection background.

### D2 — Precomputed style fields on `Model`

A small struct of `lipgloss.Style` built in `New` and refreshed when the theme or ASCII mode changes. Rationale: styles are immutable once built; rebuilding per frame is waste.

## Risks / Trade-offs

- [Style cache staleness on theme switch] → Rebuild the cache wherever `palette`/`ascii` is reassigned.

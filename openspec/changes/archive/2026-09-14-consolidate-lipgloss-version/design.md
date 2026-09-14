## Context

Two lipgloss majors coexist: v1 (`github.com/charmbracelet/lipgloss`) powers `render.Palette` (`lipgloss.Color` fields), `modelStyles`, and border/popup styling; v2 (`charm.land/lipgloss/v2`) appears only in `compose/reflow.go` for `lipgloss.Wrap`. `go.mod` declares both as direct deps. The `charm.land/*/v2` module line is the maintained one and is already present for `glamour/v2` and `x/mosaic`.

Because the two majors are not API-identical, this cannot be a mechanical import swap; it needs a short spike to confirm API parity (especially `Color`, `Style`, `Width`, `RoundedBorder`, and word-wrap semantics), then a migration.

## Goals / Non-Goals

**Goals:** one lipgloss major version in app code; rendered bytes unchanged; `go.mod` no longer declares the losing major (`charm.land/lipgloss/v2`) as a direct dependency.

**Non-Goals:** no visual or styling behavior changes; no removal of `charm.land/*/v2` if `glamour`/`mosaic` still need it transitively; no change to `render/width.go` if it already uses `x/ansi`/`runewidth` directly.

## Decisions

### D1 — Default target is `charm.land/lipgloss/v2`

Recommended because it is the maintained line and already present. The spike may reverse this if v2 lacks a behavior-equivalent API for something the code relies on, in which case the fallback is to replace `compose/reflow.go`'s lone v2 `Wrap` with a v1/`x/ansi` equivalent and drop v2 instead.

### D2 — Spike gate before any migration

The first task is a spike: enumerate v1 usages, identify v2 equivalents, and — critically — confirm the styled-output tests pass byte-for-byte after a prototype swap. No migration proceeds until the spike resolves the direction.

### D3 — Byte-identity gate

`polish_test.go`, `ascii_test.go`, `render_test.go`, and the border/popup rendering tests are the regression net; a migration step is complete only when they pass unchanged. Any intentional rendering delta (there should be none) is a scope change, not part of this refactor.

## Spike outcome

The spike (tasks 1.1–1.3) resolved the direction as **D1 fallback: v1 stays, v2 leaves app code**:

- v1 and v2 differ in their ANSI reset byte: v1 renders `\x1b[0m`, v2 renders `\x1b[m`
  (hardcoded `ResetStyle` in `charmbracelet/x/ansi`, no config knob on either v2
  version). A v2 migration would therefore change the rendered bytes of every
  styled element app-wide, breaking the D3 byte-identity gate
  (`TestArticleFrameBorderColors` in `polish_test.go` compares full rendered
  strings built from v1 lipgloss). v2 has no behavior-equivalent for the v1
  renderer, so D1's spike-reversal clause applies.
- `lipgloss.Wrap` (the lone v2 use) has a stable equivalent: `xansi.Wrap` from
  the already-direct `github.com/charmbracelet/x/ansi`. Swapping it into
  `compose/reflow.go` passes all `internal/ui/compose` tests including
  `TestReflowAfterStripWrappedHref` (OSC 8 hyperlinks preserved). No padding is
  added by reflow (the viewport pads lines), so x/ansi's no-pad wrap semantics
  are a match.
- Direction: **keep `github.com/charmbracelet/lipgloss` v1 everywhere; replace
  `compose/reflow.go`'s v2 `Wrap` with `xansi.Wrap`; drop
  `charm.land/lipgloss/v2` from `go.mod` as a direct dep** (it remains
  transitively via `glamour/v2`/`x/mosaic`).

## Risks / Trade-offs

- [v2 API parity is unproven] → D1/D2 make the direction a spike outcome, not an assumption.
- [Migration touches many style sites] → Mechanical; grouped by package with a test run after each.
- [v2 remains in the graph transitively] → Acceptable; the goal is one version in app code and dropping the v1 *direct* dep, not a single lipgloss in the whole graph.

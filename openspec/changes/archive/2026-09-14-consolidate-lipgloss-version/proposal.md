## Why

The codebase imports two lipgloss major versions: `github.com/charmbracelet/lipgloss` (v1), used broadly across `render/` (the `Palette` type and border/theme styling) and the main `ui` package (styles, borders, width helpers); and `charm.land/lipgloss/v2`, used in exactly one file (`compose/reflow.go`, a single `lipgloss.Wrap` call). Both are direct dependencies in `go.mod`, so the styling layer is split across two majors. Collapsing to one version removes the split-brain and aligns styling with the `charm.land/*/v2` line already in use (`glamour/v2`, `x/mosaic`).

## What Changes

- Spike first: determine the correct target (v2 is the recommended default, being the maintained `charm.land` line) and map every v1 API use (`lipgloss.Color`, `Style` construction, `Width`, `RoundedBorder`) to its v2 equivalent.
- Migrate `render/` and the main `ui` package's v1 styling usages to the chosen version.
- Remove the v1 direct dependency from `go.mod` (v2 may remain transitively via `glamour/v2`/`mosaic`; the goal is a single version the app code imports).
- No rendered-output change: the styling test suites (`polish_test.go`, `ascii_test.go`, `render_test.go`) are the byte-identity gate.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `go.mod`: `github.com/charmbracelet/lipgloss` v1 removed.
- `internal/ui/render/theme.go`, `border.go`: v1 → target version.
- `internal/ui/model.go`, `border.go`, `popup.go`, `image.go`, `markdown.go`: v1 → target version.
- `internal/ui/compose/reflow.go`: either stays on v2 (if v2 wins) or moves off it (if the spike favors v1).

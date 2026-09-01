## Why

The selection background `#707070` is hardcoded in three render sites instead of being a palette role, and the TUI rebuilds lipgloss styles on every frame. Adding a `Selection` role and precomputing styles removes the magic number and the per-frame churn with no visible change.

## What Changes

- Add a `Selection` color to `Palette` (`#707070` in both modes).
- Replace the hardcoded `#707070` literals in the list, tag popup, and article selection rendering with the role.
- Precompute the repeated lipgloss styles once on the model instead of constructing them per render.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/theme.go`: palette role.
- `internal/ui/model.go`: precomputed styles.
- `internal/ui/list.go`, `internal/ui/popup.go`, `internal/ui/article.go`: use the role/styles.

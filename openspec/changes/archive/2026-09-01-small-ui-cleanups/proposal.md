## Why

A handful of small smells remain after the larger refactors: a two-line `sourceID` wrapper that only forwards to `store.SourceLabel`, floating-point percentage math where integer math is exact, a swallowed store error in `Init`, and hardcoded header fold-key strings scattered through `updateList`.

## What Changes

- Delete the `sourceID` wrapper and call `store.SourceLabel` directly.
- Replace the float percent math with integer math.
- Surface `LastRefreshedAt` errors as a status message instead of ignoring them.
- Extract the header fold-key strings into named constants (full keybinding-catalog integration is deferred: fold keys are context-sensitive and would need new actions and conflict handling).
- No visible behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/list.go`, `internal/ui/model.go`, `internal/ui/popup.go`: small cleanups.

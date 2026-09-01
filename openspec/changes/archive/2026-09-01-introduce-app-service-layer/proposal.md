## Why

The TUI model reaches directly into `store`, `feed`, `image`, and `config`, which keeps orchestration logic (refresh, read-state, selection, status) inside the view layer. An application layer gives the view one dependency and a clear boundary.

## What Changes

- Add `internal/app` with a `Session` type that wraps the store, config, and feed/image coordination.
- Move refresh orchestration, read-state, selection persistence, and image-load coordination out of `ui.Model` into `Session` methods (or behind its API).
- `ui.Model` depends on `*app.Session` instead of the four packages directly.
- No behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/app`: new package.
- `internal/ui/model.go`: dependency inversion.
- `cmd/yerss/main.go`: constructs the session and hands it to `ui.New`.

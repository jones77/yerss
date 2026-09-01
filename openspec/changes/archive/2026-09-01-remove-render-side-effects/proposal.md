## Why

`View()` mutates model state: it calls `nativeImageClear`, which writes `m.nativeSent` and clears the map during rendering, and it emits a kitty delete-all every frame when an article has no native images. Render should be a pure function of state.

## What Changes

- Move kitty cleanup decisions out of `View()` into the update path (or into a pure per-frame plan that does not mutate the model).
- Stop emitting the kitty delete-all clear when the article has no native images.
- No visible behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/model.go`, `internal/ui/image.go`: render purity.

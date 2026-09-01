## 1. Pure render

- [x] 1.1 Move `nativeImageClear` computation into the update path and store its result on the model
- [x] 1.2 Change `View()` to prepend the stored prefix without mutating state

## 2. Remove per-frame delete-all

- [x] 2.1 Emit the delete-all only on view transitions, not on every no-native-image frame

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...` (kitty placement/cleanup tests)

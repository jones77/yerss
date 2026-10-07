## 1. Reproduction

- [x] 1.1 Reproduce the invoice photo's real block geometry from the local SQLite DB (article id 10431, image position 1) at a ~106x31 viewport, rendering native blocks via the real height fit
- [x] 1.2 Simulate the single-line scrolls and pin whether the blank region comes from `snapYOffset` landing in a partial-visibility state or from native-placement cleanup deleting a partially visible placement, and add a focused failing test

## 2. Fix

- [x] 2.1 Fill the visible rows of a partially visible native photo with a preview (reusing the halfblock `previewClippedImage` approach) so no blank strip appears
- [x] 2.2 Keep fully visible native photos rendering natively with no preview substitution

## 3. Verification

- [x] 3.1 Add/update tests asserting a tall-than-viewport photo shows a preview of its visible rows, never blank space, and a fully visible photo stays native
- [x] 3.2 Run `go build ./...` and `go test ./...` to confirm the fix compiles and passes

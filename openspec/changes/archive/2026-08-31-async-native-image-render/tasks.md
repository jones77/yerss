## 1. Image package: native render cache and command

- [x] 1.1 Add `Natives` in-memory cache (render-size key → rendered lines) and `NativeKey(url, width, maxHeight)` helper, mirroring `Blocks`/`Photos`
- [x] 1.2 Add `NativeMsg{Key, Lines}` message type
- [x] 1.3 Add `NativeCmd(native, photos, natives, url, width, maxHeight, attr, headerLines)` that reads cached bytes, runs the height-fit loop, renders natively, caches, and returns `NativeMsg` (short-circuit on cache hit)
- [x] 1.4 Move the wrap-line-count logic needed by the fit loop into the image package (shared via `ansi.StringWidth`) so the command and the UI's `composeBlock` agree
- [x] 1.5 Remove the `maxImageBytes` cap from `fetch()` (keep `fetchTimeout` and content-type allowlist); remove the now-unused var
- [x] 1.6 Add/extend image tests: `Natives`/`NativeKey`, `NativeCmd` cache-hit short-circuit, off-thread render output, fit-loop convergence, and fetch without the byte cap

## 2. UI integration

- [x] 2.1 Add `imgNatives *image.Natives` to `Model` and initialize it in `New`
- [x] 2.2 Rework `articleImageBlock` native branch to read from `imgNatives` and compose, with no rendering
- [x] 2.3 Change `onPhotoLoaded` to fire `NativeCmd` (instead of relying on on-thread render) and add an `onNativeLoaded` handler that recomposes on `NativeMsg`
- [x] 2.4 Add the `Natives` check to `hasImageSource`/`ensureImageSource` so resize re-renders off-thread at the new width
- [x] 2.5 Update `Update`'s message switch to handle `NativeMsg` and clear the in-flight guard
- [x] 2.6 Update UI tests (`article_image_test.go`) for the new `PhotoMsg → NativeCmd → NativeMsg` flow and cached-native-render recompose

## 3. Verification

- [x] 3.1 Run `gofmt`, `go vet`, `go build`, and `go test ./...` and fix failures
- [ ] 3.2 Manual smoke test: opening a high-resolution photo on a native terminal does not freeze the UI; re-opening the article at the same width shows the photo instantly; resizing re-renders without re-fetching

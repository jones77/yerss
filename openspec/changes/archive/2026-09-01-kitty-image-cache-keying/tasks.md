## 1. Per-render kitty image id

- [x] 1.1 Add `renderID(url string, pxW, pxH int) uint32` (never zero) and use it for the kitty image id in `Render`/`escape`.
- [x] 1.2 Replace `DeletePlacement(url)` with `DeleteByID(id uint32)`; drop the now-unused URL-only `stableID`.
- [x] 1.3 Update id-shape tests (`kittyTransmit`, `DeleteByID`, `renderID`, `NativeRenderID`).

## 2. Record the id on each native block

- [x] 2.1 Add `nativeID uint32` to `imageBlock` and populate it from the native render's returned id.
- [x] 2.2 Extract the id from the composed native render's own escape (`NativeRenderID`) at composition so the block records the exact id it was transmitted under — the composition reads the block from the `imgNatives` cache by key, so threading it through `NativeMsg`/`NativeCmd` would not reach the block.

## 3. Delete prior id on re-render and on exit

- [x] 3.1 In `nativeImageClear` and the resize re-render path, delete by the block's recorded id, including deleting the prior id before a new-size transmit.
- [x] 3.2 Ensure leaving the article emits `d=i` for every native block's recorded id.

## 4. Tests

- [x] 4.1 Add a test that two sizes of the same URL produce distinct ids and the same size reuses the id.
- [x] 4.2 Add a test that a new-size re-render emits a delete for the prior id before the transmit.
- [x] 4.3 Add a test that article exit emits `d=i` for each recorded id.

## 5. Validation

- [x] 5.1 Run `go test ./...` and `go vet ./...` and fix regressions.
- [x] 5.2 Manually confirm (Ghostty) that resizing an open photo and exiting articles no longer grows the terminal's memory.

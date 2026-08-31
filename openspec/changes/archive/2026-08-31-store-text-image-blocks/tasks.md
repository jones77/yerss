## 1. Store: image block persistence

- [x] 1.1 Add `image_block` column migration guarded by `PRAGMA table_info`, mirroring `migrateImageColumns`
- [x] 1.2 Add `SetImageBlock(id, block)` (writes block, clears `image_data`) and `GetImageBlock(id)`; remove `SetImageData`; keep `GetImageData` read-only for legacy
- [x] 1.3 Update store tests: column assertions, block round-trip, legacy `image_data` still readable and cleared by `SetImageBlock`

## 2. Image package: caches, commands, native rendering

- [x] 2.1 Add `Blocks` (URL → rendered lines) and `Photos` (URL → raw bytes) in-memory session caches
- [x] 2.2 Add `Protocol` type, `DetectProtocol()` env detection (iTerm2/WezTerm → OSC 1337, kitty/ghostty → kitty, else none, tmux-gated), and an overridable `detectProtocol` var
- [x] 2.3 Add `BlockWidth(lines)` helper returning the block's render width
- [x] 2.4 Replace `LoadedMsg` with `BlockMsg{Key, Lines}` and `PhotoMsg{Key}`; keep `FailedMsg`
- [x] 2.5 Implement `BlockCmd` hierarchy: decoded cache → width-matched stored block → legacy `image_data` (decode, render, persist, purge) → network fetch (persist block, cache decoded); `fetch=false` skips the network for the native placeholder path
- [x] 2.6 Implement `PhotoCmd`: fetch photo into `Photos`, reusing the fetch to render/persist the placeholder block when not already decoded; returns `PhotoMsg`/`FailedMsg`
- [x] 2.7 Add `NativeRenderer`: decode → scale to display size → PNG encode → emit OSC 1337 (cell-sized) or kitty escape, stable URL-derived name/id, block lines padded to content width with reserved rows
- [x] 2.8 Update/extend image tests: caches, block-width matching, block cmd hierarchy (stored-block hit, width mismatch re-fetch, legacy conversion + purge, no-fetch mode), photo cmd, native renderer line/escape output, protocol detection

## 3. UI integration

- [x] 3.1 Add `imgBlocks`, `imgPhotos`, `imgProtocol` fields to `Model` and a protocol setter mirroring `SetAscii`
- [x] 3.2 Rework `articleImageBlock` source precedence: native terminals use `Photos` (native lines) then `Blocks`; non-native use decoded `imgCache` re-render then stored `Blocks`
- [x] 3.3 Split `fireImageLoad` into `BlockCmd` (non-native) and `NativeCmd` (native) firing
- [x] 3.4 Replace `onImageLoaded` with `onBlockLoaded`/`onPhotoLoaded` (recompose preserving scroll offset); make `onImageFailed` recompose when a block became available
- [x] 3.5 Handle resize: recompose as today, re-fire the load command when the article's image has no session source matching the new width
- [x] 3.6 Update UI tests (`article_image_test.go`) for the new message flow and native placeholder-before-photo rendering

## 4. Verification

- [x] 4.1 Run `gofmt`, `go vet`, `go build`, and `go test ./...` and fix failures
- [x] 4.2 Manual smoke test: reopen an article offline renders the stored block without a request; on a native-capable terminal the block appears before the photo replaces it

## 5. Native placement cleanup on view change

- [x] 5.1 Add `NativeRenderer.Clear()` emitting `<ESC>_Ga=d,d=a,q=1<ESC>\` (delete all visible placements, quiet) for kitty and "" for OSC 1337/none
- [x] 5.2 Report the native render path from `articleImageBlock` and record it as `articleState.nativeImg`
- [x] 5.3 Prepend the clear to any `View()` frame that does not display the open article's native image (other views, placeholder frames, image scrolled out of the window); visible-image frames omit it and scrolled-back images re-transmit
- [x] 5.4 Tests: `Clear()` per protocol; kitty frame keeps a visible image, deletes on scroll-out/back-out/placeholder; iTerm frames never carry the delete
- [x] 5.5 Verify: `go vet ./...`, `go build ./...`, `go test ./...` all pass

## 6. Native photo centering

- [x] 6.1 Pad the native block's reserved rows (and the escape line) to the fitted image cell width instead of the content width, so the block spans the photo's own cell box
- [x] 6.2 Rely on the existing `composeBlock` centering and attribution wrap at the photo width for native blocks
- [x] 6.3 Tests: a height-capped native render yields w-wide rows; the composed article centers the photo with equal (±1) left/right margins
- [x] 6.4 Verify: `go vet ./...`, `go build ./...`, `go test ./...` all pass

## 7. Down-snap lands on the caption

- [x] 7.1 Track the image-row/caption split as `articleState.capStart`, reported through `articleImageBlock`/`fitBlock`
- [x] 7.2 Snap downward moves out of the photo's rows onto the caption (`capStart`, or past the block when no attribution is composed); caption rows scroll normally with no snap loop; upward reveal unchanged
- [x] 7.3 Refine the kitty placement-clear check to the photo's rows so a caption-only frame deletes the placement
- [x] 7.4 Tests: snap table with caption/no-caption cases, scroll test landing on and scrolling within the caption, caption-frame placement delete
- [x] 7.5 Verify: `go vet ./...`, `go build ./...`, `go test ./...` all pass
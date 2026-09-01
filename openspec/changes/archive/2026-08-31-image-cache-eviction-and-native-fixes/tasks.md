## 1. Bounded LRU cache infrastructure

- [x] 1.1 Add a package-internal generic LRU cache (`internal/image/lru.go`) with a byte budget, a `sizeOf` function, `Get`/`Set`, and least-recently-used eviction on overflow.
- [x] 1.2 Unit-test the LRU: inserts under budget stay, over-budget inserts evict the least-recently-used, `Get` refreshes recency, accounting reflects `sizeOf`.

## 2. Bound the four image caches

- [x] 2.1 Reimplement `Cache`, `Blocks`, `Photos`, and `Natives` on the LRU with per-instance byte budgets (`Cache` via `Dx()*Dy()*4`).
- [x] 2.2 Keep the existing `Get`/`Set` method signatures so callers are unchanged, and delete the now-unused bespoke map/mutex code.
- [x] 2.3 Add tests: inserting many distinct images keeps total bytes under budget and evicts the oldest entry.

## 3. Decoded source resolution cap

- [x] 3.1 In `BlockCmd`/`PhotoCmd` decode sites, pre-check dimensions via `image.DecodeConfig` and downscale (via `x/image/draw`) to `maxDecodedEdge` before `Cache.Set` and render.
- [x] 3.2 Add tests: an oversized source decodes at the cap; a small source decodes at natural size.

## 4. Native render cache keyed by content width

- [x] 4.1 Change `NativeKey` to `url\x00width` (drop `maxHeight`) and update `NativeCmd`, `composeImageBlock`, and `hasImageSource` call sites.
- [x] 4.2 Add tests: a height-only resize reuses the cached native render (no re-encode); a width change re-renders.

## 5. Delete kitty images by id on article exit

- [x] 5.1 Add a "leaving the article" branch to `nativeImageClear` that emits `image.DeletePlacement(url)` for each prior native image block before falling back to `Clear()`.
- [x] 5.2 Add tests asserting the emitted escape carries `d=i,i=<id>` per image on exit, and that OSC 1337 emits nothing.

## 6. Collapse the height-fit duplication

- [x] 6.1 Extract a shared `fitBudget` helper and make `fitBlock` and `renderNativeFitted` both use it, keeping `LayoutWrapLines` as the single line-count source.
- [x] 6.2 Verify existing fit/snap tests still pass unchanged.

## 7. Single-pass conversion and reused HTTP client

- [x] 7.1 Pass the already-converted body markdown from `newArticleState` into link harvesting, removing the second `convert.ConvertImages` call.
- [x] 7.2 Replace the per-request `http.Client` in `fetch` with a reused package-level client.

## 8. In-flight guard scoped to native render

- [x] 8.1 Clear `imgLoading[url]` only on `NativeMsg`/`FailedMsg` (or terminal block failure), not on the early `BlockMsg`/`PhotoMsg`.
- [x] 8.2 Add a test that a resize mid-render does not start a second native render for the same URL.

## 10. Eviction invisible to rendering

- [x] 10.1 `NativeCmd` re-reads the photo from the store when the Photos cache evicts it, so a native render never fails spuriously on eviction (test: empty Photos cache + store-backed source renders a `NativeMsg`; unavailable bytes still `FailedMsg`).
- [x] 10.2 Key the Blocks cache by content width (`url\x00width`) and carry the render width on `BlockMsg`, so a halfblock placeholder is served for the exact width even when the decoded-image cache evicted the buffer (test: stored block at a different width is not served; same-width block is).
- [x] 10.3 Update specs with the re-derivation and width-keyed placeholder requirements.

## 11. Native render path decode cap and single decode

- [x] 11.1 `NativeRenderer.Render` decodes via the `maxDecodedEdge` cap, so the native path never holds an oversized source full-size (spec requirement applying the cap to both paths).
- [x] 11.2 Split `Render` into decode + `renderDecoded` so the height-fit loop decodes the photo once and re-scales/re-encodes per candidate height instead of re-decoding per iteration.

## 9. Validation

- [x] 9.1 Run `go test ./...` and `go vet ./...` and fix regressions.
- [x] 9.2 Manually confirm (Ghostty) that opening many images keeps process RSS flat and the terminal RSS stops growing across article exits.

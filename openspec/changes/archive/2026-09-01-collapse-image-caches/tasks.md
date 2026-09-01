## 1. Generic map cache

- [x] 1.1 Add a generic `cache[V]` type with `NewCache`, `Get`, and `Set` in `internal/image`
- [x] 1.2 Reimplement `Cache`, `Blocks`, `Photos`, `Natives` as instantiations of `cache[V]` keeping their public method names

## 2. Decoded-image byte counter

- [x] 2.1 Track a running decoded-byte total in the decoded-image cache and expose `TotalBytes()`
- [x] 2.2 Unit-test the counter across `Set` and overwrite

## 3. Native render decode once

- [x] 3.1 Change `renderNativeFitted` to decode the photo once and reuse the decoded image across candidate heights
- [x] 3.2 Add a test asserting a single decode across the fit loop, or verify the existing height-fit scenarios still pass

## 4. Spec reconciliation

- [x] 4.1 Write the `image-cache-eviction` REMOVED delta
- [x] 4.2 Write the `native-image-rendering` MODIFIED delta (render-size keying, viewport-height resize re-renders, single decode, stored-photo fallback)

## 5. Validation

- [x] 5.1 Run `go test ./internal/image/... ./internal/ui/...` and `openspec validate collapse-image-caches`

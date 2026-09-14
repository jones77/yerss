## Why

`Model` (internal/ui/model.go) is a god object: ~40 fields across at least six concerns. The fattest, most self-contained tentacle is the image view — nine fields (`imgLoading`, `imgFrontier`, `nativePending`, `imgSem`, `previewCache`, `nativeSent`, `nativeSentPrev`, `nativeClearPrefix`, `nativeClearArticleID`) and ~30 methods across `image.go` (~950 lines) implementing native-image transmit/clear sequencing, halfblock previews and block composition, and the load frontier/scheduling. Extracting this into one cohesive `imageView` type is the highest-value structural refactor and the prerequisite for a later package split.

## What Changes

- Introduce an `imageView` type in `internal/ui` that owns the nine image-state fields listed above.
- Move the image methods from `image.go` onto `imageView` (native clear/transmit, block composition, load scheduling, and the `image.*Msg` handlers).
- `Model` holds an `imageView` field (e.g. `images imageView`) and delegates image-message handling to it in `Update`; `View` reads the computed native-clear prefix from it.
- Keep behavior byte-identical; the image-load/native correctness invariants (decode-once, O(N) render bounds, snap guards, one clear per transition) are preserved, not relaxed.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/image.go`: methods re-homed onto `imageView`; no logic changes.
- `internal/ui/model.go`: nine fields removed from `Model`, replaced by one `imageView` field; `Update`/`View` delegate.
- `internal/ui/article.go`: composition code that calls the block-composition methods now calls them through the `imageView`.
- `internal/ui/*_test.go`: image tests now construct/assert through `imageView` where they touched `Model` image state directly.

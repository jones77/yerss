## Context

`image.go` carries three image concerns, each with its own state:

1. **Native transmit/clear sequencing** — `nativeSent`, `nativeSentPrev`, `nativeClearPrefix`, `nativeClearArticleID` + `recomputeNativeClear`, `computeNativeClear`, `rewriteNativeReShows`, `suppressClippedNativeTransmits`.
2. **Block composition & previews** — `composeImageBlock`, `fitBlock`, `composeBlock`, `previewClippedImage`, `inlineAttrFor` (reads session image caches).
3. **Load frontier & scheduling** — `imgLoading`, `imgFrontier`, `nativePending`, `imgSem`, `previewCache` + `fireImageLoad`, `loadImageCmd`, `ensureImageSource`, `advanceFrontier`, `advanceFrontierLoads`, `fireImagesTo`, `photoBlockInView`, `gateCmd`, `hasImageSource`.

Plus the message handlers (`onBlockLoaded`, `onPhotoLoaded`, `onNativeLoaded`, `onImageFailed`), `recomposeArticle`, `flushRecompose`, and `scrollArticle` that tie the three together. The anchor helpers are being moved out by `extract-link-anchor-helpers` first, so this change starts from a clean `image.go`.

The correctness invariants in `AGENTS.md` (decode-once, at-most-O(N) synchronous renders, one clear per no-native-article transition, snap guards that never bounce back) all live in these methods and must be preserved unchanged.

## Goals / Non-Goals

**Goals:** one `imageView` type owns all image state and methods; `Model` shrinks by nine fields and ~30 methods; behavior byte-identical.

**Non-Goals:** no package split (that is `split-ui-package`, later); no sub-division of `imageView` into `nativeView`/`loader` yet; no logic or invariant changes.

## Decisions

### D1 — `imageView` with a `*Model` back-reference

`imageView` holds a `m *Model` field so the ~30 methods can move mechanically (call sites inside the methods keep working through `v.m`). This is the same "dependency inversion first, logic extraction later" pattern the `app.Session` change used: get the state off `Model` now, replace the back-reference with explicit dependencies in the package split.

### D2 — `Model` keeps `images imageView` (value, not pointer)

`imageView` is a value field on `Model`, matching how `listState`/`articleState`/`popupState` are already embedded by value. Construction stays in `New` (maps and channel initialized there).

### D3 — Delegate in `Update`, keep `View` pure

`Update` routes `image.BlockMsg`/`PhotoMsg`/`NativeMsg`/`FailedMsg` to `v.images.on*` methods, which set `v.recomposePending`/`v.recomposeCount` exactly as today. `View` continues to read the precomputed `nativeClearPrefix` (now via `v.images`) so `View` stays pure. `recomputeNativeClear` stays a `Model` method that calls into the moved compute logic, preserving the `defer m.recomputeNativeClear()` contract at the top of `Update`.

### D4 — Move in phases, compile-green after each

Phase order mirrors the dependency graph: (1) native sequencing, (2) block composition, (3) load scheduling, (4) message handlers + recompose. `go build ./...` after each phase so the diff stays reviewable and any missed call site surfaces immediately.

## Risks / Trade-offs

- [Large diff, high correctness sensitivity] → Phased moves + build/test gate after each phase; the image test suite (`image_render_count_test.go`, `frame_cache_test.go`, `article_image_test.go`, `render_safety_test.go`) is the regression net.
- [`imageView.m` back-reference is a smell] → Acceptable interim; `split-ui-package` replaces it with explicit dependencies and/or a package boundary.
- [Embedded value field vs pointer] → Value keeps zero-value construction simple and mirrors existing state structs; revisit only if `imageView` later needs sharing.

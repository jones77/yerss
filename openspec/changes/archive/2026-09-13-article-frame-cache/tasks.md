## 1. Content and native generation counters

- [x] 1.1 Add a `contentGen int` field to `Model` (model.go) and bump it at the top of `newArticleState` (article.go) so every article-state construction — open, recompose, resize, selection rebuild — invalidates the frame cache. Initialize to 0 in `New`.
- [x] 1.2 Add a `nativeGen int` field to `Model` and bump it in `recomputeNativeClear` (image.go) when the native transmit state the frame body renders from actually changes (the frame's `nativeSent` snapshot differs from the previous frame's, or `nativeClearArticleID` changed), so steady-state boundary scrolls leave it unchanged.

## 2. Memoize the article frame

- [x] 2.1 Add the frame-cache fields to `Model` (`frameKey articleFrameKey`, `frameValid bool`, `frameStr string`) and define the comparable `articleFrameKey` struct (articleID, contentGen, nativeGen, yOffset, width, height, sel, ascii, linkExpand).
- [x] 2.2 Extract the current `renderArticle` body into `renderArticleUncached` and add a memoizing `renderArticle` wrapper that builds the key, returns the cached `frameStr` on match, and otherwise renders, stores `(key, frame)`, and sets `frameValid`. `View()` keeps calling `renderArticle` and keeps prepending `nativeClearPrefix` outside the cached body.
- [x] 2.3 Confirm `renderArticle`'s full read set is covered by the key (walk `rewriteNativeReShows`, `suppressClippedNativeTransmits`, `highlightSelection`, and `RenderArticleBorder`); add `contentGen`/`nativeGen` wiring where any dependency is missing.

## 3. Regression tests

- [x] 3.1 Add a render-counting shim (mirroring the existing `ImgRenderer` counting-shim in `article_image_test.go`) that counts article-frame renders, and a test asserting that a wheel-up scroll at the top and a wheel-down scroll at the bottom each cause zero additional renders, while a real scroll causes exactly one.
- [x] 3.2 Add a byte-identity test: for a fixture with native images (and a selection + link/ascii toggle), assert the cached frame equals a freshly rendered uncached frame.
- [x] 3.3 Add a test that an image-load recompose at a boundary scroll position re-renders (contentGen bump), i.e. the cache does not serve stale content.

## 4. Verification

- [x] 4.1 Run `go test ./...` and the linter; confirm no regressions in scroll/snap, native-clear, selection, or read-state tests.
- [x] 4.2 Manual check against a stored image-heavy article in the local SQLite DB: scroll to the top, flick the wheel up repeatedly, then scroll down and confirm the app responds immediately (no multi-second backlog).

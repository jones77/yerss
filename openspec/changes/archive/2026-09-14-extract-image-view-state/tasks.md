## 1. Introduce the type and move state

- [x] 1.1 Define `imageView` in `internal/ui` with a `m *Model` back-reference and the nine fields (`loading`, `frontier`, `nativePending`, `sem`, `previewCache`, `nativeSent`, `nativeSentPrev`, `nativeClearPrefix`, `nativeClearArticleID`)
- [x] 1.2 Replace the nine fields on `Model` with `images imageView` and initialize its maps/channel in `New`
- [x] 1.3 Run `go build ./...` to find every now-dangling reference

## 2. Move native transmit/clear sequencing

- [x] 2.1 Move `recomputeNativeClear`, `computeNativeClear`, `rewriteNativeReShows`, `suppressClippedNativeTransmits`, `nativeSentEqual` onto `imageView`; adapt call sites
- [x] 2.2 Keep `Model.recomputeNativeClear` as a thin delegate so `Update`'s `defer` contract is unchanged
- [x] 2.3 Run `go build ./...`

## 3. Move block composition & previews

- [x] 3.1 Move `composeImageBlock`, `fitBlock`, `composeBlock`, `previewClippedImage`, `inlineAttrFor` onto `imageView`
- [x] 3.2 Update `article.go` composition (`spliceBody`) to call them through `m.images`
- [x] 3.3 Run `go build ./...`

## 4. Move load frontier & scheduling

- [x] 4.1 Move `fireImageLoad`, `loadImageCmd`, `ensureImageSource`, `advanceFrontier`, `advanceFrontierLoads`, `fireImagesTo`, `photoBlockInView`, `gateCmd`, `hasImageSource`, `imagesEnabled`, `nativeImages` onto `imageView`
- [x] 4.2 Run `go build ./...`

## 5. Move message handlers & recompose

- [x] 5.1 Move `onBlockLoaded`, `onPhotoLoaded`, `onNativeLoaded`, `onImageFailed`, `recomposeArticle`, `flushRecompose`, `toggleLinkExpand`, `scrollArticle` onto `imageView`
- [x] 5.2 Update `Model.Update` to delegate the four `image.*Msg` cases to `m.images` and still flush recompose through the model
- [x] 5.3 Run `go build ./...`

## 6. Validation

- [x] 6.1 Run `go test ./...` (image, frame-cache, article-image, render-safety suites green)
- [x] 6.2 Manually open a photo-heavy article and confirm decode-once / render-count bounds hold (per `AGENTS.md` reproduction steps)

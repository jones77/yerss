## 1. convert: inline image discovery

- [ ] 1.1 Add a `newMarkedConverter()` in `convert/convert.go` whose `img` renderer writes `"\x00img:" + src + "\x00"` instead of the `[alt]`/`[image]` placeholder
- [ ] 1.2 Add `ConvertImages(html string) (string, []string)` that runs the marked converter and returns the marked markdown plus the ordered URL list (scanned from the sentinels); keep `Convert` unchanged
- [ ] 1.3 convert tests: `ConvertImages` returns sentinels in document order and the correct URL list; `Convert` still emits `[alt]`/`[image]` placeholders (no regression)

## 2. articleState: ordered image blocks

- [ ] 2.1 In `internal/ui/article.go`, add an `imageBlock{url string; imgStart, capStart, imgEnd int; nativeImg bool}` type and an `imageBlocks []imageBlock` field on `articleState` (lead = index 0, then inline in order); keep scalar `imgStart`/`imgEnd`/`capStart`/`nativeImg` for the lead image or migrate callers/tests to the list, choosing the least-churn path
- [ ] 2.2 In `newArticleState`, build the ordered `imageBlocks` list as blocks are composed

## 3. Interleaving inline image blocks

- [ ] 3.1 In `newArticleState`, call `convert.ConvertImages(a.Content)`, split the body markdown on the `\x00img:<url>\x00` sentinel into alternating text/image segments, render text segments via `renderMarkdown`, and compose an image block for each image segment (generalize `articleImageBlock`/`fitBlock`/`composeBlock` to take a URL + attribution instead of reading `a.ImageURL`)
- [ ] 3.2 Compute inline attribution as the image's alt text, else `"photo: " + sourceID(a.Link, a.FeedURL)`
- [ ] 3.3 While an inline image is still loading, compose a blank placeholder line so the text flow is stable and the block inserts at the right offset on load
- [ ] 3.4 Record each inline block's `[imgStart, capStart, imgEnd]` in `imageBlocks`

## 4. Multi-image async load

- [ ] 4.1 Change `fireImageLoad`/`loadImageCmd` to fire `BlockCmd`/`PhotoCmd`/`NativeCmd` for each image URL (lead + inline) with one in-flight guard per URL
- [ ] 4.2 Update `onBlockLoaded`/`onPhotoLoaded`/`onNativeLoaded`/`onImageFailed` to match on URL (not only `a.ImageURL`) and re-compose
- [ ] 4.3 Persist inline images at positions 1..N via the unified `SetArticleImage` (reuse the image-package persistence added in `unified-image-storage`)

## 5. Scroll: generalized snap + edge-snap

- [ ] 5.1 Change `snapYOffset` to accept `[]imageBlock` and apply the per-block rules (down into photo → caption, down fully-visible → caption, up into photo → reveal, up fully-visible native → 0, caption normal) across the ordered list
- [ ] 5.2 Add edge-snap: down move bringing an inline image into view from below snaps to its top; up move bringing one into view from above snaps its bottom to the viewport bottom
- [ ] 5.3 Update `scrollArticle` to pass `m.article.imageBlocks` and `m.article.nativeImg`
- [ ] 5.4 UI tests: interleaved block composition + attribution; late inline load inserts at the right offset; `snapYOffset` with a block list handles a mid-document block (down/up into photo, edge-snap on entry)

## 6. Verification

- [ ] 6.1 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [ ] 6.2 Run `openspec validate inline-image-rendering`
- [ ] 6.3 Manual smoke test: an article with inline images renders them in place; scrolling snaps each into view then past; halfblock and native terminals both behave

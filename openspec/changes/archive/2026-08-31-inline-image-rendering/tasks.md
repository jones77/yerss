## 1. convert: inline image discovery

- [x] 1.1 Add a `newMarkedConverter()` in `convert/convert.go` whose `img` renderer writes `"\x00img:" + src + "\x00"` instead of the `[alt]`/`[image]` placeholder
- [x] 1.2 Add `ConvertImages(html string) (string, []string)` that runs the marked converter and returns the marked markdown plus the ordered URL list (scanned from the sentinels); keep `Convert` unchanged
- [x] 1.3 convert tests: `ConvertImages` returns sentinels in document order and the correct URL list; `Convert` still emits `[alt]`/`[image]` placeholders (no regression)

## 2. articleState: ordered image blocks

- [x] 2.1 In `internal/ui/article.go`, add an `imageBlock{url string; imgStart, capStart, imgEnd int; nativeImg bool}` type and an `imageBlocks []imageBlock` field on `articleState` (lead = index 0, then inline in order); keep scalar `imgStart`/`imgEnd`/`capStart`/`nativeImg` for the lead image or migrate callers/tests to the list, choosing the least-churn path
- [x] 2.2 In `newArticleState`, build the ordered `imageBlocks` list as blocks are composed

## 3. Interleaving inline image blocks

- [x] 3.1 In `newArticleState`, call `convert.ConvertImages(a.Content)`, split the body markdown on the `\x00img:<url>\x00` sentinel into alternating text/image segments, render text segments via `renderMarkdown`, and compose an image block for each image segment (generalize `articleImageBlock`/`fitBlock`/`composeBlock` to take a URL + attribution instead of reading `a.ImageURL`)
- [x] 3.2 Compute inline attribution as the image's alt text, else `"photo: " + sourceID(a.Link, a.FeedURL)`
- [x] 3.3 While an inline image is still loading, compose a blank placeholder line so the text flow is stable and the block inserts at the right offset on load
- [x] 3.4 Record each inline block's `[imgStart, capStart, imgEnd]` in `imageBlocks`

## 4. Multi-image async load

- [x] 4.1 Change `fireImageLoad`/`loadImageCmd` to fire `BlockCmd`/`PhotoCmd`/`NativeCmd` for each image URL (lead + inline) with one in-flight guard per URL
- [x] 4.2 Update `onBlockLoaded`/`onPhotoLoaded`/`onNativeLoaded`/`onImageFailed` to match on URL (not only `a.ImageURL`) and re-compose
- [x] 4.3 Persist inline images at positions 1..N via the unified `SetArticleImage` (reuse the image-package persistence added in `unified-image-storage`)

## 5. Scroll: generalized snap + edge-snap

- [x] 5.1 Change `snapYOffset` to accept `[]imageBlock` and apply the per-block rules (down into photo → caption, down fully-visible → caption, up into photo → reveal, up fully-visible native → 0, caption normal) across the ordered list
- [x] 5.2 Add edge-snap: down move bringing an inline image into view from below snaps to its top; up move bringing one into view from above snaps its bottom to the viewport bottom
- [x] 5.3 Update `scrollArticle` to pass `m.article.imageBlocks` and `m.article.nativeImg`
- [x] 5.4 UI tests: interleaved block composition + attribution; late inline load inserts at the right offset; `snapYOffset` with a block list handles a mid-document block (down/up into photo, edge-snap on entry)

## 6. Verification

- [x] 6.1 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [x] 6.2 Run `openspec validate inline-image-rendering`
- [x] 6.3 Manual smoke test: an article with inline images renders them in place; scrolling snaps each into view then past; halfblock and native terminals both behave

## 7. Smoke-test fixes (Intercept article "Ecocide in Iran")

The 6.3 smoke test surfaced three defects on an article with five images (a lead
image, three related-post thumbnails, a book-cover figure, and a TomDispatch
promo logo): a caption leak, stray link fragments, and broken two-motion scroll.

- [x] 7.1 Attribution leak: make `ImageCredit` figure-local for every image (document-global first-figure/credit fallbacks removed entirely) so no image borrows another's caption — the book-cover caption no longer appears under the lead, the related-post thumbnails, or the TomDispatch logo; update `convert` and `internal/ui` tests
- [x] 7.2 Link-aware sentinel handling: `inlineBodyParts` and `stripDuplicateSentinels` consume the wrapping `[...](...)` markdown link around a sentinel so no stray `[`/`](...)` renders (observed as a literal `[` before the logo); add a `convert` test for `<a><img></a>` markup
- [x] 7.3 Two-motion inline snap: `snapYOffset` snaps a down move entering a photo from above to `imgStart` (motion 1), a down move starting at/inside the photo to `capStart` (motion 2), and limits the fully-visible skip to the lead block; update `TestScrollSnapsThroughInlineImage` and the snap table tests
- [x] 7.4 Recompose snap: after a load re-composes the article, an offset landing inside a photo range or with an image top mid-window snaps to the image's top; update `recomposeArticle` and tests
- [x] 7.5 Re-run the 6.3 smoke test on the Intercept article: each inline image snaps into view on the first down press and skips to its caption on the second; no stray brackets; captions appear only under their own image

## 8. Second smoke-test round (border, paging, promo captions)

The 7.5 smoke test found the native photo placements drawing over the article
border, paging (space) snapping to each image and skipping the text between,
and the TomDispatch promo banner rendering a leftover link tail and a
`photo: <source>` caption. Decisions: single-line moves snap in two stages
(bottom then top down, top then bottom up); multi-line moves land naturally
even mid-photo; linked promo images use their link text as the caption;
relative link targets resolve to absolute.

- [x] 7.6 Native border fix: `nativeImageClear` deletes placements not fully contained in the viewport window and `suppressClippedNativeTransmits` blanks the transmit escape on a partially visible image's first line, so an image never paints over the article border; tests added
- [x] 7.7 Multi-line moves don't snap: `scrollArticle` takes a snap flag (line moves snap, page/half-page/goto-bottom land naturally even mid-photo so paging never skips the text between images); tests added
- [x] 7.8 Two-stage single-line snap: a down move snaps an inline image bottom-aligned to the viewport bottom then top-aligned, then out; up mirrors it; `TestSnapYOffsetBlockList` and flow tests updated
- [x] 7.9 Link-text sentinel + caption: `nextSentinel` consumes a wrapped link with text between the sentinel and the destination (promo banners), reports the cleaned link text, and `inlineAttribution` uses it as a caption fallback (alt → figure credit → link text → source); convert + UI tests
- [x] 7.10 Relative href resolution: `resolveLink` resolves relative destinations like `/tomdispatch` against the article's origin for the links popup and open-url; tests added

## 9. Final smoke test (user)

- [x] 9.1 Re-run on the Intercept article: images never overlap the border; space pages through text without skipping to images; j/k snap each inline image bottom-then-top and out; the TomDispatch logo caption reads its banner text with no leftover `](/tomdispatch)`; the links popup lists `https://theintercept.com/tomdispatch`; 6.3 is then complete

## 10. Third smoke-test round (blank space between images)

The 9.1 smoke test confirmed the border fix, but scrolling left the photos
sitting partially visible between the images: the skip of one image leaves the
next image's top already inside the window (consecutive tall images are spaced
closer than the viewport height), so the entry-from-below bottom-snap never
fired, the placement was deleted, and the user scrolled through a blank region
("a lot of negative space at the bottom") for many presses.

- [x] 10.1 Down bottom-snap keyed on partial visibility: `snapYOffset` bottom-snaps any inline image partially visible in the window (top inside, bottom below the fold), covering both entry from below and the skip-leaves-next-partial case, so the next down press aligns it to the viewport bottom instead of leaving it blank; `TestSnapYOffsetSkipLeavesNextImagePartial` and `TestScrollDownNeverLeavesInlineImageBlank` added
- [x] 10.2 Up mirror: an up move that cuts an inline image's bottom below the fold snaps it fully below the fold on the next move so it scrolls off the bottom cleanly instead of rendering a blank strip; `TestScrollUpNeverLeavesInlineImageBlank` added
- [x] 10.3 Re-run on the Intercept article: j-scrolling through the inline images shows each photo fully (no blank strip between images, no manual scrolling through empty rows); k-scrolling back up leaves each photo until it scrolls off in one snap

## 11. Fourth smoke-test round (blank strip and caption consistency)

The 10.3 smoke test found the bottom snap hiding a one-line caption entirely
(`end - vpH` put the last caption line exactly at the fold, showing only the
beginning of a wrapped caption for others) and a transient empty strip where a
partially visible native image pokes into the window between close images.

- [x] 11.1 Bottom snap keeps the caption on screen: the bottom-aligned target is `imgEnd - vpH + 1` (down bottom snap, down top-snap gate, and up bottom snap), so a one-line caption is fully visible at the bottom snap and a wrapped one is not cut; snap table/flow tests updated
- [x] 11.2 No blank strip for a poking-in native image: `previewClippedImage` renders the cached halfblock (at the native block's row count, so it aligns) into the visible rows of a partially visible native image, so a photo poking into the window shows a blocky preview instead of an empty region; `TestClippedNativeImageShowsHalfblockPreview` added
- [x] 11.3 Up scroll-off reveals the previous image when its target would land inside the previous image's photo range (images spaced closer than the viewport), so the up transition never leaves the previous image partially visible
- [x] 11.4 Re-run on the Intercept article: the bottom snap shows each photo with its caption at the bottom edge (a one-line caption included); between the Trump's War thumbnail and the book cover the two Fallujah paragraphs read normally and the book cover's top shows a blocky preview rather than an empty strip before it snaps in

## 12. Fifth smoke-test round (scroll-k loop)

The 11.4 smoke test on the hurricane article
(`https://theintercept.com/2026/08/29/hurricane-helene-ice-asheville-nc-g20-trump/`)
found scroll-k stuck after a photo snapped to the top of the screen. Its two
gallery photos share a long caption that overflows the reserved fit budget
(`inlineAttrFor` passes no link text, so the native fit reserves too little for
the caption), making each block fill the viewport exactly: the bottom-aligned
position equals the top-aligned one, so the up bottom-snap returned the
pre-move offset and re-fired forever.

- [x] 12.1 Snap loop guard: every `snapYOffset` result is rejected when it equals the pre-move offset (which would re-fire identically and loop), so unchanged offsets fall through to normal scrolling; `TestSnapYOffsetFullViewportBlockDoesNotLoop` added
- [x] 12.2 Full-viewport stage collapse: an up move from a block's top-aligned position whose bottom-aligned position is at or past the offset just reached (the block fills the viewport, so there is no distinct bottom stage) scrolls the image fully off, revealing the previous image when the target lands in its range; `TestScrollUpThroughFullViewportImageDoesNotLoop` added
- [x] 12.3 Re-run on the hurricane article from the bottom: scroll-k shows each gallery photo once (filling the screen) and then scrolls it off to the previous image or content, with no stuck offset; scroll-j through the whole article stays clean

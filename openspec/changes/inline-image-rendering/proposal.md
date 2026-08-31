## Why

The reader only ever shows the publisher-attached lead image (an enclosure or
media:thumbnail), always pinned below the header. Inline `<img>` elements in the
article body are discovered but discarded — `convert.renderImage` writes an
`[alt]`/`[image]` placeholder and throws away the `src`. Articles that carry
their illustrations in the body (a common pattern for technical posts and news)
read as text with `[image]` stubs. Inline images should render in place, scroll
with the content, and behave like the lead image: a real photo on native
terminals, halfblock art otherwise, with a centered attribution, persisted so a
later open re-renders them offline.

## What Changes

- `convert` gains a marked-conversion path that emits a per-image sentinel token
  in the markdown instead of the `[alt]`/`[image]` placeholder, and returns the
  ordered list of inline image URLs. The existing `Convert` API is unchanged.
- The article view interleaves inline image blocks into the rendered body at the
  discovered positions, above/below the surrounding text, each with its own
  attribution and its own scroll identity.
- Inline images are fetched asynchronously like the lead image, persisted to the
  unified `article_images` table (positions 1..N), and re-rendered offline.
- Scroll snap is generalized from a single lead-image block to an ordered list of
  image blocks, and gains edge-snap: when scrolling brings an inline image into
  view it snaps to the viewport edge, and the next same-direction scroll takes it
  out.

## Capabilities

### New Capabilities

- `inline-image-rendering`: discovers inline `<img>` URLs in article content,
  renders each as an interleaved block with attribution, persists each, and
  snap-scrolls across them.

### Modified Capabilities

- `reader-ui`: "Atomic image scroll behavior" generalizes from a single lead
  image to every image block (lead and inline), and adds edge-snap for images
  scrolled into view.

## Impact

- `convert/convert.go`: new `ConvertImages(html) (string, []string)` using a
  second converter whose image renderer emits sentinels; `Convert` stays
  placeholder-based.
- `internal/ui/article.go`: `articleState` replaces scalar `imgStart/imgEnd/
  capStart/nativeImg` with an ordered `[]imageBlock`; `newArticleState`
  interleaves inline blocks into the body.
- `internal/ui/image.go`: `snapYOffset` and `scrollArticle` operate on the block
  list; the load path (`fireImageLoad`/`onBlockLoaded`/`onPhotoLoaded`/
  `onNativeLoaded`) keys images by URL+position instead of lead URL only.
- `internal/store`: reuse `article_images` (positions 1..N) — no schema change.
- Tests across `convert`, `internal/ui`, and `internal/image`.

**Sequencing note (required):** this change builds on `unified-image-storage`
(the `article_images` table) and on `photo-scroll-skip` (the single-image scroll
snap it generalizes). Archive `async-native-image-render`, `photo-scroll-skip`,
then `unified-image-storage` before this change, so its `reader-ui` delta applies
against the post-archive spec tree.

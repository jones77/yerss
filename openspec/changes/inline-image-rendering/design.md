## Open Questions

Resolved for this change. Inline image attribution uses the image's alt text, or
`photo: <source>` fallback (full per-image `figcaption` association is a
follow-up). The exact edge-snap alignment is a first implementation to be tuned
in the manual smoke test.

## Options

- Marked-conversion with sentinel tokens (chosen) vs changing `Convert`'s return
  type to a segment list. Chosen sentinels: `Convert` keeps its signature, the
  article view opts into the marked path, and all other callers are untouched.
- Second converter instance (chosen) vs per-call registry on the shared
  converter. Chosen a second instance: stateless, no concurrency concerns.

## Problems

- Inline images have no natural "header below" anchor; each must carry its own
  `[imgStart, capStart, imgEnd]` triple and the scroll snap must find the block
  relevant to a given offset.
- The load flow is keyed by the single lead `a.ImageURL`; it must become
  multi-image, keyed by URL + position, with one in-flight guard per image.
- A late-loading inline image inserts lines mid-document; the existing
  offset-bump-on-insert must generalize to the insertion point of each image.

## Remaining Decisions

Per-image `figcaption` attribution is deferred. Edge-snap alignment may need
smoke-test tuning.

## Details

### 1. convert: inline image discovery

Add a second converter and a public API (keeping `Convert` unchanged):

```go
// ConvertImages converts html to markdown where each inline <img> is replaced
// by a sentinel token "\x00img:<url>\x00" (in document order), and returns the
// ordered list of image URLs.
func ConvertImages(html string) (string, []string)
```

`newMarkedConverter()` is identical to `newConverter()` except its `img`
renderer writes `"\x00img:" + src + "\x00"` instead of the `[alt]`/`[image]`
placeholder. URLs cannot contain NUL, so the sentinel is unambiguous. The URL
list is derived by scanning the output in order. `Convert` continues to use the
placeholder renderer, so list view, JSON dump, and ASCII placeholders are
unchanged.

### 2. articleState: ordered image blocks

Replace the scalar fields with a slice:

```go
type imageBlock struct {
    url                       string
    imgStart, capStart, imgEnd int
    nativeImg                 bool
}
// articleState
imageBlocks []imageBlock   // lead (index 0) then inline, in document order
```

Keep the existing single-image fields as derived helpers for the lead image (or
migrate call sites to `imageBlocks`). Tests reference `imgStart`/`capStart`/
`imgEnd` heavily — add accessors `leadBlock()` / `blockAt(i)` and update tests,
or keep the scalar fields for the lead image plus a slice for inline; the
implementer chooses the least-churn path, but the scroll snap must see the full
ordered list.

### 3. newArticleState interleaving

1. `markdown, urls := convert.ConvertImages(a.Content)`.
2. Split `markdown` on the `\x00img:<url>\x00` sentinel into alternating
   text/image segments.
3. Compose the lead image block below the header exactly as today.
4. For each segment: render text via `renderMarkdown`; for each image segment,
   compose an image block using the existing `articleImageBlock`/`fitBlock`/
   `composeBlock` machinery (generalized to take a URL + attribution instead of
   reading `a.ImageURL`), or a blank line while the image is still loading.
5. Record each image block's `[imgStart, capStart, imgEnd]` in `imageBlocks`,
   in document order (lead = index 0).

Attribution for an inline image: `image.alt` text if non-empty, else
`"photo: " + sourceID(a.Link, a.FeedURL)`.

### 4. Multi-image async load

`fireImageLoad`/`loadImageCmd` fire `BlockCmd`/`PhotoCmd`/`NativeCmd` for each
image URL (lead + inline) rather than only `a.ImageURL`; the in-flight guard
`imgLoading` keys by URL. `onBlockLoaded`/`onPhotoLoaded`/`onNativeLoaded`/
`onImageFailed` match on URL and re-compose; `recomposeArticle` preserves the
reading position (the existing inserted-lines bump already keys off the
re-composed line delta, so it generalizes as long as late images insert their
block at the right position).

### 5. Scroll: generalized snap + edge-snap

`snapYOffset(yOffset, vpH, blocks []imageBlock, direction int, native bool)`
applies, for the block(s) whose range brackets or adjoins the move:

- Down landing in a photo range `[start, capStart)` → `capStart`.
- Down with a photo fully visible (`yOffset <= start && yOffset+vpH >= end`) →
  `capStart`.
- Up landing in a photo range → `start` (reveal).
- Up with a photo fully visible and `yOffset > 0` (native) → `0`.
- Caption rows scroll normally both ways.

Edge-snap (new, inline blocks):

- Down: when the move brings an inline image into view from below — the image's
  top `start` is within or below the viewport top and the image is not yet fully
  visible — snap `yOffset` to `start` (image top-aligned). The next down move
  then falls into the "fully visible → `capStart`" rule and skips it.
- Up: the mirror — when the move brings an inline image into view from above,
  snap `yOffset` so the image bottom aligns with the viewport bottom
  (`yOffset = end - vpH + 1`, clamped), then the next up move reveals it.

The lead image (index 0) is at the top of the article and never "scrolls into
view from below", so the edge-snap naturally only fires for inline images. This
alignment is the part most likely to need tuning on a real terminal; keep it
isolated in `snapYOffset` so the smoke test can adjust it without touching
composition.

### Tests

- `convert`: `ConvertImages` returns sentinels in document order + the URL list;
  `Convert` still emits `[alt]`/`[image]` (no regression).
- `internal/ui`: an article with an inline `<img>` composes an interleaved block
  with attribution; a late inline load inserts its block at the right offset;
  `snapYOffset` with a block list handles a mid-document block (down into the
  photo → caption, up into the photo → reveal, edge-snap on entry).
- `internal/image`: reuse of the existing commands for a second URL (no API
  change beyond persistence at positions 1..N via `SetArticleImage`).

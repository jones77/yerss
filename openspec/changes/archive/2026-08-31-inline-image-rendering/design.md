## Open Questions

Resolved for this change. Inline image attribution is per-image: alt text, else
the caption/credit of the image's own figure (a src-matched `<figure>`), else
the text of any markdown link wrapping the image (promo banners label
themselves there), else `photo: <source>`; there is no document-global credit
fallback for any image (the smoke test showed the first figure's caption
attaching to every image in the article — the lead, the related-post
thumbnails, and the logo). Single-line scrolls snap inline images in two
stages: down brings one in bottom-aligned to the viewport bottom, then
top-aligned, then skips it onto its caption; up mirrors it. Multi-line moves
(page, half-page, goto-bottom) do not snap at all — they land wherever they
land, even mid-photo, so paging never skips the text between images. The lead
keeps the one-press skip. Linked `<img>` sentinels consume their whole wrapping
markdown link (including any link text between the sentinel and the
destination) so no `[`/`](...)` fragments render, and relative link targets are
resolved to absolute against the article's origin. Native photo placements are
drawn only when fully contained in the viewport window, so a partially visible
image never paints over the article border.

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

Stable load placeholders — reserving the final image box from the HTML
`width`/`height` attributes so the document never shifts while images load —
are deferred to a follow-up; this change mitigates load-time scroll chaos with
a recompose snap instead. The smoke test used an Intercept article ("Ecocide
in Iran") whose five images include three related-post thumbnails, a
book-cover figure, and a TomDispatch promo banner.

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

Attribution for an inline image: `image.alt` text if non-empty, else the
`figcaption`/credit of the figure whose `<img>` src matches the URL, else the
text of any markdown link that wraps the image (promo banners render as a
linked image whose link text labels them, e.g. "Read Our Complete Coverage
TomDispatch"), else `"photo: " + sourceID(a.Link, a.FeedURL)`. There is no
document-global first-figure or first-credit fallback in `ImageCredit` for any
image: applied to every image, that fallback attached the article's first
figure caption to unrelated images (observed on the Intercept smoke-test
article, where the book-cover caption appeared under the lead image, three
related-post thumbnails, and the TomDispatch promo logo). A sentinel rendered
inside a markdown link (`<a><img>…text…</a>` emits `[SENTINEL text](href)`)
must consume the whole link construct — the `[`, any link text after the
sentinel, and the `](destination)` — when splitting, so no stray `[`, link
text, or `](...)` renders around the image (observed as a literal `[` before
the TomDispatch logo); the consumed link text is the caption fallback above.
Relative link destinations (e.g. `/tomdispatch`) resolve to absolute against
the article's origin when harvesting article links.

### 4. Multi-image async load

`fireImageLoad`/`loadImageCmd` fire `BlockCmd`/`PhotoCmd`/`NativeCmd` for each
image URL (lead + inline) rather than only `a.ImageURL`; the in-flight guard
`imgLoading` keys by URL. `onBlockLoaded`/`onPhotoLoaded`/`onNativeLoaded`/
`onImageFailed` match on URL and re-compose; `recomposeArticle` preserves the
reading position (the existing inserted-lines bump already keys off the
re-composed line delta, so it generalizes as long as late images insert their
block at the right position).

### 5. Scroll: two-stage single-line snap, multi-line lands as-is

`scrollArticle(fn, snap)` runs the viewport move and, only for single-line moves
(snap true, the line keys and the mouse wheel), applies `snapYOffset` across the
ordered block list. Multi-line moves (page, half-page, goto-bottom) pass snap
false and land wherever they land, even mid-photo, so paging never skips the
text between images (the first smoke test's space-bar complaint).

`snapYOffset(yOffset, vpH, before, blocks []imageBlock, direction int, native bool)`
applies `before`, the offset before the move, to gate each stage so the snaps
cannot loop. Every snap is guarded to never return the pre-move offset — a
snap that bounces back to `before` re-fires identically on the next move and
loops. Inline images snap in two stages:

- Down, bottom snap: an inline image partially visible in the window — its top
  inside the window and at-or-below the viewport top, its bottom below the fold
  (`start >= yOffset && start < yOffset+vpH && capStart > yOffset+vpH`) → the
  block's last line aligns to the viewport bottom (`end - vpH + 1`), fully
  visible with its caption at the bottom edge. This fires both when a down move
  first brings an image's top into view from below and when a skip of the
  preceding image leaves the next image's top already inside the window
  (consecutive tall images are spaced closer than the viewport height). A fully
  visible image or one entirely below the fold does not fire it, and the target
  leaves the image fully contained, so it cannot loop. The `+ 1` (not `end -
  vpH`) keeps the last caption line on screen: `end - vpH` put it exactly at the
  fold, hiding a one-line caption entirely at the bottom snap and showing only
  the beginning of a wrapped one — the fourth smoke test's "sometimes we get the
  beginning of a caption when snapped at the bottom, sometimes we don't".
- Down, top snap: from the bottom-aligned position (`before == end-vpH+1`, still
  above the image) the next down move → `start`.
- Down landing in a photo range: entered from above (`before < start`) → `start`
  (the top stage); starting at or inside it → `capStart` (skip out). The skip
  intentionally lands on the preceding image's caption even when the next image
  pokes into the window there: chaining straight to the next image would skip
  the article text between (the Intercept article has two paragraphs between the
  related-post thumbnail and the book cover figure), and the poke-in frame
  renders a halfblock preview rather than a blank strip.
- Up mirrors: an image entering from above (`end > yOffset && end <= before`)
  → `start` (top-aligned); from the top-aligned position (`before == start`)
  the next up move → `end - vpH + 1` (bottom-aligned); a move landing in a photo
  range → `start` (reveal). Once an up move cuts the image's bottom below the
  fold (the same partial condition as the down bottom snap), the next up move
  snaps it fully below the fold (`start - vpH`) so it scrolls off the bottom
  cleanly instead of rendering a blank strip; the target puts its top at the
  fold, so it cannot loop. When that target would land inside the previous
  inline image's photo range (images spaced closer than the viewport), the
  previous image is revealed instead. A block whose image plus wrapped caption
  fills the viewport exactly (a long shared gallery caption overflows the
  reserved fit budget) has its bottom-aligned position equal to its top-aligned
  one, so there is no distinct bottom stage: the up move from the top-aligned
  position scrolls the image fully off (revealing the previous image) instead
  of snapping back onto itself — the fifth smoke test's "doesn't snap to the
  bottom on the next scroll-k" was this loop on the hurricane article's two
  gallery photos.
- Down with the **lead** photo fully visible (`yOffset <= start && yOffset+vpH
  >= end`) → `capStart`, and a down move landing in the lead's photo → `capStart`
  (the one-press skip). The fully-visible skip is limited to the lead block:
  applied to inline blocks it preempts the two-stage snap and skips images that
  were never shown (observed in the smoke test).
- Up with the lead photo fully visible and `yOffset > 0` (native) → `0`. The
  jump-to-top is limited to the lead block (index 0): it exists because
  re-transmitting the top photo per header step is costly, and the article top
  is directly above it; inline images must scroll normally, so a fully-visible
  inline photo never jumps to the article top.
- Caption rows scroll normally both ways.

The lead image (index 0) is at the top of the article and never "scrolls into
view from below", so the bottom/top snaps naturally only fire for inline images.
The lead-exclusion keys off index 0, which is correct as long as a lead block is
composed; an article whose lead image fails to compose or has no `image_url`
falls back to blocks[0] being the first inline image, so the rules key on
`imageBlock` (a `lead` flag or URL match) rather than hardcoding index semantics
where practical.

Recompose snap: after an image load re-composes the article, the offset must
not be left inside a photo range or with an image's top materialized mid-window
(loads land asynchronously; while an image loads its 1-line placeholder means no
block is registered, so the offset can land mid-photo when the load inserts its
rows). `recomposeArticle` snaps such an offset to the image's top so the
two-stage contract holds even when a load lands while the user is reading at
the image's position.

Native border containment: on a native-image terminal a photo's kitty placement
extends past the viewport window when the photo is only partially visible,
drawing over the article border (observed as "images scroll into the border").
`nativeImageClear` deletes any placement not fully contained in the window
(`imgStart >= yOffset && capStart <= yOffset+vpH`), and
`suppressClippedNativeTransmits` blanks the transmit escape on a partially
visible image's first line in the rendered frame so the frame cannot
re-transmit it. The visible rows of a partially visible image would then render
blank, so `previewClippedImage` fills them with the halfblock preview (rendered
from the cached decoded image at the native block's row count, so it aligns and
centers identically): a photo that pokes into the window at its top or bottom
shows a blocky preview there rather than an empty strip (the fourth smoke
test's "empty space before The Trillion Dollar War Machine scrolls into view").
The result: an image appears as a sharp native photo only when fully in the
window, and a poking-in photo shows a preview, never a blank strip. This
alignment is the part most likely to need tuning on a real terminal; keep it
isolated in `snapYOffset`/`recomposeArticle` so the smoke test can adjust it
without touching composition.

### Tests

- `convert`: `ConvertImages` returns sentinels in document order + the URL list;
  `Convert` still emits `[alt]`/`[image]` (no regression).
- `internal/ui`: an article with an inline `<img>` composes an interleaved block
  with attribution; a late inline load inserts its block at the right offset;
  `snapYOffset` with a block list handles a mid-document block (down into the
  photo → caption, up into the photo → reveal, edge-snap on entry).
- `internal/image`: reuse of the existing commands for a second URL (no API
  change beyond persistence at positions 1..N via `SetArticleImage`).

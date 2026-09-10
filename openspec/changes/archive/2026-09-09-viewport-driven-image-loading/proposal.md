## Why

Image loading is eager and unbounded: opening an article fires a `tea.Batch` of
every image's fetch/decode at once, and each photo triggers its own expensive
native render (CatmullRom scale + PNG encode) the instant its bytes arrive. A
CPU profile of a photo-heavy article shows 47% of all CPU going to eager native
renders of all 11 photos simultaneously, saturating every core and starving the
UI goroutine that processes input. Loading and native rendering should be
proportional to what is visible, not to the article's length.

## What Changes

- Image loads become viewport-driven: on open, the article loads only the
  images in or near the visible viewport (a load frontier), and the frontier
  advances as the user scrolls. Scrolled-away images already fetched keep their
  cached block; images never reached are never fetched, decoded, or rendered.
- The native render is fired lazily: a photo whose block is composed but whose
  rows are not in view keeps the cheap halfblock placeholder and is rendered
  natively only when it approaches the viewport.
- Fetch/decode concurrency is bounded — a small in-flight limit instead of one
  command per image all at once — so a 30-photo article cannot spawn 30
  simultaneous decodes and renders.
- The visible result is unchanged: images still appear in place with their
  captions, the scroll-snap behavior is preserved, and scrolled-into-view
  images render from cache when available.

## Capabilities

### New Capabilities

- `image-load-scheduling`: article image loading and native rendering are driven
  by the viewport and bounded in concurrency, so work is proportional to what is
  visible.

### Modified Capabilities

- `inline-image-rendering`: the "loads asynchronously" contract becomes
  viewport-driven — an article with many images renders text immediately and
  loads only the images near the reading position, loading the rest on scroll.
- `article-recomposition`: recomposition stays bounded as loads land — the
  per-message recompose contract is retained, and scroll-driven load frontiers
  must not increase recompose work per message.

## Impact

- `internal/ui/image.go` — `fireImageLoad`, `loadImageCmd`, `onPhotoLoaded`,
  `ensureImageSource`, and the scroll path (frontier tracking).
- `internal/ui/model.go` — in-flight bound state and scroll hookup.
- `internal/ui/compose` — snap geometry interaction (block rows are known only
  once a block is composed; the frontier must work with placeholder rows).
- `internal/image` — bounded-concurrency helpers for fetch/decode if needed.
- No storage or schema changes; the database image table is unchanged.
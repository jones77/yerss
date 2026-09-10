## Why

Opening a photo-heavy article does a burst of redundant work: a CPU profile of
the Helene article (11 inline photos, all already stored in the database) shows
~5s of CPU, 33 recomposes, and **349 halfblock renders on the UI goroutine** for
11 images. The recompose path re-renders a halfblock from the decoded-image
cache on every image message even though an identical, width-matched rendered
block is already cached, and the native render re-decodes the same JPEG the
halfblock path decoded moments earlier.

## What Changes

- `composeImageBlock` serves a width-matched block from the rendered-block cache
  (`ImgBlocks`) before re-rendering from the decoded-image cache (`ImgCache`),
  so a recompose never re-runs the mosaic render for an image whose block is
  already composed at the current width. Re-render from the decoded image only
  when no width-matched block exists (resize, first load).
- The native render path reuses the decoded image produced by the halfblock path
  instead of re-decoding the raw photo bytes a second time (`decodeCapped` is
  the single hottest function in the profile at 34% of CPU).
- A regression test drives the full open → fire-image-loads → message-loop →
  recompose pipeline on a synthetic multi-image article and asserts the number
  of renders performed through the UI renderer is O(N) in the image count, not
  O(N × recomposes).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `article-recomposition`: re-composition SHALL NOT re-render an image block
  that is already cached at the current width; the recompose serves the cached
  block and re-renders only when no width-matched block exists.
- `native-image-rendering`: the native render SHALL reuse the decoded image
  cached by the halfblock path instead of re-decoding the source bytes.

## Impact

- `internal/ui/image.go` — `composeImageBlock` cache-serve order.
- `internal/image/native_cmd.go` — `NativeCmd`/`renderNativeFitted` decode reuse.
- `internal/image/image.go` — cache-hierarchy helpers used by both paths.
- `internal/ui/article_image_test.go` (and new test file) — regression coverage.
- No storage or schema changes; persisted blocks/photos and rendering output are
  unchanged.
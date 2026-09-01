## Why

A terminal session running yerss against a kitty-graphics-protocol terminal
(Ghostty) ballooned to 9.4 GB. The cause is two-fold: the in-memory image
caches (decoded images, raw photo bytes, and native renders) grow without
bound for the whole session, and the kitty native-image path re-transmits each
visible photo's full base64 payload on every frame while never freeing the
terminal's cached image data. Images must render without unbounded memory
growth, on both the yerss side and the terminal side.

## What Changes

- Bound the session image caches (decoded images, rendered blocks, raw photo
  bytes, native renders) with a size cap and eviction so memory stays flat no
  matter how many articles are opened.
- Cap the source resolution at decode so an oversized image cannot be held
  full-size in memory.
- Re-key the native render cache so a viewport-height change no longer forces a
  full re-render and a new cache entry.
- Delete kitty images by their stable id when leaving an article (not merely
  clearing visible placements), so the terminal's image cache is freed.
- Fix the image in-flight guard so a resize mid-render cannot start a duplicate
  native render.
- Stop converting the article HTML to markdown twice per open (body render and
  link harvest share one pass).
- Collapse the two duplicated height-fit loops (halfblock and native) into one.
- Reuse a single HTTP client for image fetches instead of one per request.

## Capabilities

### New Capabilities

- `image-cache-eviction`: Bounds the in-memory image caches (decoded images,
  raw photo bytes, rendered blocks, native renders) with a byte budget and
  least-recently-used eviction, and caps the source resolution accepted for
  decoded caching.

### Modified Capabilities

- `native-image-rendering`: Native image placement cleanup now deletes the
  terminal's cached image by its stable id when leaving the article (not only
  visible placements), and the native render cache is keyed by content width
  rather than width-and-viewport-height so height resizes reuse the render.

## Impact

- Code: `internal/image/image.go`, `internal/image/native.go`,
  `internal/image/native_cmd.go`, `internal/ui/image.go`,
  `internal/ui/article.go`, `internal/ui/model.go`.
- Tests: cache eviction/bounding, native cleanup escape emission, render-cache
  keying, in-flight guard, single-pass conversion.
- No new dependencies. No changes to the config surface, the store schema, or
  the public CLI.

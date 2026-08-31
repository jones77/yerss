## Why

On full-image-capable terminals (iTerm2/kitty/WezTerm), the native photo render
— decoding the source, scaling it to the display box, and re-encoding it as PNG
plus base64 — runs on the UI goroutine inside article composition. High-resolution
lead photos therefore freeze the UI on open and on every re-open, because the
rendered output is thrown away and recomputed each time instead of being cached.

## What Changes

- Native photo rendering SHALL move off the UI goroutine: the decode/scale/encode
  SHALL happen inside a background `tea.Cmd`, and the rendered escape-sequence
  lines SHALL be delivered to the UI via a message, like the block and photo fetch
  already are.
- The rendered native output SHALL be cached in memory keyed by render size, so
  re-opening an article at the same content width SHALL display the photo
  immediately from cache with no network request and no re-render. Resizing
  SHALL re-render asynchronously from the cached photo without re-fetching.
- The `maxImageBytes` download cap SHALL be removed; the fetch timeout and the
  `image/*` content-type allowlist remain the only guards. (This removes the
  "exceeded byte limit" failure mode from the async image load lifecycle.)

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `native-image-rendering`: the native render becomes asynchronous (never on the
  UI thread) and its output is cached in memory per size, so opening or resizing
  never blocks the UI and re-opening at the same size is instant.
- `reader-ui`: the async image load lifecycle no longer fails on an exceeded
  byte limit (the download cap is removed).

## Impact

- `internal/image`: new in-memory `Natives` cache (render-size → rendered lines),
  a new `NativeMsg` message, and a new render command that performs the
  decode/scale/encode (and the height-fit iteration) off the UI goroutine.
- `internal/ui`: the native branch of `articleImageBlock` becomes a cache read +
  compose with no rendering; the photo-load path fires the render command and
  re-composes on its message; resize re-renders via the command.
- `internal/image` `fetch()` drops the byte cap (keeps timeout + content-type
  allowlist).
- Tests: `internal/image` (cache, command, message) and `internal/ui`
  (`article_image_test.go`) updated for the new message flow.
- No new external dependencies.

**Sequencing note:** this change builds on `store-text-image-blocks`, whose
delta specs (including the new `native-image-rendering` capability) are complete
but not yet archived into `openspec/specs/`. Archive that change first so this
change's deltas apply against the post-archive spec tree.

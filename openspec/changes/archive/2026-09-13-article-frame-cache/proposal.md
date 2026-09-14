## Why

Flicking the mouse wheel at the top or bottom of the article view queues up
no-op scrolls that take seconds to drain. bubbletea runs `View()` synchronously
after every message and its message channel is unbuffered, so each wheel tick
that lands while already at a boundary re-runs the article's full lipgloss
viewport composition on the UI goroutine before the next tick — or the
subsequent scroll-down the user actually wanted — can be processed. The result
is a visible multi-second lag after a burst of boundary wheel input.

## What Changes

- Memoize the rendered article frame keyed on a cheap signature of the state
  the article view actually draws from (article, content generation, native
  clear generation, scroll offset, geometry, selection, glyph mode, link mode).
- Serve the cached frame when the signature is unchanged, so a no-op boundary
  scroll becomes O(1) and never re-runs the viewport composition, native
  re-show/suppress transforms, or border rendering.
- Keep the native-image clear sequence (`nativeClearPrefix`) outside the cached
  body so its per-frame transitions don't defeat the cache, and track the
  native transmit-state transitions the body render does depend on with a
  generation counter rather than a manual dirty flag.
- Preserve all existing rendering, scrolling, snap, native-clear, selection,
  and read-state output byte-for-byte — this is purely a performance change.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `reader-ui`: adds a requirement that a no-op mouse-wheel scroll at the
  article view's top or bottom boundary SHALL NOT re-render the article frame,
  so rapid boundary wheel input cannot backlog the UI.

## Impact

- `internal/ui/model.go` — `View()` and the frame-cache state (`frameKey`,
  `frameValid`, `frameStr`); `Update` already funnels native-clear through
  `recomputeNativeClear`.
- `internal/ui/article.go` — `renderArticle` becomes the memoized entry point
  (wrap the current body render); `newArticleState` bumps `contentGen`.
- `internal/ui/image.go` — `recomputeNativeClear` bumps `nativeGen` when the
  transmitted-id state (`nativeSent`) changes; `renderArticle` splits the
  native-clear prefix out of the cached body.
- Tests in `internal/ui/*` asserting a boundary no-op does not re-run the
  viewport render (a render-counting shim, mirroring the existing
  `ImgRenderer` counting-shim pattern), plus byte-identity of cached vs
  uncached output for a fixture with native images.

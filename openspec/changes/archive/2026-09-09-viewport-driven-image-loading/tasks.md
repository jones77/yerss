## 1. Anchor rows and frontier state

- [x] 1.1 Add an `anchorRows` slice to `articleState` aligned with `imageURLs`,
      filled by `composeArticle`/`spliceBody` (block `ImgStart` when composed,
      placeholder row while uncomposed)
- [x] 1.2 Add frontier state to the model: a monotonic frontier index and a
      `nativePending` set of URLs whose photo is fetched but native render is
      deferred
- [x] 1.3 Compute the frontier from `anchorRows` + viewport + lookahead margin
      (one viewport height), only ever growing within an article open

## 2. Viewport-driven loads

- [x] 2.1 Change `fireImageLoad`/`ensureImageSource` to fire loads only for
      images at or before the frontier (lead + first in-view inline images on
      open)
- [x] 2.2 Add a scroll hook: after a viewport move, advance the frontier and
      return a command firing the newly in-range loads
- [x] 2.3 Keep `imgLoading` as the per-URL in-flight guard so frontier advances
      never double-fire

## 3. Lazy native render

- [x] 3.1 Change `onPhotoLoaded` to record the URL in `nativePending` and start
      `nativeRenderCmd` only when the photo's block is in/near view; otherwise
      keep the halfblock placeholder
- [x] 3.2 On scroll/recompose, fire `nativeRenderCmd` for `nativePending` URLs
      now in view; clear them from the set once rendered (or on article close)

## 4. Bounded concurrency

- [x] 4.1 Add a shared semaphore (capacity ~3) on the model gating the expensive
      portion of the block, photo, and native commands
- [x] 4.2 Acquire/release around the fetch+decode+render work in each command
      path so at most the bound runs concurrently

## 5. Tests and verification

- [x] 5.1 Update existing open-path image tests that assert all images load on
      open to assert the frontier behavior instead (lead + first inline load;
      distant images do not until scrolled)
- [x] 5.2 Add tests: opening a multi-image article loads only in-view images;
      scrolling advances the frontier and loads the next images; a deferred
      native render fires when scrolled into view; concurrency never exceeds the
      bound
- [x] 5.3 Run the full `internal/ui` and `internal/image` suites plus `go vet`;
      re-run the Helene-article profile and confirm eager native-render CPU
      drops to near the in-view count
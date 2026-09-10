## 1. Cache-serve fix in composeImageBlock

- [x] 1.1 Reorder `composeImageBlock` in `internal/ui/image.go` to serve a
      width-matched `ImgBlocks` block before re-rendering from `ImgCache`,
      keeping the native-render cache lookup and the width guard intact
- [x] 1.2 Verify existing image/recompose tests still pass
      (`go test ./internal/ui -run 'Image|Recompose'`)

## 2. Native decode reuse

- [x] 2.1 Change `NativeCmd`/`renderNativeFitted` in `internal/image/native_cmd.go`
      to read the decoded image from the session `ImgCache` (keyed by URL) when
      present and skip its own `decodeCapped`; decode only on a cache miss
- [x] 2.2 Confirm the shared decoded image respects the 2048px width cap on both
      paths (no behavior change for oversized sources)

## 3. Regression test

- [x] 3.1 Add a test that builds a synthetic N-image article (small generated
      PNGs seeded into the store, no network), opens it, drains all image-load
      messages through `Update`, and asserts the session `ImgRenderer` performed
      at most `3N` renders and `recomposeCount` is bounded
- [x] 3.2 Remove the scratch profiling benchmark (`internal/ui/zz_prof_bench_test.go`)
- [x] 3.3 Run the full `internal/ui` and `internal/image` test suites and `go vet`

## 4. Verification

- [x] 4.1 Re-run the profiling benchmark harness (scratch or one-off) on the
      Helene article and confirm the render count drops from ~349 to O(N) and
      total CPU drops materially
- [x] 4.2 Manual check: open the Helene article on kitty/ghostty and confirm the
      article accepts input promptly while photos load

## 5. Follow-up: first inline image snaps when the lead is suppressed

- [x] 5.1 When the article's lead URL also appears inline (a promo banner
      reusing the featured photo), the lead is suppressed and the first composed
      block is a regular inline image not shown on open. Thread `leadShown`
      through `articleState` → `scrollArticle` → `SnapYOffset` so index 0 gets
      the inline entry/bottom snaps in that case instead of the pre-consumed
      lead treatment, and verify the suppressed-lead article snaps its caption
      bottom into view on the down scroll
- [x] 5.2 A native render on a short viewport can collapse a photo to a single
      image row with a multi-line caption; after the rise to its top the next
      down move lands on the caption's first line, past the photo range, so the
      caption scrolled line by line. Make `SnapYOffset` skip the whole block
      past its caption from the image's top so it leaves as one unit
- [x] 5.3 Source prose embeds U+00AD soft hyphens as invisible line-break hints;
      the terminal renders each as a visible cell where the app counts it as
      zero-width, so a line padded to the content width overflows the frame and
      wraps the right border (scrollbar) onto the next line's left. Strip soft
      hyphens from the article content in `computeDerivation` before rendering
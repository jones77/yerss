## Context

See proposal.md - Why. The current pipeline, measured on the 11-photo Helene
article: `composeImageBlock` checks the native render cache, then the
decoded-image cache (`ImgCache`, a hit once any load landed), then the
rendered-block cache (`ImgBlocks`). Because `ImgCache` is checked before
`ImgBlocks`, every recompose re-runs the mosaic render on the UI goroutine for
every image that is decoded but not yet natively rendered — 349 renders across
33 recomposes for 11 images. Separately, `NativeCmd` reads raw photo bytes from
`Photos` and calls `decodeCapped` even though the halfblock path already decoded
the same bytes into `ImgCache` — `decodeCapped` is 34% of the profile.

## Goals / Non-Goals

**Goals:**
- Serve width-matched rendered blocks instead of re-rendering them on recompose.
- Decode each image's bytes at most once per session across the halfblock and
  native paths.
- A regression test that pins the O(N) render bound on the UI renderer.

**Non-Goals:**
- Viewport-driven/lazy loading (separate change: viewport-driven-image-loading).
- Bounding native render payload size or per-frame re-transmit (separate change:
  trim-native-image-transmit).
- Changing any rendered output, storage layout, or cache eviction policy.

## Decisions

### D1: Serve `ImgBlocks` before `ImgCache` in `composeImageBlock`

Reorder the cache checks: when a width-matched block exists in `ImgBlocks`
(`image.BlockWidth(lines) == contentW`), serve it; otherwise fall through to the
native-render cache and the decoded-image re-render path. The existing
`composeImageBlock` already width-guards the `ImgBlocks` branch, so this is safe
on resize: a stale-width block fails the guard and the decoded-image re-render
takes over, exactly as today.

Small images render at a content-width-independent width (their natural width
capped at `smallImageMaxCells`), so the strict `== contentW` guard never matches
them on a normal-width column and they would keep re-rendering per recompose.
The guard therefore also serves a cached block when the cached decoded image's
source dimensions confirm `image.RenderWidth(srcW, srcH, contentW)` equals the
block's width — the same resize-safety property, applied to the small-image
case (a stale-width small block fails this check too and re-renders).

- Alternative considered: caching the composed block (image + caption) keyed by
  (url, width) on the article state. Rejected — the cache layer already holds the
  raw block; composing the caption lines is cheap relative to the mosaic render
  being eliminated, and a new cache would need the same invalidation logic.

### D2: Native render reads the decoded image from `ImgCache`

`NativeCmd`/`renderNativeFitted` takes the decoded image from `ImgCache` when
present (keyed by URL) and skips its own `decodeCapped`; when absent it decodes
as today. Both paths already apply the same width cap, so the native render
scales the identical capped bitmap.

- Alternative considered: threading the decoded image through the `PhotoMsg`
  → `NativeCmd` message flow. Rejected — `ImgCache` is already the shared
  handoff point keyed by URL; a message field would duplicate state that a
  resize-driven re-render (`ensureImageSource`) also needs to consult.

### D3: Regression test drives the full message loop

A test builds a synthetic article with N inline images (small generated PNGs
seeded into the store, no network), opens it, drains every image-load message
through `Update`, and asserts the session `ImgRenderer` (the UI-thread renderer)
performed at most `O(N)` renders (e.g. `≤ 3N`), with `recomposeCount` bounded.
This is the "soup to nuts" guard: it pins the whole open→load→recompose
pipeline, not a single function.

- The profile benchmark (`zz_prof_bench_test.go`, uncommitted scratch) is the
  template; it is replaced by the committed regression test.

### D4: Suppressed-lead first inline image snaps like an inline

When the article's lead image URL also appears as an inline image (a promo
banner reusing the featured photo, as in the Intercept article the reader
flagged), `computeDerivation` suppresses the lead (`leadURL == ""`), so the
first composed block is a regular inline image that was NOT shown on open. The
snap logic assumed index 0 is always the lead "shown on open" and excluded it
from the entry and bottom snaps, so that first image scrolled in line-by-line
and never snapped its caption into view. `articleState` now records whether a
genuine lead block was composed (`leadShown`), `scrollArticle` passes it to
`SnapYOffset` (and `recomposeArticle` to `SnapAfterRecompose`), and when it is
false index 0 follows the inline boundary transitions (entry snap to its bottom
boundary, rise, then skip). A genuine
lead keeps its pre-consumed stages; the existing lead-snap tests pin that.

- Alternative considered: distinguishing the lead by its URL or ImgStart
  position. Rejected — the promo banner reuses the lead's exact URL, so a URL
  match is ambiguous, and a position threshold is fragile for short intro
  paragraphs.

### D5: One-row image skips its caption as a unit

A native render on a short viewport collapses a photo to a single image row
(the fit reserves the multi-line caption and caps the image to one row), so
after the two-stage entry the next down move lands on the caption's first line
— exactly one row past the image top — which the photo-range skip
(`yOffset < CapStart`) cannot catch, and the caption scrolled line by line (the
halfblock placeholder, being the tall stored block, snapped instead, matching
the reader's "only snaps when it's half blocks" report). `SnapYOffset` now
skips the whole block past its caption when the move starts at the image's top
and lands on its caption, so the image and its caption leave as one unit. This
implements the spec's existing "a move starting at [the image] skips the block
past its caption" for the one-row case.

### D6: Strip soft hyphens from rendered content

Jacobin (and other sites) embed U+00AD soft hyphens in prose as invisible
line-break hints. The app's width functions count U+00AD as zero-width, but the
terminal renders each as a visible cell, so a line padded to the content width
displays one column wider per soft hyphen: the frame line exceeds the terminal
width and the right border (scrollbar) wraps onto the next line's left, and a
wrapped line near the top of a frame shifts the display and loses the top
border. `computeDerivation` strips U+00AD from the article's title, author,
link, and content before rendering, so the composed lines carry no character
whose terminal width the app miscounts.

## Risks / Trade-offs

- [Cache-serve changes behavior on halfblock terminals] → The width guard and
  the identical-output scenario in the spec cover it; `BlockCmd` already
  produces the same block the recompose would render.
- [Decode reuse couples the two paths through `ImgCache`] → `ImgCache` is
  keyed by URL only; a decode is never wrong, just possibly reused when the
  bytes changed (feeds re-fetch into a fresh session cache each run), and the
  store row's photo bytes are immutable per open.
- [Regression test asserts O(N) but N is small] → The bound is a ratio
  (`≤ 3N`), so it holds at any N; the failure mode it catches is the
  O(N × recompose) blowup (measured 32N today).

## Migration Plan

No storage or config migration. The change is a pure in-process behavior fix;
rollback is reverting the composeImageBlock reorder and the NativeCmd decode
reuse.

## Open Questions

None.
## Context

See proposal.md - Why. Today `fireImageLoad` fires one fetch/decode/render pair
per image in a single `tea.Batch`, and `onPhotoLoaded` starts a native render
for every photo the instant its bytes land. The profile shows 47% of CPU in
eager native renders of all 11 photos at once. The article state already knows
each image's document order (`imageURLs`), and `spliceBody` already reserves a
blank placeholder line per uncomposed image so the text flow stays stable — the
frontier can lean on that.

## Goals / Non-Goals

**Goals:**
- Load only images in/near the viewport on open, advancing the frontier on scroll.
- Defer native renders until the photo approaches the viewport.
- Bound concurrent fetch/decode/render operations to a small constant.
- Keep the rendered output, scroll-snap behavior, and reading-position rules
  unchanged from today.

**Non-Goals:**
- Changing the per-message recompose contract (covered by
  eliminate-redundant-image-renders).
- Payload/transmit size reduction (trim-native-image-transmit).
- Cache eviction or storage changes.

## Decisions

### D1: Frontier is a monotonic index into `imageURLs`

The frontier is the largest index `k` such that image `k` (lead + inlines in
document order) is within the viewport plus a lookahead margin. It only grows
within an article open. Index monotonicity makes the frontier immune to row
shifts: when a load lands and the recompose shifts every row below it, the
frontier is recomputed from fresh anchors and never re-fires an already-loaded
image (`imgLoading` remains the per-URL in-flight guard).

- Alternative considered: a row-based frontier (load images whose row is below
  the fold). Rejected — rows are unstable across recomposes; an index over the
  stable `imageURLs` order needs no invalidation.

### D2: Anchor rows recorded alongside the composed article

`composeArticle`/`spliceBody` record, per `imageURLs` entry, the content row
where the image's block starts when composed, or the placeholder row while
uncomposed. This parallel `anchorRows` slice is what the frontier compares
against the viewport. It is recomputed on every recompose exactly like
`imageBlocks`, so it is always consistent with the current composition.

- Alternative considered: deriving anchors by walking `imageBlocks` + sentinel
  scans. Rejected — `imageBlocks` omits uncomposed images; a parallel slice is
  simpler than re-scanning the body markdown.

### D3: Native render fires when the photo enters the frontier

`onPhotoLoaded` records the URL in a `nativePending` set and starts the native
render only when the photo's block is in/near view. A photo outside the frontier
keeps its halfblock placeholder. On scroll (and on recompose), any `nativePending`
URL now in view gets its `nativeRenderCmd`. A cached native render at the current
size displays immediately when scrolled into view (`composeImageBlock` already
serves `ImgNatives` first).

- Alternative considered: rendering natively on arrival regardless of position.
  Rejected — that is the current behavior being fixed (47% of CPU).

### D4: Concurrency bound via a shared semaphore

A small buffered channel (capacity ~3) on the model gates the expensive portion
of the image commands (fetch/decode/render, including the native render). The
commands acquire before doing work and release after, so at most the bound of
operations run at once regardless of how many are fired. `fireImageLoad` still
fires a batch — the semaphore, not the batch structure, enforces the bound.

- Alternative considered: a completion-driven worker that fires the next load
  when one finishes. Rejected — it threads load scheduling through every message
  handler; a semaphore is local, testable, and keeps the message handlers'
  shape unchanged.

### D5: The scroll path advances the frontier

`scrollArticle` (and the recompose after a frontier advance) checks whether the
new viewport brings any not-yet-loaded image within the margin; if so it returns
a command firing those loads and any newly in-view `nativePending` renders. The
lookahead margin is one viewport height below the fold, so images begin loading
before the reader reaches them while never loading the whole article.

## Risks / Trade-offs

- [Existing tests assume all images load on open] → Tests that open an article
  and drain every image message (e.g. the new regression test from
  eliminate-redundant-image-renders) must scroll or assert on the frontier; the
  lead image and first inline are always in view, so open-path assertions on
  those hold.
- [Frontier + scroll-snap interaction] → An unloaded image is a placeholder row
  and takes no snap stages; when its block lands, the existing
  `SnapAfterRecompose` repositioning applies. The snap geometry tests remain the
  authority.
- [Lazy native render can leave a fast-scrolled photo as halfblock] → Acceptable:
  the halfblock is a valid render, and scrolling back re-fires the native render
  from the cached photo. Cached native renders always display immediately.
- [Semaphore scope] → If the bound is applied per-command rather than shared,
  the total concurrency is the bound times the command types; the semaphore must
  be shared across block, photo, and native commands to bound end-to-end work.

## Migration Plan

No storage or config migration. Behavior is a pure scheduling change; rollback
is reverting the frontier and semaphore wiring (the open path falls back to
firing all loads).

## Open Questions

None.
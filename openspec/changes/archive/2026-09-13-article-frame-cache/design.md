## Context

The article view's `renderArticle` runs the full render pipeline on every
`View()`: `viewport.View()` (a lipgloss `Width/Height/MaxWidth/MaxHeight.Render`
over the visible lines — measuring, wrapping, and padding every line), the
native re-show and clipped-transmit transforms, selection highlighting, and the
article border. bubbletea v1.3.10 runs `View()` synchronously after every
message (its event loop does `renderer.write(model.View())` unconditionally)
and delivers input over an unbuffered `chan Msg`, so a burst of boundary wheel
ticks — each a no-op scroll that `scrollArticle` already short-circuits to
`return nil` for the image-load side — still re-runs `renderArticle` once per
tick before the next message is read. See proposal.md — Why for the observed
impact.

Relevant current state:

- `scrollArticle` (image.go) records `before := viewport.YOffset`, runs the
  scroll closure, and returns early when the offset is unchanged — so the image
  frontier/load work is already bounded for no-ops. Only the render is not.
- `renderArticle` (article.go) is a pure function of the model state it reads,
  except that `rewriteNativeReShows` / `suppressClippedNativeTransmits` /
  `previewClippedImage` read `nativeSentPrev` and lazily populate `previewCache`,
  both of which are advanced in the update path (`recomputeNativeClear`) not the
  render path.
- `newArticleState` (article.go) is the single chokepoint every article-state
  construction flows through — open, empty state, `recomposeArticle`, resize,
  and the selection-path rebuild — so it is the one place to bump a content
  generation counter.
- `recomputeNativeClear` (image.go) runs as a `defer` in `Update` before the
  message switch and mutates `nativeSentPrev`, `nativeSent`,
  `nativeClearArticleID`, and `nativeClearPrefix`. `View()` prepends
  `nativeClearPrefix` to the frame (model.go), so the prefix is applied outside
  the body `renderArticle` returns.
- The derivation cache (`derivCache`/`derivKey`) already removes the
  HTML→markdown/glamour cost from recompose; the remaining per-frame cost is
  exactly the lipgloss viewport render this change memoizes.

## Goals / Non-Goals

**Goals:**

- Make a no-op boundary scroll O(1): the cached article frame is returned
  without re-running viewport composition, native transforms, or the border.
- Keep the frame cache correct for every render-relevant input, including the
  `viewport.SetYOffset` calls that bypass `scrollArticle` (resize, recompose
  snap, selection rebuild) — by keying on actual state, not a manual dirty flag.
- Preserve the article's displayed output byte-for-byte.

**Non-Goals:**

- The list view and the list cursor wrap-around behavior (unchanged).
- Coalescing/debouncing wheel messages (a different mechanism; the frame cache
  removes the render cost without changing input delivery).
- Memory-bounding the cache beyond its single-frame size, or caching the list
  view / popups.

## Decisions

### D1: Memoize the article body on a cheap signature of render state

Split the current `renderArticle` body into an uncached renderer and a memoizing
wrapper. `View()` calls the wrapper, which builds a cheap `articleFrameKey` from
the model fields the render reads and returns the cached string when the key
matches the last rendered key; otherwise it renders and stores `(key, frame)`.

The key is a value comparison, not a manually-toggled dirty flag. This is what
makes it robust against the `SetYOffset` bypasses: the offset is a key field, so
any scroll — via `scrollArticle` or a direct `SetYOffset` — invalidates the
cache exactly when the offset actually changed, and a no-op (offset unchanged)
hits the cache. The key fields are:

- `articleID` — the open article (`m.article.id`).
- `contentGen` — bumped in `newArticleState` (D3); covers viewport content,
  image blocks, header/title/date, and scroll-total changes from recompose,
  open, and resize.
- `nativeGen` — bumped in `recomputeNativeClear` when `nativeSent` changes (D2);
  covers the per-frame `rewriteNativeReShows` / `suppressClippedNativeTransmits`
  output that depends on `nativeSentPrev`.
- `yOffset`, `width`, `height` — the viewport offset and terminal geometry.
- `sel` — the `textSelection` struct (comparable: ints + bools).
- `ascii`, `linkExpand` — glyph mode and link display mode.

`palette` and `Display.ShowSeconds` are excluded because they are fixed for the
process lifetime (resolved once in `New`); `linkExpand` is included even though
toggling it recomposes (and thus bumps `contentGen`) so the key stays correct
without depending on that coupling.

**Alternatives considered:**

- *A dirty flag set in `scrollArticle`.* Misses the `SetYOffset` calls in
  resize/recompose/selection, which is exactly the bug-prone surface the
  value-keyed signature avoids.
- *Coalesce/debounce wheel messages.* Removes the backlog but still burns a
  render per delivered message and changes input timing; the user rejected it in
  favor of the cache.
- *Cache the whole `View()` output including the popup overlay and the native
  clear prefix.* Over-broad: the prefix and popup overlay are applied around the
  body and would otherwise thrash the cache.

### D2: Fold `nativeClearPrefix` out; track native transmit state with a generation

`View()` returns `m.nativeClearPrefix + s` — the clear sequence is prepended
outside the body. Keep it that way: the memoized body is `s` only, so the
prefix's per-frame transitions (including the delete-all on first entry into a
no-native article, gated by `nativeClearArticleID`) do not defeat the cache.

The body render *does* depend on `nativeSentPrev` (via `rewriteNativeReShows`
and `suppressClippedNativeTransmits`), which is `nativeSent` as it stood before
this frame's `computeNativeClear`. So `recomputeNativeClear` bumps `nativeGen`
whenever `computeNativeClear` actually changes `nativeSent` (compare the
pre-snapshot to the post-compute map, and `nativeClearArticleID`). Because the
maps are bounded by the article's image count, this is O(images) — far cheaper
than the viewport render it gates.

**Trade-off:** `recomputeNativeClear` still runs every `Update` (it is cheap),
and only the render it feeds is skipped. This keeps the native-clear semantics
identical while the expensive composition is memoized.

### D3: Bump `contentGen` in `newArticleState`

`newArticleState` is the sole constructor of `articleState`, so a single
`m.contentGen++` at its top covers open, recompose (which reassigns
`m.article = m.newArticleState(a)`), resize, and the selection rebuild. No other
site needs to remember to invalidate the frame for content changes.

### D4: Assert with a render-counting shim, not a byte-write assertion

The terminal write is already deduplicated by bubbletea's framerate renderer
(identical frames are skipped before line-diffing), so the win is the
synchronous `View()` cost, not the bytes written. Mirror the existing
`ImgRenderer` counting-shim pattern: a test that wraps the article's viewport
render (or counts `renderArticle` calls) and asserts a boundary no-op scroll
does **not** increment it, that a real scroll does, and that the cached output
is byte-identical to the uncached render for a fixture containing native
images. Assert complexity bounds (at most a constant number of renders for a
batch of no-ops), not exact counts.

## Risks / Trade-offs

- [A render input is missing from the key, so a state change serves a stale
  frame] → Cover the full read set of `renderArticle` (D1); add a
  byte-identity test comparing cached vs freshly-rendered output across
  fixtures that exercise native images, selection, resize, recompose, and
  link/ascii toggles.
- [`previewClippedImage` lazily populates `previewCache` during render, which a
  cache hit skips] → A partially visible native image only becomes previewable
  after a recompose (which bumps `contentGen` and forces a re-render), so the
  preview is already present before any no-op scroll that would need it.
- [`nativeGen` over-approximates and thrashes the cache] → It bumps only on
  actual `nativeSent`/`nativeClearArticleID` transitions; steady-state scrolls
  at a boundary leave it unchanged, so no-op scrolls still hit the cache.
- [The prefix is prepended but the cached body was rendered under a stale
  `nativeSentPrev`] → `nativeGen` forces a body re-render exactly when
  `nativeSentPrev` would change the body output, so the cached body and the
  prefix stay consistent.

## Migration Plan

- No data/schema or external dependency changes. Implementation order: D3
  (`contentGen`) + D2 (`nativeGen`) first (small, independent counters), then
  D1 (the memoizing wrapper), then D4 (tests). Land with tests green and
  `git diff` showing no unintended side effects.

## Context

The image pipeline has four session caches — decoded images (`image.Cache`),
rendered halfblock blocks (`image.Blocks`), raw photo bytes (`image.Photos`),
and native renders (`image.Natives`) — all keyed maps with no bound or
eviction. Native renders are keyed by `url\x00width\x00maxHeight`, so every
viewport-height change produces a new, never-freed entry whose payload is a
base64-encoded PNG of up to 4096 px. On kitty terminals the frame repaint
re-transmits each visible image's full payload, and cleanup uses `d=a` (delete
placements) but never `d=i` (delete image data), so the terminal's image cache
accumulates. See proposal.md for motivation.

Constraints: Go 1.27, bubbletea/lipgloss, SQLite via glebarez/go-sqlite. The
`AGENTS.md` image-scroll-snap and glyph-pair invariants are unchanged by this
work.

## Goals / Non-Goals

**Goals:**

- Bound all four caches so session memory is flat regardless of article count.
- Cap decoded source resolution so a large image cannot pin a large buffer.
- Make native renders reusable across viewport-height changes.
- Free terminal-side kitty image data on article exit.
- Remove the duplicate HTML→markdown conversion and the duplicate height-fit
  loop, and fix the in-flight guard so a resize cannot start duplicate renders.

**Non-Goals:**

- Changing the halfblock rendering output or the scroll-snap geometry.
- Changing the OSC 1337 (iTerm) path's placement semantics.
- Removing or disabling native/kitty support.
- Changing the store schema or the config surface.
- A per-image data cap on persisted `photo` BLOBs (tracked separately if
  needed).

## Decisions

### D1 — One generic bounded cache replaces the four bespoke caches

Introduce a small `lru` type holding `map[string]entry` plus a FIFO/LRU
eviction list, parameterized by a `sizeOf(value) int64` function. `Cache`,
`Blocks`, `Photos`, and `Natives` become instances of it with per-instance byte
budgets. On `Set`, if the budget is exceeded, evict least-recently-used entries
until under budget. `Get` refreshes recency.

- Byte accounting: `Photos`/`Natives` use `len(bytes)` and
  `len(strings.Join(lines, "\n"))`; `Cache` uses `Dx()*Dy()*4`.
- Budgets are package constants, chosen well above one article's images but
  small enough to keep a long session flat (e.g. tens of MB total).

*Alternative considered:* a global byte pool shared across caches. Rejected —
separate budgets per cache are simpler and keep the "photo bytes must stay
longer than decoded pixels" ordering explicit.

### D2 — Decode with a pre-checked dimension cap

Before `image.Decode`, run `image.DecodeConfig` to read dimensions. If either
side exceeds `maxDecodedEdge` (4096), decode then downscale once (reusing
`golang.org/x/image/draw`) before `Cache.Set` and render. Small images pass
through unchanged. This replaces the current post-hoc `pixelDims` cap, which
only bounds the native re-encode.

*Alternative considered:* store a downscaled copy only on the native path.
Rejected — the halfblock path (mosaic) also works from the decoded image and
benefits from the same cap.

### D3 — Re-key the native render cache by content width

`NativeKey` drops `maxHeight`, becoming `url\x00width`. The height fit becomes
a composition-time concern, not a cache key. The halfblock and native fits are
collapsed into one shared helper (D5) so the reserved row count is computed the
same way both paths; the native render still caps to the viewport at render
time but is reused when only the height changed.

*Alternative considered:* bucketing height into coarse ranges. Rejected — width
alone is the true content determinant; height capping stays a render-time
parameter, and a height-only resize reuses the cached render.

### D4 — Delete kitty images by id on article exit

`nativeImageClear` gains a "leaving the article" branch: before emitting the
frame, iterate the previous article's `imageBlocks` and emit
`image.DeletePlacement(url)` (which sends `d=i,i=<id>`) for each native image,
instead of only `m.imgNative.Clear()` (`d=a`). `d=i` frees both the placement
and the cached image data, so the terminal's cache is reclaimed.

*Alternative considered:* a single `d=a` on exit plus periodic `d=i` sweep.
Rejected — `d=a` only clears visible placements and never the image data; the
explicit per-image `d=i` is the only reliable free.

### D5 — Share one height-fit budget helper

Extract the `base - headerLines - 3`, `attrLines`, `i == 2` math from
`fitBlock` (`internal/ui/image.go`) and `renderNativeFitted`
(`internal/image/native_cmd.go`) into a single `fitBudget(...)` (or a shared
closure) so the halfblock and native paths cannot drift. `LayoutWrapLines` stays
the single line-count source.

### D6 — Single-pass conversion and reused HTTP client

`newArticleState` computes the marked body markdown once and passes it to link
harvesting, removing the second `convert.ConvertImages` call in
`harvestArticleLinks`. The image `fetch` uses a package-level `http.Client`
(reused across requests) instead of constructing one per call.

### D7 — In-flight guard scoped to the native render stage

The `imgLoading` guard is cleared only on `NativeMsg`/`FailedMsg` (or a
terminal block failure), not on the early `BlockMsg`/`PhotoMsg`, so a resize
mid-render does not start a second native render for the same URL.

## Risks / Trade-offs

- [Eviction causes a visible re-render or re-fetch after a cache miss] → Misses
  re-derive from the stored `article_images` photo first, so network is only
  touched when the photo was never stored; budgets are sized to make this rare.
- [Deleting kitty images by id could clear an image a re-render still needs] →
  Deletion happens only when leaving the article view; re-entering re-renders
  from the photo cache, and the stable id is re-transmitted on re-entry.
- [Budget accounting for decoded images is an approximation (Dx×Dy×4)] → It is
  a conservative upper bound for RGBA and correct for the RGBA outputs the
  scale path produces; acceptable for an eviction heuristic.
- [Reducing the native key to width changes cache-hit behavior] → Behavior is
  covered by the new viewport-height-reuse scenario; the render still caps to
  the current viewport at render time so no oversized block is served.

## Migration Plan

No data migration. On the next build, existing sessions start with empty caches
and populate under the new bounds; stored `article_images` rows are unchanged.
Rollback is a git revert; no store or config changes are involved.

## Open Questions

None.

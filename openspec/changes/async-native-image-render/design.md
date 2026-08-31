# Design: async-native-image-render

## Context

Today the native photo render is the only remaining CPU-bound work on the
bubbletea UI goroutine. `PhotoCmd` fetches the photo, decodes it, renders and
persists the halfblock placeholder, and caches the raw bytes — all off-thread.
But `articleImageBlock` (`internal/ui/image.go`) then calls
`NativeRenderer.Render` *during* article composition (`newArticleState`, in the
`Update` loop): it re-decodes the raw bytes, scales to the display box, PNG
encodes, and base64 encodes, and its `fitBlock` loop repeats that up to three
times. The rendered lines are never cached, so every open, resize, and re-open
pays the full cost on the UI thread. See proposal.md for the motivation.

## Goals / Non-Goals

**Goals:**

- No decode/scale/encode of a native photo ever runs on the UI goroutine.
- A native render is computed once per (article, render size) and cached, so
  re-opening at the same size is a pure cache read.
- Resize re-renders from the cached photo off-thread, never re-fetching.
- Preserve the existing height-fit behavior (block + attribution + one body line
  within the viewport) and the scroll-position-preserving recompose.

**Non-Goals:**

- No change to the halfblock path (it already renders off-thread and is fast).
- No eviction policy on the in-memory caches — unbounded for the session, per the
  "lots of RAM is fine" decision.
- No dropping of the raw-bytes `Photos` cache (kept as-is; not in scope).

## Decisions

### 1. A render command + a `Natives` cache, instead of on-thread rendering

Add to `internal/image`:

- `Natives` — a thread-safe cache mapping a composite render-size key to the
  rendered native block lines. The key is `NativeKey(url, width, maxHeight)`, a
  single string joining the three fields; both the command (writing) and the UI
  (reading) derive it via the exported helper so they cannot drift.
- `NativeMsg{Key, Lines}` — the rendered native lines delivered to the UI,
  mirroring `BlockMsg`.
- `NativeCmd(native NativeRenderer, photos *Photos, natives *Natives, url string,
  width, maxHeight int, attr string, headerLines int) tea.Cmd` — reads the raw
  bytes from `photos`, runs the fit loop, renders natively off-thread, caches the
  result, and returns `NativeMsg`. On a cache hit it short-circuits to a
  `NativeMsg` immediately (like `PhotoCmd` short-circuits on a `Photos` hit).

The UI's `articleImageBlock` native branch becomes:

```
if lines, ok := m.imgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
    return m.composeBlock(lines, attr, contentW), imageRows(lines, ...), true
}
return nil, 0, false
```

No decode, scale, encode, or fit loop on the UI goroutine — just a map read and
string padding in `composeBlock`.

### 2. The height-fit iteration moves into the command (accepted coupling)

`fitBlock` (`internal/ui/image.go`) iterates because the height cap depends on
the rendered width, which depends on the cap. It cannot iterate against a
goroutine, so it moves into `NativeCmd`, which owns the whole fit decision. To do
that the command needs the attribution text and the header-line count; this
couples `internal/image` to reader layout, which was explicitly accepted ("I
don't care"). The wrap-count logic uses `ansi.StringWidth`, already imported by
the image package, and a small local copy of the wrap-line count so `composeBlock`
(still in the UI) and the command's fit loop agree. The command returns the
*fitted* lines; the UI still centers and wraps the attribution in `composeBlock`
(cheap).

Alternative considered: pre-rendering at a generous cap and truncating reserved
rows in the UI. Rejected — the kitty/iTerm placement is sized `w×h` at encode
time, so truncating trailing rows would let the terminal-side image overlap the
following text; the height cap must be known before encoding.

### 3. Message flow: `PhotoMsg` then `NativeCmd` then `NativeMsg`

Native open fires `Batch(BlockCmd(fetch=false), PhotoCmd(...))` as today.
`PhotoMsg` no longer triggers an on-thread render; its handler fires `NativeCmd`
(which short-circuits to a cache hit on re-open). `NativeMsg` recomposes the
article. This keeps `PhotoCmd`'s responsibility (fetch the bytes) separate from
the render, and re-opening an already-rendered article costs two near-instant
message round trips with zero CPU. `imgLoading` continues to guard against
duplicate in-flight loads; the `NativeCmd` result also clears it.

### 4. Resize re-renders through the command

`ensureImageSource` / `hasImageSource` gain a `Natives` check: a cached native
render at the current width satisfies the article, so no load fires; on a width
change it fires the load, whose `PhotoMsg` (instant, bytes cached) → `NativeCmd`
(re-renders at the new width, off-thread) → `NativeMsg` (recompose). The rare,
slow resize path is therefore also off-thread and non-blocking.

### 5. Remove the `maxImageBytes` cap

`fetch()` drops the byte cap and the `io.LimitReader`/size check, keeping the
15-second `fetchTimeout` and the `image/*` content-type allowlist. Rationale: the
cap was never a requirement, legitimate sites serve compressed photos, and RAM is
acceptable; the timeout still bounds wall-clock download time. This removes the
"exceeded byte limit" failure clause (see the `reader-ui` delta).

## Risks / Trade-offs

- Unbounded in-memory caches (raw bytes + decoded image + block + native lines)
  per photo, with no cap on downloaded size. → Accepted: the session lifetime is
  short and RAM is deliberately spent for speed; each fetch is still bounded by
  the timeout and the content-type allowlist.
- Moving the fit loop into `internal/image` couples it to reader layout. →
  Accepted; the wrap-count logic is small and shared via `ansi.StringWidth`.
- A misbehaving server could stream a very large body before the timeout. →
  Accepted trade-off of removing the cap; the timeout limits the window.
- Two message round trips on re-open. → Negligible (both are cache hits).

## Migration Plan

- No schema or persisted-data change; the `Natives` cache is in-memory only.
- The `NativeCmd`/`NativeMsg` change is additive to the message flow; old stored
  blocks and legacy photo bytes continue to work through the unchanged
  `BlockCmd`/`PhotoCmd` hierarchy.

## Open Questions

None that change the specs, approach, or task breakdown.

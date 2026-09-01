## Context

`stableID(url)` (`internal/image/native.go:187`) hashes only the URL into a
32-bit kitty image id, and `kittyTransmit` (`native.go:161`) sends the
base64-encoded PNG (re-encoded at the current cell box via `pixelDims`) under
that id with `i=<id>, p=<id>`. `DeletePlacement` (`native.go:137`) deletes by
that URL-only id. Because the payload's pixel dimensions change with content
width, a resize re-transmits different data under the same id. See proposal.md
for motivation.

Constraints: kitty graphics protocol (`i` image id, `p` placement id, `d=i`
delete), Ghostty as the primary kitty-family terminal under test.

## Goals / Non-Goals

**Goals:**

- Make each distinct render size a distinct, knowable terminal image id.
- Delete the prior render's id when a render is superseded by a resize.
- Ensure article-exit cleanup deletes every id the article transmitted.

**Non-Goals:**

- Stopping the per-frame re-transmit of the payload (a larger architecture
  change, tracked separately).
- Changing the OSC 1337 (iTerm) path, which has no terminal-side image cache
  keyed by id.
- Compressing or otherwise shrinking the native payload.

## Decisions

### D1 — Derive the kitty image id from URL and pixel dimensions

Replace `stableID(url)` with `renderID(url string, pxW, pxH int) uint32` that
hashes `url` plus the pixel dimensions, never zero. The placement id (`p`) can
remain the URL-only hash or follow the render id; the image id (`i`) must be
per-render so each render is independently deletable. `NativeRenderer.Render`
already computes `pxW, pxH`, so the id is derived there.

*Alternative considered:* keep `i` per-URL and rely on the terminal replacing
same-id re-transmits. Rejected — that relies on a terminal behavior that is not
guaranteed and is the suspected source of the observed growth.

### D2 — Record the id on each composed native block

`imageBlock` (`internal/ui/article.go:47`) gains a `nativeID uint32` (zero when
not a native block). The compose path reads the native block from the `Natives`
cache by key, so the id is read back out of the block's own kitty escape
(`image.NativeRenderID`, which parses the `i=` key from the escape it actually
transmitted) rather than threaded through `NativeMsg`/`NativeCmd`. The escape is
the single source of truth for what the terminal received, so cleanup and resize
deletion reference the exact id with no drift.

*Alternative considered:* recompute the id from the block's geometry at delete
time. Rejected — the geometry that produced the id must be retained anyway to
recompute it, so storing the id directly is simpler and avoids drift.
*Alternative considered:* thread the id through `NativeMsg`/`NativeCmd`. Rejected
— composition reads the block from the cache by key, not from the message, so a
message field would be lost before the block is built.

### D3 — Delete the prior id before a new-size transmit

`nativeSent` (a per-URL map of the kitty image id last transmitted) tracks what
the terminal holds. When a native render for a URL is replaced by one at a
different id (a resize), the frame emits `d=i,i=<old id>` for the prior id before
the new transmit: `nativeImageClear` deletes every held id the current frame does
not re-transmit under that exact id — scrolled-out or clipped placements, blocks
superseded by a re-render at a new geometry, and blocks that reverted to a
placeholder — and records the ids it does transmit.

*Alternative considered:* a delete-then-transmit on every frame for the same
URL. Rejected — it adds per-frame delete churn and does not resolve the
id-stability question directly.

### D4 — Delete all recorded ids on exit

Extending the exit-delete from `image-cache-eviction-and-native-fixes` (whose
exit path still fell back to the `d=a` delete-all), leaving the article view
emits `d=i` for every id in `nativeSent` — every image the terminal holds — and
clears the tracking, so the terminal's cached image data is freed deterministically
rather than only clearing visible placements.

## Risks / Trade-offs

- [A per-render id means the terminal can briefly hold two sizes during a
  resize transition] → The prior id is deleted before the new transmit, so the
  window is one frame and the old image is freed.
- [Recomputing the id must exactly match the id used at render time] → The id
  is derived from the same `pxW, pxH` `Render` computes, and stored on the
  block, so there is a single source of truth.
- [Ghostty may already replace same-id re-transmits, making this more
  conservative than necessary] → The scheme is correct regardless; a spike can
  later simplify if replacement is confirmed.

## Migration Plan

No data migration. Existing terminals start fresh; stored database images are
unaffected. Rollback is a git revert.

## Open Questions

- Whether Ghostty replaces or accumulates on same-id re-transmit (informs a
  possible later simplification, not the correctness of this change).

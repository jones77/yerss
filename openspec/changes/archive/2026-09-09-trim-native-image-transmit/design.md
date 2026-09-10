## Context

See proposal.md - Why. Today a native photo renders at full cell-pixel
resolution (cell box × the terminal's reported pixel-per-cell, e.g. ~1920px wide
on a retina terminal, up to the 4096 `maxEdge`) and `kittyTransmit` re-sends the
entire base64 payload on every frame the photo is visible — a multi-megabyte
escape sequence per keystroke. `nativeSent` already tracks the last transmitted
kitty image id per URL, which is exactly the bookkeeping a placement-reference
re-show needs.

## Goals / Non-Goals

**Goals:**
- Bound the encoded pixel dimensions of native renders at a fixed maximum edge
  independent of the terminal's cell pixel size.
- On kitty, re-show a visible photo by placement reference (`a=p`) instead of
  re-transmitting its payload on frames after the first.
- Preserve identity tracking, delete-by-id cleanup, and the rendered cell box.

**Non-Goals:**
- Changing the off-thread render pipeline, cache keys, or storage.
- Viewport-driven rendering (viewport-driven-image-loading).
- Redundant-render and decode-reuse fixes (eliminate-redundant-image-renders).

## Decisions

### D1: Fixed maximum encoded edge length

`pixelDims` keeps computing the full cell-pixel box but then clamps the longer
edge to a constant `maxNativeEdge` (e.g. 1280px) instead of the current 4096,
scaling the shorter edge to preserve aspect ratio. At typical content widths a
full-width photo encodes at the cap (~1280px, a few hundred KB of PNG) while
remaining sharp: the terminal upscales to the cell box.

- Alternative considered: encoding at 2 pixels per cell regardless of the
  reported cell size. Rejected — on a retina terminal that is a visible
  sharpness regression; the fixed-edge cap preserves sharpness up to the cap and
  only bounds the worst case.

### D2: Kitty placement-reference re-show

The native block's first line carries either the full transmit (`a=T`, as today)
or a placement reference (`a=p`) naming the same image id. The choice is driven
by `nativeSent`:

- First frame the image is displayed (or its data was freed by a scroll-out
  delete): full transmit, then record the id.
- Later frames while the image stays visible: `a=p,i=<id>,p=<id>,c=W,r=H,C=1`
  with no payload — the terminal re-places the cached image.

The reference escape still carries `i=<id>`, so `NativeRenderID` keeps parsing
the id and delete-by-id cleanup is unchanged.

### D3: Scroll-out clears the transmitted-id tracking

`computeNativeClear` currently deletes a scrolled-out placement by id but leaves
`nativeSent[url]` stale. With re-show-by-reference, a stale entry would make the
scroll-back frame emit a placement reference for data that was freed by the
`d=I` scroll-out delete. The scroll-out branch must clear `nativeSent[url]`, so
the scroll-back frame does a full transmit once, and later frames reference
again.

- Alternative considered: switching the scroll-out delete to the placement-only
  form (`d=i`) so data survives for reference re-show. Rejected — it changes
  terminal memory behavior and contradicts the existing data-freeing cleanup
  contract; clearing the tracking id is a one-line fix.

## Risks / Trade-offs

- [Reference re-show breaks if the terminal evicts the cached image] → The
  window between frames is tiny and the image was just transmitted; kitty keeps
  the cache. On any doubt, the clear-on-scroll-out rule forces a full transmit
  after any absence.
- [Encode cap reduces sharpness on very large photos] → The cap (1280) is at or
  above the cell-pixel width of most terminals at typical window sizes; only
  unusually large windows on high-DPI displays hit it, and the displayed box is
  unchanged.
- [Ghostty's partial delete handling] → Unchanged by this work; the existing
  Ghostty delete-all fallback on leaving the article still applies.

## Migration Plan

No storage or config migration. Rollback is reverting the escape construction
(always `a=T`) and the `pixelDims` cap.

## Open Questions

None.
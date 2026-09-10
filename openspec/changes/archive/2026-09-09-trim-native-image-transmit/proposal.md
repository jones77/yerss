## Why

Native image renders are encoded at full cell-pixel resolution (the cell box at
the terminal's reported pixel size, up to a 4096px edge) and their complete
base64 payload is re-transmitted to the terminal on every frame the photo is
visible. On kitty-family terminals a full-width photo can be a multi-megabyte
escape sequence written per keystroke, which makes the terminal — and therefore
input handling — feel sluggish even though the Go event loop itself is not
blocked.

## What Changes

- Native renders are encoded at a bounded resolution that keeps photos sharp
  while shrinking payloads: a lower encode cap than full cell-pixel resolution
  (e.g. 2× the cell box instead of the full reported pixel size, and/or a
  smaller `maxEdge` than 4096).
- On kitty graphics-protocol terminals, frames re-show a photo whose payload was
  already transmitted by referencing the cached image placement (`a=p` with the
  stable id) instead of re-transmitting the base64 bytes every frame. OSC 1337
  terminals (cell-bound) are unaffected and continue to re-emit, since the
  repaint erases them.
- The existing per-render-size identity, delete-by-id cleanup, and
  scrolled-out/clipped placement handling are preserved; the transmit change
  only alters how a fully-visible image is re-shown.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `native-image-rendering`: native renders SHALL be payload-bounded (encoded
  resolution capped below full cell-pixel resolution), and kitty frames SHALL
  re-display an already-transmitted photo by placement reference rather than
  re-transmitting its payload.

## Impact

- `internal/image/native.go` — `pixelDims` encode cap, `kittyTransmit`/escape
  construction, and a placement-reference path for re-show.
- `internal/ui/image.go` — `computeNativeClear`/`suppressClippedNativeTransmits`
  interplay with placement references (a re-shown placement must still be
  deleted when scrolled out or clipped).
- `internal/ui/article_image_test.go` — render/clear expectations.
- No storage or schema changes; persisted photo bytes and rendered blocks are
  unchanged.
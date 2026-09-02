## Why

The reader decodes every article image at its full source resolution and holds
that decoded bitmap in memory for the whole session. Opening one 4096×2731 hero
photo pushes the status-bar "RAM" figure to ~48 MB — not the download size, but
the decoded `width × height × 4` estimate — and it stays there after leaving the
article. The database already stores the compressed original bytes; the problem
is the full-resolution decoded copy held in memory.

## What Changes

- Decode article images — the lead image and each inline image — capped at 2048
  pixels wide. A source image wider than 2048 is downscaled on decode, before it
  is cached or rendered.
- Apply the cap to both the halfblock render path and the native render path.
- The database continues to store the original compressed bytes unchanged: no
  re-encode, no downscale-then-persist, no migration.

## Capabilities

### New Capabilities

- `image-decode-cap`: bounds the in-memory size of decoded article images by
  capping their decoded width, without altering stored bytes.

### Modified Capabilities

(none)

## Impact

- `internal/image`: one shared decode-and-cap helper used by `BlockCmd`,
  `PhotoCmd`, and `renderNativeFitted`; the decoded image cached in `ImgCache`
  never exceeds 2048 pixels wide.
- `internal/ui`: the status-bar "RAM" figure drops substantially per image.
- `internal/store`: unchanged — photo bytes remain the compressed originals.

## Non-Goals

- Bounded/LRU eviction of the in-memory caches (each entry is now smaller, but
  the cache count is still unbounded).
- Downscaling or re-encoding what is stored in the database.
- CDN URL rewriting to request already-resized variants (a possible future
  bandwidth optimization, not part of this change).
- Fixing the kitty graphics image-persistence cleanup bug.
## Why

Every time a stored photo is re-rendered at a new content width (a terminal
resize), the code writes the article's full raw photo BLOB back to SQLite along
with the new block. The photo bytes are already stored and unchanged, so the
write is pure amplification: it inflates the WAL, churns disk, and re-writes
multi-megabyte BLOBs on every resize. The photo bytes should be written once,
on acquisition; re-renders should update only the block.

## What Changes

- Split image persistence so the photo BLOB is written only on first
  acquisition (network fetch), and a block re-render updates the block and
  width columns without touching the photo.
- Change the stored-photo re-render path in the block loader to use the
  block-only write.
- Keep the fetch path writing block + photo together (unchanged behavior for
  first acquisition).

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `image-block-storage`: The photo bytes are written once on acquisition; a
  block re-render at a new width updates only the block and width, never the
  photo BLOB.

## Impact

- Code: `internal/store/store.go` (add a block-only upsert), `internal/image/image.go`
  (`persistImage` and its call sites in `BlockCmd`).
- Tests: store-level test that a block-only write preserves the existing photo;
  image-level test that a width-change re-render does not rewrite the photo.
- No schema change, no config surface change, no new dependencies.
- Related to `image-cache-eviction-and-native-fixes` (already proposed): that
  change re-keys caches; this one reduces storage write traffic. They are
  independent.

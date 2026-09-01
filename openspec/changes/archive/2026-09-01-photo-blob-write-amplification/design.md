## Context

`persistImage` (`internal/image/image.go:278`) writes the block, the photo BLOB,
and the width together via `Store.SetArticleImage` (`internal/store/store.go:601`),
whose `ON CONFLICT ... DO UPDATE` always assigns `photo = excluded.photo`. The
block loader's stored-photo branch (`image.go:212`) calls `persistImage` after
re-rendering a block from an already-stored photo, so every width-changing
resize re-writes the full photo BLOB. See proposal.md for motivation.

Constraints: Go 1.27, SQLite via glebarez/go-sqlite. The `article_images` table
has a composite primary key `(article_id, position)` and columns
`url, credit, block, photo, width`.

## Goals / Non-Goals

**Goals:**

- Stop re-writing the photo BLOB on block re-renders.
- Preserve the photo bytes when only the block changes.

**Non-Goals:**

- Capping the size of a stored photo BLOB.
- Changing when the photo is fetched or stored (still on acquisition).
- Compressing or otherwise transforming stored photo bytes.

## Decisions

### D1 — Add a block-only upsert alongside the full upsert

Add `Store.SetArticleImageBlock(id int64, position int, url string, block string, width int)`
that upserts `url`, `block`, and `width` while leaving `credit` and `photo`
untouched. It uses an `INSERT ... ON CONFLICT(article_id, position) DO UPDATE`
whose `SET` clause omits the `photo` and `credit` columns. The existing
`SetArticleImage` remains the full (block + photo + credit) upsert for first
acquisition and migration.

*Alternative considered:* add a `persistPhoto bool` flag to `SetArticleImage`.
Rejected — a separate method makes the two distinct operations explicit and
prevents a caller from accidentally dropping the photo by passing a nil slice.

### D2 — Route the stored-photo re-render through the block-only write

In `BlockCmd`, the branch that re-renders from a stored photo
(`len(img.Photo) > 0`, `image.go:208`) calls `SetArticleImageBlock` instead of
`persistImage`, since the photo is already stored. The network-fetch branch
(`image.go:236`) and `PhotoCmd`'s fetch branch (`image.go:270`) keep
`persistImage` (first acquisition). `persistImage` itself is left as-is; only
the one re-render call site changes.

*Alternative considered:* track a "photo already persisted" flag and pass it
through. Rejected — the stored-photo branch is self-evidently the case where
the photo is already on disk, so the call site alone determines which write to
use.

## Risks / Trade-offs

- [A block-only upsert could clobber a photo written concurrently by a fetch]
  → SQLite serializes writes per connection; the fetch and re-render both
  originate from the same store and the fetch path only runs when no stored
  photo exists, so the two do not race on the photo column.
- [Behavioral drift between the two upserts] → The block-only upsert shares the
  same conflict target and column list minus `credit`/`photo`; a store test
  asserts the photo survives the block-only write.

## Migration Plan

No schema migration. Existing rows are untouched; the next block re-render
simply stops rewriting the photo. Rollback is a git revert.

## Open Questions

None.

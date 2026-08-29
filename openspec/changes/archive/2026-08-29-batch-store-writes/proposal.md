## Why

Three store methods perform per-row SQL round trips that could be single batch
statements: `MarkAllRead` issues one `UPDATE` per article id, `SetArticleTags`
does a `SELECT id` plus two `INSERT`s per tag, and `UpsertArticle` does an
`INSERT … ON CONFLICT` followed by a separate `SELECT id` to recover the row id.
For a large feed refresh or "mark all read" across hundreds of articles this is
hundreds of needless statements in WAL mode. The fixes are behavior-preserving
and reduce both latency and write amplification.

## What Changes

- Rewrite `MarkAllRead` to a single `UPDATE … WHERE id IN (…)`, chunked to stay
  under SQLite's bound-parameter limit.
- Rewrite `UpsertArticle` to use SQLite's `RETURNING id` clause, removing the
  follow-up `SELECT id` round trip.
- Rewrite `SetArticleTags` to upsert categories and their join rows with fewer
  statements per tag (batch the category `INSERT … RETURNING id`, then insert the
  join rows in one pass).
- Keep the public method signatures unchanged so callers in `internal/ui` and
  `internal/feed` are untouched.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

_None._ This is an internal performance refactor with no observable behavior
change, so this change declares `skip_specs: true`.

## Impact

- **Data layer**: `internal/store/store.go` — `MarkAllRead`, `UpsertArticle`,
  `SetArticleTags` (and possibly a small chunking helper). Signatures unchanged.
- **Tests**: extend `store_test.go` to cover the batch paths (many ids, the
  `RETURNING id` path for insert vs update, tag replacement with duplicates and
  empty names) while keeping existing dedup/read-status behavior green.
- **Runtime**: fewer statements per refresh/mark-all; no change to the schema or
  to stored data.

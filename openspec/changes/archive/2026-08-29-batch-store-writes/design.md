## Context

See `proposal.md` - Why. The store uses `database/sql` with
`github.com/glebarez/go-sqlite` (a modernc pure-Go driver bundling a recent
SQLite ≥ 3.45), so `RETURNING` (≥ 3.35) is available. Existing helpers already
exist for null handling (`unixOrNull`, `nullIfEmpty`); no schema change is
needed.

## Goals / Non-Goals

**Goals:**

- Collapse per-row loops into batch statements where it is safe and clear.
- Keep public method signatures and observable behavior identical.

**Non-Goals:**

- No schema/migration changes, no index additions, no connection pooling or
  prepared-statement cache tuning.
- No change to the feed-fetch concurrency model.

## Decisions

**D1 — `MarkAllRead`: one `UPDATE … WHERE id IN (…)`, chunked.**

Build placeholders for the ids and run the update in chunks (e.g. 500 ids) to
stay well under SQLite's `SQLITE_MAX_VARIABLE_NUMBER` bound. The whole operation
already runs in a transaction; batching inside one transaction preserves
atomicity. Rationale: one statement per chunk replaces one statement per id.

**D2 — `UpsertArticle`: use `RETURNING id`.**

Change the final statement to `INSERT … ON CONFLICT(feed_url, guid) DO UPDATE
SET … RETURNING id` and scan the id directly from the same `Exec`/`QueryRow`.
Rationale: removes a separate `SELECT id` and any race between them.

- *Alternative considered*: keep the `SELECT` fallback for driver
  compatibility. Rejected — modernc supports `RETURNING`; a test pins it.

**D3 — `SetArticleTags`: batch category upsert with `RETURNING id`.**

Within the existing transaction: delete the article's join rows, then for each
unique tag name `INSERT INTO categories(name) VALUES(?) ON CONFLICT(name) DO
UPDATE SET name=name RETURNING id`, and insert all `(article_id, category_id)`
join rows in a second pass (a single multi-row `INSERT` or a prepared loop).
This replaces the per-tag `INSERT OR IGNORE` + `SELECT id` with one `RETURNING`
statement per tag and one batched join insert. Rationale: `RETURNING` collapses
the upsert+lookup; the join insert is the remaining hot loop.

- *Alternative considered*: a single `INSERT … SELECT` with a temp table.
  Rejected — overkill for the tag count seen in practice; keep it readable.

**D4 — Add a small `chunkIDs(ids []int64, n int)` helper** (or inline the
chunking in `MarkAllRead`) so the bound-parameter limit is handled in one place.

## Risks / Trade-offs

- **`RETURNING` availability** → Mitigation: the driver bundles SQLite ≥ 3.45;
  a `store_test.go` case exercises the `RETURNING id` path on both fresh insert
  and conflict update.
- **Parameter-limit edge case** (a "mark all" over a very large corpus) →
  Mitigation: chunking keeps each statement under the limit; a test marks all
  across > 1 chunk boundary.
- **Behavioral regression in dedup/read-status** → Mitigation: existing
  `store_test.go` (upsert dedup, read toggle, tag replace) must stay green;
  signatures are unchanged so callers are untouched.

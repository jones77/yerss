## Context

See `proposal.md` for motivation. The gate's trust check is `HasVerifiedFeed`, which
runs `SELECT COUNT(*) FROM feeds` and returns `n > 0`. But `UpsertArticle` runs
`INSERT OR IGNORE INTO feeds (url) VALUES (?)`, creating feed rows with NULL
`last_fetched_at`. The trust doc claims "a non-empty feeds table is proof a feed
parsed", which only holds because `UpsertArticle` is currently always preceded by
`UpsertFeed` (which sets `last_fetched_at`). The gate reads feeds via
`feed.LoadFeeds`, whose `os.Open` error the gate treats uniformly as the empty case;
a permission-denied error is indistinguishable from an absent file. `openArticle`
indexes `m.list.articles[m.list.cursor]` with only an `len == 0` guard.

## Goals / Non-Goals

**Goals:**
- Make the gate's offline trust depend on a recorded fetch time, not mere row
  presence.
- Give feeds-file read errors a distinct diagnostic and exit code from empty/missing.
- Make `openArticle` panic-proof against a stale/out-of-range cursor.

**Non-Goals:**
- Changing the exit codes for the empty (2), no-working-feed (3), or DB/load (1)
  cases.
- Adding new gate features (e.g., retry, partial trust). Pure correctness fix.
- Re-clamping policy changes in `loadList` (it already clamps; the guard is defense
  in depth).

## Decisions

**D1 — `HasVerifiedFeed` requires a fetch time.** Change the query to
`SELECT 1 FROM feeds WHERE last_fetched_at IS NOT NULL LIMIT 1` and return true iff a
row is found. This is a strict superset of the old behavior under the current call
ordering (every verified feed has a fetch time) and is immune to stub rows.
*Alternative:* add a separate `verified_at` column — rejected, schema migration for
no behavioral gain; `last_fetched_at` already records a successful fetch because
`UpsertFeed` only runs after a successful parse.

**D2 — Classify the feeds-file error in the gate.** `LoadFeeds` returns the raw
`os.Open` error (unwrapped) for open failures and a wrapped scanner error for read
mid-stream failures. In `gate.go`, after `LoadFeeds`, classify with
`errors.Is(err, os.ErrNotExist)`: if the error is not-exist → treat as missing
(exit 2, existing "no feeds configured" message); if the error is any other I/O
error → print `cannot read feeds file <path>: <err>` and exit 1; if no error but
`len(urls) == 0` → exit 2 (empty). *Alternative:* make `LoadFeeds` return a typed
sentinel — rejected, `os.ErrNotExist` already distinguishes the cases the OS can
distinguish, and wrapping would churn the feed package's API.

**D3 — `openArticle` bounds guard.** Replace the `len == 0` early return with a
single `if m.list.cursor < 0 || m.list.cursor >= len(m.list.articles) { return }`
guard before indexing. This covers empty, stale-high, and any defensive negative.
*Alternative:* re-clamp here instead of no-op — rejected; a no-op is safer and
matches "do nothing on invalid state".

## Risks / Trade-offs

- **[Tightened trust could reject a previously-accepted DB]** → Only if a feed row
  had NULL `last_fetched_at` but was considered verified before. `UpsertFeed` always
  sets it, so no legitimate verified feed is affected. Mitigation: T5 asserts the
  stub-only case is rejected and the verified case is accepted.
- **[Read-error exit 1 vs scripts expecting 2]** → Scripts keying on exit 2 for "no
  feeds" still get 2 for empty/missing; only genuine I/O errors now yield 1. Low
  risk; documented in the spec.
- **[openArticle no-op hides a real bug]** → A stale cursor reaching `openArticle`
  indicates `loadList` clamping misbehaved; silently no-oping could mask it.
  Mitigation: the guard is defense in depth; `loadList` clamping remains the primary
  control, and T10 documents the invariant.

## Migration Plan

No data or config migration. The `HasVerifiedFeed` query change is read-only against
existing rows. Rollback is a plain revert.

## Open Questions

- None. (Use exit code 1 for feeds-file read errors, consistent with the existing
  load/DB error code.)

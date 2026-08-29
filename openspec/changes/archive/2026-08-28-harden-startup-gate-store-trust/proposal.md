## Why

The startup gate's offline trust rule is fragile, and its refusal diagnostics
conflate two different failure modes. `HasVerifiedFeed` treats *any* row in the
`feeds` table as proof a feed was verified, but `UpsertArticle` inserts stub feed
rows with a NULL `last_fetched_at`; today this only works because of call ordering,
and any future caller of `UpsertArticle` would silently flip the gate to
"offline-trust" without ever verifying a feed. Separately, the gate reports a
feeds-file **read error** (permission denied, I/O failure) with the same "no RSS
feeds configured" message and exit code as a genuinely empty/missing file, which is
misleading. A latent bounds bug in `openArticle` indexes the article list without a
guard and would panic on a stale cursor.

## What Changes

- **Trust by fetch time, not row presence:** The gate's offline trust rule SHALL
  treat a feed as "previously verified" only when its `last_fetched_at` is set
  (non-NULL). `HasVerifiedFeed` SHALL require at least one feed with a recorded
  fetch time, so stub feed rows created by article persistence cannot satisfy the
  gate on their own.
- **Distinct feeds-file read-error handling:** When the feeds file exists but
  cannot be read, the gate SHALL print a distinct `cannot read feeds file <path>:
  <error>` message and exit with code 1, separate from the empty/missing case
  (exit 2). Empty and missing remain exit 2.
- **Open-article bounds safety:** Opening an article SHALL be a no-op when the
  cursor is out of range (empty list or stale index), and SHALL NOT panic.
- **Tests:** Add a store test that an article inserted via `UpsertArticle` only
  (no `UpsertFeed`) does not satisfy `HasVerifiedFeed`; a gate test that an
  unreadable feeds file produces the read-error message and a distinct exit from
  the empty case; and a UI test that `openArticle` with a stale/out-of-range
  cursor does not panic.

## Capabilities

### New Capabilities
<!-- None — all affected capabilities already exist. -->

### Modified Capabilities
- `feed-pipeline`: tightens the startup gate's "previously-verified feed" trust
  definition to require a recorded `last_fetched_at`, and distinguishes a feeds-file
  read error (exit 1, distinct message) from an empty/missing feeds file (exit 2).
- `reader-ui`: adds an open-article bounds-safety invariant so opening an article
  never panics on a stale or out-of-range cursor.

## Impact

- **Code:** `internal/store/store.go` (`HasVerifiedFeed` query), `cmd/yerss/gate.go`
  (separate read-error path from empty/missing), `internal/ui/article.go`
  (`openArticle` bounds guard).
- **APIs:** No public API changes. Internal: `HasVerifiedFeed` semantics tighten
  (stricter, safe superset of prior behavior under current call ordering).
- **Dependencies:** None added.
- **Behavior:** A feeds file that exists but is unreadable now exits 1 with a clear
  message instead of exiting 2 with the "no feeds configured" message. The gate no
  longer trusts stub-only feed rows. No change to verified-feed startup, refresh,
  or TUI behavior.

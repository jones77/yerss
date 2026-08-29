## Why

Reopening yerss always starts at the top of the list with the reader closed,
losing where the user was. For a daily reader this is a small but persistent
friction: after a restart you must re-find the article you were reading and
scroll back to the spot you left. Persisting the last selection — the list
cursor, an open article, and its scroll position — makes a restart land exactly
where the previous session stopped.

## What Changes

- Persist the reader selection on quit: the current view (list or article), the
  article under the cursor (or the day-group key when the cursor is on a day
  header), and, when the reader is open, the article's viewport scroll offset.
- On startup, restore the list cursor to the previously selected row (the
  article by ID, or the day header by its calendar-day key).
- When the reader was in the article view at quit, reopen that article and
  restore its scroll position, so pressing `b`/`esc`/`h` returns to the list
  with the same article still selected.
- Persist in the existing SQLite database via a new `state` key-value table; a
  selection that no longer exists (article pruned, day group gone) falls back
  to the default clamped cursor with no error.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: An "Restore reader state on start" requirement is added covering
  list-cursor restore, article-view reopen, scroll-position restore, and the
  fallback when the saved selection no longer exists.

## Impact

- **Store**: `internal/store/store.go` — new `state` table in the migration plus
  `LastSelection` (view, article id, header key, article offset) with
  `SaveLastSelection`/`LoadLastSelection`.
- **Model startup**: `internal/ui/model.go` — `Init` calls `restoreSelection`
  after the list loads.
- **Model quit**: `internal/ui/list.go`, `internal/ui/article.go`,
  `internal/ui/popup.go` — each quit branch calls `persistSelection`.
- **Selection logic**: `internal/ui/selection.go` (new) — persist and restore.
- **Dependencies**: none new — JSON marshaling and the existing SQLite store.
- **Tests**: store save/load roundtrip; cursor restore for article and header
  rows; article reopen with scroll-position restore; fallback when the saved
  article is gone.
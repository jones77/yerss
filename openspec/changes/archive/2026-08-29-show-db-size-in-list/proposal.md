## Why

The list view status bar already reports article counts and the last refresh
date, but gives no signal of how much local storage the database is consuming.
Once article images are persisted (see the `add-article-images` change), the
SQLite file can grow to hundreds of megabytes; surfacing the on-disk size in
the status bar makes that growth visible without leaving the TUI.

## What Changes

- Display the database size on disk in the list view status bar, formatted as
  a human-readable value (e.g. `1.2 MB`, `450 KB`).
- The size SHALL reflect the actual on-disk footprint of the database,
  including the WAL sidecar file when present (the store runs in WAL mode).
- The size SHALL appear on the right side of the status bar, alongside the
  last-refresh date.
- The size SHALL be refreshed when the article list is loaded (startup and
  after a refresh), not on every render frame.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "List view status bar" requirement is modified so the
  status bar also displays the on-disk database size, alongside the existing
  article count, active filter, and last-refresh date.

## Impact

- **UI layer**: `internal/ui/list.go` — `renderStatusBar` adds a
  human-readable size segment to the right cluster; the size value is computed
  in `loadList` (which already runs at startup and after refresh) and stored
  on the Model.
- **Data layer**: `internal/store` — new helper to report the on-disk size of
  the database (main file plus `-wal` sidecar when present), sourced from the
  resolved DB path.
- **Config**: `internal/config` — `DBPath()` is already available; no config
  changes.
- **Tests**: status bar renders the size; size formatting across unit
  boundaries (B/KB/MB/GB); WAL sidecar included when present, absent file
  handled gracefully.

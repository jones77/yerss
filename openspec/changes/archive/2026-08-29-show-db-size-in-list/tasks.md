## 1. Store layer

- [x] 1.1 Store the database path on the `Store` struct at `Open` time (add `path` field) so file-system queries don't depend on the config layer
- [x] 1.2 Add `DBSize() (int64, error)` to `Store` that `os.Stat`s the main DB file and the `-wal` sidecar (path + `-wal`), summing both sizes; treat a missing sidecar as zero, treat a missing main file as an error
- [x] 1.3 Test: `DBSize` returns main-only when no WAL exists, includes WAL when present, returns the sum, handles missing main file as an error

## 2. Size formatting

- [x] 2.1 Add a human-readable size formatter using binary unit suffixes (B, KB, MB, GB) with one decimal place for values ≥ 1 KB (e.g. `1.4 MB` for 1_500_000 bytes, `450 B` for 450 bytes)
- [x] 2.2 Test: formatting across unit boundaries (B, KB, MB, GB) and the exact `1_500_000 → 1.4 MB` case from the spec

## 3. UI integration

- [x] 3.1 Add a `dbSize int64` field to the `Model` and compute it in `loadList` via `m.store.DBSize()` (loadList already runs at startup and after refresh)
- [x] 3.2 Update `renderStatusBar` to append the formatted size to the right cluster, before the last-refresh date (e.g. `1.4 MB · last refresh 2026-08-28 15:04`); when refresh has never run keep the size + `never refreshed`
- [x] 3.3 Handle the width-calc: the right cluster now has two segments joined by ` · `; ensure the padding between left and right still accounts for total visible width
- [x] 3.4 Test: status bar contains the formatted size alongside the count and refresh date; size updates after a refresh completes

## 4. Validation

- [x] 4.1 Run `go test ./...` and linter; fix any failures
- [x] 4.2 `openspec validate --changes show-db-size-in-list --strict` passes
- [x] 4.3 Manual smoke test: confirm the size appears in the list view status bar and grows after a refresh that stores new articles

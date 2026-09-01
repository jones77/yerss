## Why

The team decided to keep session image caches unbounded rather than pre-emptively bound them. That decision is fine to revisit later, but only with real numbers: how large stored photos actually are, how many images an article holds, and how much memory decoding them all would take. There is currently no way to inspect this from the app.

## What Changes

- Add an `--image-stats` command-line flag that prints a read-only report of the database's image footprint and exits 0 without entering the TUI.
- The report covers the database size (with WAL), article and image counts, per-image photo-byte distribution, rendered-block bytes, and an estimated decoded-byte total derived from each photo's dimensions.
- Add a store method that returns the aggregated image statistics.

## Capabilities

### New Capabilities

- `image-stats`: a read-only diagnostic that reports on-disk and estimated in-memory image sizes.

### Modified Capabilities

## Impact

- `cmd/yerss`: new flag handling and report printing.
- `internal/store`: a new image-stats query.
- `go.mod`: none (dimension probing reuses the standard-library image packages already imported).

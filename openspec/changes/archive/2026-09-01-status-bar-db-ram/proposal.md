## Why

The status bar reports the on-disk database size but nothing about memory. With image caches now intentionally unbounded, the number worth watching is the session's decoded-image footprint. Showing both, as whole megabytes, makes growth legible at a glance.

## What Changes

- Change the status bar's right side from `?: help · <size> · <percent>% · <n>/<total>` to `?: help · RAM <n>MB · DB <n>MB · <percent>% · <n>/<total>`.
- `RAM` is the session's total decoded-image bytes; `DB` is the on-disk database size including the WAL sidecar.
- Both values render as a floored whole number of megabytes with no space between the number and `MB`; values under one megabyte render as `0MB`.

## Capabilities

### New Capabilities

### Modified Capabilities

- `reader-ui`: the list-view status bar right-side layout and formatting.

## Impact

- `internal/ui/list.go`: status bar rendering.
- `internal/image`: exposes the decoded-image byte total (added by the collapse-image-caches change).
- `cmd/yerss/initdb.go` and `internal/ui/size.go`: the `--init-db` prompt keeps the existing `formatSize`; the status bar uses a new whole-megabyte formatter.

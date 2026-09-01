## Why

Calendar-day bucketing and date formatting are implemented independently in the JSON export and the list view, and the day-ordinal logic is subtly wrong beyond two-digit days. A single date helper package removes the drift and fixes the ordinal.

## What Changes

- Add `internal/timeutil` with `DayStart`, `DayKey`, `LongDate`, `DayLabel`, and the shared time-layout constants.
- Reuse the shared day key in `cmd/yerss/json.go`'s bucketing.
- Fix the ordinal suffix to be correct modulo 100.
- No visible behavior change for real dates (days 1–31 already render correctly).

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/timeutil`: new package.
- `internal/ui/geometry.go`, `internal/ui/list.go`, `cmd/yerss/json.go`: use the shared helpers.

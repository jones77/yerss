## 1. Shared package

- [x] 1.1 Create `internal/timeutil` with the date helpers and layout constants
- [x] 1.2 Unit-test `DayKey`, `LongDate`, and the modulo-100 ordinal

## 2. Adopt

- [x] 2.1 Replace `dayStart`/`dayKey` in `internal/ui/geometry.go` with `timeutil`
- [x] 2.2 Replace `longDate`/`dayLabel` in `internal/ui/list.go` with `timeutil`
- [x] 2.3 Replace the inline day key in `cmd/yerss/json.go` with `timeutil.DayKey`

## 3. Validation

- [x] 3.1 Run `go test ./...`
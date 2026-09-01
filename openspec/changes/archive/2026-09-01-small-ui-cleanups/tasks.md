## 1. Cleanups

- [x] 1.1 Delete `sourceID` and call `store.SourceLabel` directly
- [x] 1.2 Replace float percent math with integer math in the status bar and tag popup
- [x] 1.3 Handle `LastRefreshedAt` errors by setting a status message
- [x] 1.4 Extract the header fold-key strings into named constants

## 2. Validation

- [x] 2.1 Run `go test ./internal/ui/... ./cmd/yerss/...`

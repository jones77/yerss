## 1. Remove dead code

- [x] 1.1 Delete `convert/wrap.go` (removes `WrapText`, `linkTokens`, `truncateURL`, `isTrailingPunct`, `linkOpenSeq`)
- [x] 1.2 Delete `convert/wrap_test.go` (removes the 14 `WrapText` tests)

## 2. Fix imports and verify

- [x] 2.1 Remove now-unused imports (`unicode`, `github.com/charmbracelet/x/ansi`, `strings`) from the `convert` package if `go vet` reports them
- [x] 2.2 Run `go build ./...` and confirm no references to `WrapText` remain
- [x] 2.3 Run `go test ./convert/...` and confirm the remaining `Convert` tests pass
- [x] 2.4 Run `go vet ./...` and confirm a clean result

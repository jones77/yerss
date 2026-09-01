## 1. Shared package

- [x] 1.1 Create `internal/textutil` with `FormatSize`, `EscapeMarkdown`, `MaxLineWidth` and unit tests

## 2. convert relocation

- [x] 2.1 Move `convert/` to `internal/convert/` and update all import paths

## 3. Dedupe call sites

- [x] 3.1 Replace `cmd/yerss/initdb.go` `formatSize` with `textutil.FormatSize`
- [x] 3.2 Replace `internal/ui/size.go` `formatSize` usage and remove the duplicate
- [x] 3.3 Replace the escape loops in `convert` and `internal/ui` with `textutil.EscapeMarkdown`
- [x] 3.4 Replace the three max-line-width implementations with `textutil.MaxLineWidth`

## 4. Validation

- [x] 4.1 Run `go test ./...` and `go vet ./...`

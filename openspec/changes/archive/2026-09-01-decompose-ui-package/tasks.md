## 1. Extract leaf packages

- [x] 1.1 Move rendering primitives (border, glyphs, palette, width, scroll) into `internal/ui/render`
- [x] 1.2 Move article composition (markdown, image blocks, snap) into `internal/ui/compose`

## 2. Components

- [x] 2.1 Reorganize list, article, and popup files by component
- [x] 2.2 Keep `model.go` as the thin coordinating top of the package

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...` and `go build ./...`

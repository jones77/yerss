## 1. Shared scanner

- [x] 1.1 Implement a single escape-aware cell scanner in `internal/ui`
- [x] 1.2 Unit-test it against the existing `skipEscape`/`styledCells` expectations

## 2. Adopt

- [x] 2.1 Rewrite `styledCells`/`cutStyledWidth`/`skipEscape` in `markdown.go` onto the scanner
- [x] 2.2 Rewrite `parseLinkSpans`/`oscEnd` in `article.go` onto the scanner

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...`

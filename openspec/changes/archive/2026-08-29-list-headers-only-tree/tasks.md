## 1. Rendering changes

- [x] 1.1 `internal/ui/list.go`: drop the `corner` parameter and the
      `rail := corner + g.h + " "` prefix from `renderArticleRow`; start the
      row at the publication time and widen the title budget by the freed
      three columns
- [x] 1.2 Replace `railGlyph`'s flat-index bookends with per-day-group corners:
      first group `┌`, last group `└`, interior `├`; have `renderList` pass the
      group index and count only for header rows

## 2. Tests

- [x] 2.1 Update `internal/ui/polish_test.go` call sites that pass a corner to
      `renderArticleRow` and re-verify `TestArticleRowSelectionBackground`
      still asserts the full-row highlight prefix

## 3. Verification

- [x] 3.1 Run `go build ./...` and `go test ./...`; confirm the list renders
      headers with `┌`/`├`/`└` and article rows with no glyph
- [x] 3.2 Run `openspec validate`
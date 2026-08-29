# Tasks: article-title-ellipsis

## 1. Glyph

- [x] 1.1 Add an `ellipsis` field to `borderGlyphs` in `internal/ui/border.go`
- [x] 1.2 Set `ellipsis: "…"` for Unicode and `ellipsis: "..."` for ASCII in `glyphsFor`

## 2. topBorder rewrite

- [x] 2.1 In `topBorder`, build `leftRail = g.tl+g.h`, `rightRail = g.h+g.tr`, `core = " "+date+" · "+title+" "`
- [x] 2.2 Compute `coreW`, `prefix`, `titleW`; if the title fits, render `leftRail + core + h*fill + rightRail`
- [x] 2.3 If the title does not fit, truncate it to `titleW - width(ellipsis)`, append `g.ellipsis`, and render `leftRail + core + rightRail` (no fill dashes)
- [x] 2.4 When `date == ""`, drop the `· ` separator so the core is `" "+title+" "`

## 3. Tests

- [x] 3.1 Title fits: no ellipsis, fill dashes present, full line width == w
- [x] 3.2 Title too long: ends with `…`, exactly one dash before the right corner, date intact, width == w
- [x] 3.3 ASCII mode: truncated title ends with `...`
- [x] 3.4 CJK title: truncation respects display width
- [x] 3.5 Empty date: no dangling `· ` separator

## 4. Verification

- [x] 4.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [x] 4.2 Visually confirm a long-title article shows `┌─ date · title… ─╖` and a short title still dash-fills
  (confirmed via rendered topBorder output: short → `┌─ 2026-01-02 · Short title ───╖`, long → `┌─ 2026-01-02 · This is a very long article ti… ─╖`, ascii → `+- ... -+`)

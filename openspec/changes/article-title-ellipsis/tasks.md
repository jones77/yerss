# Tasks: article-title-ellipsis

## 1. Glyph

- [ ] 1.1 Add an `ellipsis` field to `borderGlyphs` in `internal/ui/border.go`
- [ ] 1.2 Set `ellipsis: "…"` for Unicode and `ellipsis: "..."` for ASCII in `glyphsFor`

## 2. topBorder rewrite

- [ ] 2.1 In `topBorder`, build `leftRail = g.tl+g.h`, `rightRail = g.h+g.tr`, `core = " "+date+" · "+title+" "`
- [ ] 2.2 Compute `coreW`, `prefix`, `titleW`; if the title fits, render `leftRail + core + h*fill + rightRail`
- [ ] 2.3 If the title does not fit, truncate it to `titleW - width(ellipsis)`, append `g.ellipsis`, and render `leftRail + core + rightRail` (no fill dashes)
- [ ] 2.4 When `date == ""`, drop the `· ` separator so the core is `" "+title+" "`

## 3. Tests

- [ ] 3.1 Title fits: no ellipsis, fill dashes present, full line width == w
- [ ] 3.2 Title too long: ends with `…`, exactly one dash before the right corner, date intact, width == w
- [ ] 3.3 ASCII mode: truncated title ends with `...`
- [ ] 3.4 CJK title: truncation respects display width
- [ ] 3.5 Empty date: no dangling `· ` separator

## 4. Verification

- [ ] 4.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [ ] 4.2 Visually confirm a long-title article shows `┌─ date · title… ─╖` and a short title still dash-fills

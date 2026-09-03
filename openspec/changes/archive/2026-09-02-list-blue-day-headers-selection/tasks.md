## 1. Day header format and color

- [x] 1.1 Add a `selStatus` field to `modelStyles` in `internal/ui/model.go` and build it in `buildModelStyles` as `Foreground(p.StatusBar).Background(p.Selection)`
- [x] 1.2 Update `renderDayHeader` in `internal/ui/list.go` to render `corner + strings.Repeat(g.H, 4) + " "` before the label (Unicode `┌──── `, ASCII `+---- `), using `m.styles.status` when unselected and `m.styles.selStatus` when selected
- [x] 1.3 Update `internal/ui/daygroups_test.go` rail-prefix assertions: `┌ `→`┌──── `, `└ `→`└──── `, `railGlyph+" "`→`railGlyph+"──── "`, and the ASCII case `+ `→`+---- `

## 2. Theme-aware selection highlight

- [x] 2.1 Update `internal/ui/render/theme.go`: `DarkPalette.Selection` → `#242424`, `LightPalette.Selection` → `#d3d3d3`, and refresh the `Palette` doc comment that describes the selection role as `#707070`
- [x] 2.2 Update `internal/ui/polish_test.go` hardcoded `#707070` selection expectations to `#242424`
- [x] 2.3 Update `internal/ui/tags_test.go` hardcoded `#707070` selection expectations to `#242424`

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...` and confirm the render, day-group, polish, and tags tests pass with the new prefix format and selection colors
- [x] 3.2 Run `go vet ./...` and confirm no new issues
## 1. Move remaining pure helpers

- [x] 1.1 Move `overlay`/`spliceStyled` and `formatMB` into `render` (or `compose`) with their tests
- [x] 1.2 Move the glyph/status-bar pure parts (`railGlyph`, day-header/status-bar text builders) into `render`
- [x] 1.3 Run `go build ./...` and `go test ./...`

## 2. Extract `listview`

- [x] 2.1 Create `internal/ui/listview` with `listState`, `articleItem`, `dayGroup`, `rowKind`, `visibleRow` and the list rendering/input methods (`list.go`, `list_update.go`)
- [x] 2.2 Export the type and update `Model` to hold and delegate to it
- [x] 2.3 Move list tests; run `go build ./...` and `go test ./...`

## 3. Extract `popup`

- [x] 3.1 Create `internal/ui/popup` with `popupState`, `tagItem`, and the tag/links/help popup methods (`popup.go`, `popup_help.go`)
- [x] 3.2 Update `Model` to delegate popup rendering and input
- [x] 3.3 Move popup tests; run `go build ./...` and `go test ./...`

## 4. Extract `imageview`

- [x] 4.1 Create `internal/ui/imageview` from the `imageView` type and `image.go`
- [x] 4.2 Replace the `*Model` back-reference with explicit parameters/return signals (per design D2); update `Update`'s image-message routing
- [x] 4.3 Move image tests; run `go build ./...` and `go test ./...`

## 5. Extract `articleview`

- [x] 5.1 Create `internal/ui/articleview` with `articleState`, `articleLink`, composition, article rendering, and selection (`article.go`, `article_update.go`, `article_selection.go`)
- [x] 5.2 Wire its `imageview` dependency for block composition
- [x] 5.3 Move article tests; run `go build ./...` and `go test ./...`

## 6. Slim `Model` to wiring

- [x] 6.1 Reduce `model.go` to construction, `Update` routing, `View` composition, status, refresh, and keybindings
- [x] 6.2 Run `go build ./...` and `go test ./...`; confirm no behavior change

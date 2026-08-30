## 1. List view status bar

- [x] 1.1 Render a literal `?: help` hint as the first right-aligned element in
  `renderStatusBar` (`internal/ui/list.go`), separated from the database
  size/position indicator by the bullet
- [x] 1.2 Add a status bar test asserting `?: help` is the first right-aligned
  element with the bullet separator

## 2. Article view bottom border

- [x] 2.1 Render a literal `?: help` hint as the first right-aligned element of
  the position indicator in `bottomBorder` (`internal/ui/border.go`), reading
  `?: help · <percent>% · <bottomLine>/<totalLines>` with every bullet in the
  dim role
- [x] 2.2 Add a bottom-border test asserting `?: help` leads the right-aligned
  indicator with the bullet separator

## 3. Verification

- [x] 3.1 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm
  all pass
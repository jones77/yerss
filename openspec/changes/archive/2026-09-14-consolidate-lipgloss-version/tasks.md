## 1. Spike

- [x] 1.1 Enumerate every v1 lipgloss usage (`render/`, `model.go`, `border.go`, `popup.go`, `image.go`, `markdown.go`) and its v2 equivalent
- [x] 1.2 Confirm `lipgloss.Wrap` (the lone v2 use in `compose/reflow.go`) has a stable v2/v1/`x/ansi` behavior-equivalent
- [x] 1.3 Prototype the swap on one package and run its styling tests; record the direction (v2 default, or v1 fallback)
- [x] 1.4 Record the chosen direction in this change's design/notes before migrating

## 2. Migrate

- [x] 2.1 Migrate `internal/ui/render` (`theme.go`, `border.go`) to the target version; run `go test ./internal/ui/render/...`
- [x] 2.2 Migrate `internal/ui/model.go`, `border.go`, `popup.go`, `image.go`, `markdown.go` to the target version; run `go test ./internal/ui/...`
- [x] 2.3 Align `compose/reflow.go` with the target version (keep on v2 or move off it per the spike)
- [x] 2.4 Remove the losing major from `go.mod`; run `go mod tidy`

## 3. Validation

- [x] 3.1 Run `go build ./...`
- [x] 3.2 Run `go test ./...` — `polish_test.go`, `ascii_test.go`, `render_test.go` must pass byte-identically
- [x] 3.3 Run `go vet ./...`

## Why

The repository has **no commits and no CI**, so nothing ever runs `go build ./...`
or `go test ./...` to catch breakage. The headline defect the user hit — `go run
cmd/yerss/main.go` reporting `undefined: runStartupGate` — is a single-file
invocation artifact that a whole-module build gate would have prevented from
shipping (and would have stopped the single-file-run footgun from being the
documented command). Standing up CI also turns a linter on, which surfaces a real
security defect: on Windows, `openURLCmd` runs `cmd /c start "" <url>` with the
feed-controlled article URL, so a malicious `<link>` containing `&` or `|` can inject
additional commands. There is also no README telling users how to build or run the
program correctly.

## What Changes

- **CI workflow:** Add `.github/workflows/ci.yml` that runs on push and pull
  request: `go build ./...`, `go vet ./...`, `go test ./... -race`, and
  `golangci-lint run`. The whole-module compile step is the safety net that catches
  the "referenced symbol not defined" class of issue.
- **Linter config:** Add `.golangci.yml` enabling `gosec`, `govet`, `staticcheck`,
  `errcheck`, and related linters so the first pass has a defined rule set.
- **Windows URL-open safety:** `openURLCmd` on Windows SHALL NOT invoke `cmd.exe`
  with the raw URL. It SHALL open the URL via a non-shell opener that receives the
  URL as a single argument (`rundll32 url.dll,FileProtocolHandler <url>`), so
  feed-controlled metacharacters cannot inject commands. The macOS and Linux paths
  already pass the URL as a single exec argument without a shell and are unchanged.
- **Argument-safety test:** Add a test asserting the URL is passed to the platform
  opener as a single exec argument and that no `cmd.exe /c` shell form is used with
  the raw URL.
- **README:** Create `README.md` with build and run instructions, documenting
  `go run ./cmd/yerss` and `go build` (and explicitly *not* the single-file
  `go run cmd/yerss/main.go` form that fails to compile sibling files).

## Capabilities

### New Capabilities
<!-- None — the affected capability already exists. -->

### Modified Capabilities
- `reader-ui`: adds a URL open/copy argument-safety requirement so opening an
  article URL never routes a feed-controlled string through a shell, preventing
  command injection on Windows.

## Impact

- **Code:** `internal/ui/open.go` (Windows opener), `.github/workflows/ci.yml`
  (new), `.golangci.yml` (new), `README.md` (new), `internal/ui/open_test.go` (new
  or extended).
- **APIs:** No public API changes.
- **Dependencies:** `golangci-lint` is a CI tool, not a Go module dependency; no
  `go.mod` changes.
- **Behavior:** On Windows, `o` (open article URL) opens the link via
  `rundll32 url.dll,FileProtocolHandler` instead of `cmd /c start`, closing the
  injection vector. macOS/Linux behavior unchanged. CI now gates merges.

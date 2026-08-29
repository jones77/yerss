## 1. CI workflow and linter config

- [x] 1.1 Add `.github/workflows/ci.yml` running on push and pull_request: `actions/setup-go` with cache, then `go build ./...`, `go vet ./...`, `go test ./... -race`, then `golangci/golangci-lint-action` running `golangci-lint run`
- [x] 1.2 Add `.golangci.yml` enabling `gosec`, `govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`; set `max-issues-per-linter: 0` and `max-same-issues: 0`
- [x] 1.3 Run the same four commands locally to confirm a clean baseline before relying on CI

## 2. Windows URL-open safety

- [x] 2.1 In `internal/ui/open.go`, replace the Windows `exec.Command("cmd", "/c", "start", "", url)` with `exec.Command("rundll32", "url.dll,FileProtocolHandler", url)`; leave macOS (`open`) and Linux (`xdg-open`) paths unchanged; leave the copy path (Stdin pipe) unchanged

## 3. Tests

- [x] 3.1 T6: add a test in `internal/ui` asserting the URL is passed to the platform opener as a single exec argument (inspect the constructed `exec.Cmd` args, e.g. via a seam/test hook) and that the Windows branch does not use `cmd /c` with the raw URL. Use a `runtime.GOOS`-aware table so the test runs on all platforms

## 4. README

- [x] 4.1 Create `README.md` with: short synopsis; `go build ./cmd/yerss` and `go run ./cmd/yerss` as the correct build/run commands; config (`$XDG_CONFIG_HOME/yerss/config.toml`), feeds (`feeds.txt`), and database (`$XDG_DATA_HOME/yerss/yerss.sqlite`) locations; and an explicit note that `go run cmd/yerss/main.go` (single file) fails to compile because sibling files in `package main` are required

## 5. Verification

- [x] 5.1 Run `go build ./...`, `go vet ./...`, `go test ./... -race`, `golangci-lint run` locally; resolve any findings (the Windows opener should clear the `gosec` G204/injection finding)
- [x] 5.2 Confirm the linter passes on the changed `open.go` and the new test fails without the D3 fix (red/green)
- [x] 5.3 Commit the change so the new CI workflow has a commit to run on (the repo currently has zero commits)

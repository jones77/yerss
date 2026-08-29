## Context

See `proposal.md` for motivation. `internal/ui/open.go` builds the platform opener
with `exec.Command`. On Windows it uses `exec.Command("cmd", "/c", "start", "",
url)`; `cmd.exe` interprets `&`, `|`, etc., so a feed-controlled `<link>` can inject
commands. The macOS (`open`) and Linux (`xdg-open`) paths pass `url` as a single
argv element without a shell, and `copyURLCmd` pipes the URL to `pbcopy`/`clip`/
`xclip` via `Stdin` (already safe). The repo has no CI and no README; `git log`
shows zero commits.

## Goals / Non-Goals

**Goals:**
- Gate every push/PR on a whole-module `go build`, `go vet`, `go test -race`, and a
  `golangci-lint` pass.
- Close the Windows URL-open command-injection vector by using a non-shell opener.
- Tell users the correct build/run commands.

**Non-Goals:**
- Cross-compiling or releasing binaries in CI.
- Coverage thresholds or benchmark gates.
- Replacing `xdg-open` on Linux or `open` on macOS (they are already safe).
- Adding a Makefile (out of scope per the chosen CI scope).

## Decisions

**D1 — CI runs the standard Go gate plus golangci-lint.** `.github/workflows/ci.yml`
uses `actions/setup-go` (cache enabled) and `golangci/golangci-lint-action`. Steps:
`go build ./...`, `go vet ./...`, `go test ./... -race`, then `golangci-lint run`.
The `go build ./...` step is the one that compiles every package as a unit — the
exact safety net for the "undefined symbol from a sibling file" class. *Alternative:*
`go test ./...` alone (test compilation covers build) — rejected, an explicit
`go build ./...` is clearer in the CI log and catches packages with no tests.
*Alternative:* only run lint — rejected, lint does not guarantee a clean build.

**D2 — `.golangci.yml` enables a focused rule set.** Enable `gosec` (catches the
subprocess-with-variable G204 and injection patterns), `govet`, `staticcheck`,
`errcheck`, `ineffassign`, `unused`. Set `max-issues-per-linter: 0` and
`max-same-issues: 0` so nothing is silently truncated. *Alternative:* enable every
linter — rejected, noisy on a first pass and risks blocking on style nits.

**D3 — Windows opener switched to `rundll32 url.dll,FileProtocolHandler`.** Replace
`exec.Command("cmd", "/c", "start", "", url)` with
`exec.Command("rundll32", "url.dll,FileProtocolHandler", url)`. `rundll32` is not a
shell: it receives the URL as a single argv element and hands it to the URL protocol
handler, which opens the default browser. No `&`/`|` interpretation. *Alternative:*
quote the URL inside the `cmd /c start` form — rejected, quoting is fragile and
`cmd.exe` parsing has edge cases; removing the shell entirely is robust.
*Alternative:* pull in a third-party browser-opening package — rejected to avoid a
new dependency for a two-line fix. Note: `rundll32` requires the URL not to contain
certain whitespace edge cases; RSS `<link>` URLs are URL-encoded in practice, and the
prior `cmd /c start` form had the same constraint.

**D4 — README documents the correct invocation.** Create `README.md` with: synopsis,
`go build ./cmd/yerss` and `go run ./cmd/yerss`, the config/feeds/data file
locations, and a short note that `go run cmd/yerss/main.go` (single file) will *not*
compile because sibling files in `package main` are required. *Alternative:* stay
silent — rejected, the footgun just recurred.

## Risks / Trade-offs

- **[golangci-lint version drift]** → Pinned via `golangci-lint-action`'s version
  and/or `.golangci.yml` `run.go` pinning; CI fails loudly on version mismatch.
- **[Lint may flag pre-existing issues]** → The first run may surface findings beyond
  the Windows opener. Mitigation: fix findings needed to reach green within this
  change; the Windows opener is the known one. If a finding is a false positive,
  narrow its config with a documented rationale rather than `//nolint` blanket
  disables.
- **[`rundll32` availability]** → Present on all supported Windows client/server
  editions. Mitigation: the copy path is unaffected; open-failure is already
  surfaced as a status message via `urlActionMsg`.
- **[CI adds latency to PRs]** → Negligible for a module this size.

## Migration Plan

No data or config migration. CI is additive. The Windows opener change is a
runtime-behavior fix with no persisted state. Rollback is a plain revert.

## Open Questions

- Should CI also run on `workflow_dispatch` and schedule (e.g., weekly) to catch
  upstream linter/Go releases? Deferred — can be added later without changing these
  specs.

## 1. Dependency and path helpers

- [x] 1.1 Add `github.com/spf13/pflag` to `go.mod` and run `go mod tidy`
- [x] 1.2 Add exported path helpers to `internal/config` (e.g. `DefaultConfigPath()`, `DefaultFeedsPath()`) returning the XDG default locations, so `cmd/yerss` can resolve config.toml without parsing it

## 2. Editor invocation

- [x] 2.1 Create `cmd/yerss/edit.go` with an editor-resolution routine (`$EDITOR` then `$VISUAL`, no fallback) and a `sh -c` invocation that single-quote-escapes the target path and propagates the editor's exit code
- [x] 2.2 Define `exitNoEditor = 4` (and any other edit-related exit codes) mirroring the `gate.go` exit-code pattern
- [x] 2.3 Add a function that ensures the target file exists (via bootstrap or equivalent) before invoking the editor

## 3. Flag wiring in main.go

- [x] 3.1 Replace stdlib `flag` with `pflag`; register `-e`/`--edit-feeds` and `-c`/`--edit-config` as bool flags; remove the `--config`/`-config` flag
- [x] 3.2 Implement the dispatch: `-c` only edits config.toml then exits 0; `-e` only best-effort-loads config to resolve feeds path (falling back to XDG default on parse failure), edits it, then falls through to normal startup; both flags edit config then feeds then exit 0; neither flag runs today's flow unchanged
- [x] 3.3 Change `config.Load` call to `config.Load("")` (empty path = XDG default) since the `--config` flag is gone

## 4. feeds.txt seed enrichment

- [x] 4.1 Update the `feedsSeed` constant in `internal/config/bootstrap.go` to a header explaining one URL per line and `#` comment skipping, plus a commented-out `https://www.dropsitenews.com/feed` example line, retaining the phrase "one RSS URL per line" so `TestBootstrapCreatesFiles` stays green

## 5. Tests

- [x] 5.1 Test editor resolution: `$EDITOR` wins over `$VISUAL`; `$VISUAL` used when `$EDITOR` unset; both unset returns `exitNoEditor` and does not invoke any editor
- [x] 5.2 Test editor invocation via a marker-touching `EDITOR` script (e.g. `sh -c 'touch <marker>'`) to confirm the right file path is opened and arguments in `EDITOR` are honored
- [x] 5.3 Test `-e` honors a `feeds_file` override from config.toml and falls back to the XDG default when config is unparseable
- [x] 5.4 Test exit-code propagation when the editor exits non-zero
- [x] 5.5 Test the `-c`-only edit-then-exit-0 and both-flags edit-then-exit-0 flows do not enter the TUI
- [x] 5.6 Test the new `feeds.txt` seed contains the example feed URL line prefixed with `#` and the format header phrase (update `TestBootstrapCreatesFiles` if the retained phrase moved)

## 6. Docs

- [x] 6.1 Update `README.md`: document `-e`/`--edit-feeds` and `-c`/`--edit-config`, note `feeds.txt` supports comments and ships an example feed, remove the `--config`/`-config` instruction

## 7. Verification

- [x] 7.1 Run `go build ./...` and `go test ./...` (quiet flags) and confirm all pass
- [x] 7.2 Run the linter (e.g. `golangci-lint run`) and confirm no new findings
- [x] 7.3 Run `openspec validate --change add-edit-flags` and confirm the change validates

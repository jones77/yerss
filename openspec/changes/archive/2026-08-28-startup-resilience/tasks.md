## 1. Render safety (reader-ui)

- [x] 1.1 In `internal/ui/list.go` `renderList`, clamp the effective height to a minimum of 1 before the `lines[:height]` truncation, and skip overwriting the last line with the status bar when the resulting slice is empty
- [x] 1.2 In `internal/ui/model.go` `New`, seed the model with non-zero default width and height (e.g. 80×24) so the first frame renders before `WindowSizeMsg`
- [x] 1.3 Audit `renderArticle`/`renderArticleBorder` for any height/width-derived slice index and confirm it is already guarded (clamp to min 1); add a guard if any unguarded path is found
- [x] 1.4 Add render edge-case tests: `renderList` at height 0, height 1, width 0, empty article list, and a non-empty list at height 1 — assert each returns a string without panicking
- [x] 1.5 Add `listWindow` unit tests covering n≤height, cursor near top/bottom, and n>height wraparound
- [x] 1.6 Add a truncation test for a very long title and a double-width (CJK) rune title at narrow widths

## 2. Config schema registry and bootstrap (configuration)

- [x] 2.1 Create `internal/config/releases.go` defining `ConfigSchemaVersion` (int) and an ordered registry of options introduced per version (section, key, default, one-line doc)
- [x] 2.2 Populate the registry with all currently-supported options as the v1 baseline so the seeded template reflects the real schema
- [x] 2.3 Add a `config.BuildCommit` variable (settable via ldflags, defaulting to `"dev"`) for use in the seeded file header
- [x] 2.4 Write the full commented `config.toml` template constant (every option commented out, grouped by section, with the schema-version + build-commit header and a pointer to `feeds.txt`)
- [x] 2.5 Implement `config.Bootstrap(path string) error` that creates the config directory and seeds `config.toml` and `feeds.txt` (with a comment header) only when each file is absent; existing files are never overwritten
- [x] 2.6 Wire `Bootstrap` into `config.Load` (or `main`) so it runs before the gate and before the TUI; ensure an unwritable config directory is non-fatal (fall back to defaults)
- [x] 2.7 Add a test asserting every option key in `releases.go` appears in the template constant (drift check, per design D7)
- [x] 2.8 Add bootstrap tests: first run creates dir + both files with expected headers; rerun is a no-op; pre-existing `config.toml`/`feeds.txt` are preserved byte-for-byte

## 3. Config self-update (configuration)

- [x] 3.1 Implement parsing of the `# yerss config — schema vN` header from an existing `config.toml`
- [x] 3.2 Implement `config.MigrateFile(path string) error`: for each registry option whose version is greater than the file's recorded version, append it as a commented line under a "new in vN" banner only if its key is not already present anywhere in the file; then bump the header schema version to current
- [x] 3.3 Make `MigrateFile` best-effort: any read/parse/write error is swallowed and logged, never aborting startup
- [x] 3.4 Wire `MigrateFile` into the load path so it runs after `Bootstrap` when a `config.toml` exists
- [x] 3.5 Add self-update tests: older-version file gets new options appended and header bumped; running twice is idempotent (no duplicates); a manually-uncommented option is not re-appended; a file with no parseable header is left untouched; a simulated write failure does not abort

## 4. Startup feed-verification gate (feed-pipeline)

- [x] 4.1 Add `store.HasVerifiedFeed() (bool, error)` as a `SELECT COUNT(*) FROM feeds > 0` query
- [x] 4.2 Implement a bounded verify helper in `internal/feed` (e.g. `VerifyFeeds(st, urls, perFeed, overall) (FetchResult, error)`) that fetches concurrently with a 10s per-feed timeout and a 30s overall deadline, reusing the existing fetch+persist path
- [x] 4.3 In `cmd/yerss/main.go`, after `store.Open` and `Bootstrap`/load, run the gate before `tea.NewProgram`: load feeds; if empty/missing print the path and `os.Exit(2)`; if `HasVerifiedFeed` proceed; else print `verifying feeds...` to stderr, call `VerifyFeeds`, and on zero successful feeds print errors and `os.Exit(3)`
- [x] 4.4 Ensure a successful gate fetch counts as the first refresh so `Init()`'s `NeedsStartupRefresh` skips (verify `last_fetched_at` is stamped by the verify path)
- [x] 4.5 Add gate tests using temp dirs + the existing `withXDG` helper: empty feeds file → exit 2; missing feeds file → exit 2; populated feeds table → no network call (use a fake/short-circuited store or count calls); first run with a working feed → proceeds; first run with all-bad feeds → exit 3; `verifying feeds...` is written to stderr

## 5. Integration and verification

- [x] 5.1 Run `go vet ./...` and `go build ./...` cleanly
- [x] 5.2 Run `go test ./...` and confirm all new and existing tests pass
- [x] 5.3 Manually verify the end-to-end first-run flow on a clean `XDG_CONFIG_HOME`/`XDG_DATA_HOME`: confirm dir + `config.toml` + `feeds.txt` are created, the app refuses with exit 2 while `feeds.txt` is empty, and after adding a live URL it starts
- [x] 5.4 Manually verify a second launch (feeds table populated) starts instantly with no network check, including with the network disabled
- [x] 5.5 Run `openspec validate --change startup-resilience --strict` and resolve any reported issues

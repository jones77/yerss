## Why

Users must hand-edit `feeds.txt` and `config.toml` in an external editor today, and
discover the XDG paths themselves. Two flags that open each file in `$EDITOR` lower the
friction of first-run setup and ongoing feed management, matching the ergonomics of
`git config -e` and `crontab -e`. The first-run `feeds.txt` seed is also a single bare
comment, giving new users no example to uncomment.

## What Changes

- Add `-e` / `--edit-feeds`: open `feeds.txt` in `$EDITOR` (then `$VISUAL`), honoring a
  `[data] feeds_file` override from `config.toml` when present, else the XDG default
  path. After the editor closes, continue into the TUI so the freshly-edited feeds are
  re-read and verified by the startup gate.
- Add `-c` / `--edit-config`: open `config.toml` at the XDG default path in `$EDITOR`
  (then `$VISUAL`). After the editor closes, exit 0, because config changes (theme,
  keybindings, refresh options) are read at load time and require a relaunch to apply.
- When both flags are given, edit `config.toml` then `feeds.txt` sequentially and exit 0
  (no TUI).
- If neither `$EDITOR` nor `$VISUAL` is set, print an error to stderr naming both
  variables and exit with a distinct code (4); there is no fallback editor.
- Editor commands that include arguments (e.g. `EDITOR="code --wait"`) are invoked via
  `sh -c`, and the editor's non-zero exit code is propagated.
- **BREAKING**: Remove the existing `--config` / `-config` command-line flag. The config
  file is always loaded from the XDG default path. Path relocation of `feeds.txt` via the
  `[data] feeds_file` config option is unaffected.
- Enrich the first-run `feeds.txt` seed with a header explaining the format (one URL per
  line, `#` lines ignored) and a commented-out example feed,
  `https://www.dropsitenews.com/feed`, that users can uncomment. Comments are already
  skipped by the feed loader; this change only updates the seed content.
- Switch CLI flag parsing from the standard library `flag` package to `spf13/pflag`, which
  provides clean short/long alias support and an automatic `-h` / `--help`. This is a new
  direct dependency.

## Capabilities

### New Capabilities

<!-- None: both affected capabilities already exist. -->

### Modified Capabilities

- `configuration`: add a requirement for editing `config.toml` and `feeds.txt` via
  `$EDITOR` (`-c` / `--edit-config` and `-e` / `--edit-feeds`), with editor resolution,
  no-fallback error behavior, edit-then-exit vs edit-then-continue semantics, and removal
  of the `--config` path-override flag. Also modify the first-run bootstrap requirement so
  the `feeds.txt` seed includes a format header and a commented-out example feed URL.
- `feed-pipeline`: modify the feed URL source requirement to state explicitly that blank
  lines and lines beginning with `#` are ignored (the loader already does this; the spec
  does not yet record it).

## Impact

- **Code**: `cmd/yerss/main.go` replaces stdlib `flag` with pflag, drops the `--config`
  flag, and dispatches to a new `cmd/yerss/edit.go` (mirroring `gate.go`) that resolves
  paths, ensures target files exist, and invokes the editor. `internal/config/bootstrap.go`
  updates the `feedsSeed` constant. A small helper may be added to `internal/config` to
  expose the default config/feeds paths without parsing config.
- **Dependencies**: adds `github.com/spf13/pflag` as a new direct dependency.
- **Docs**: `README.md` documents `-e` / `--edit-feeds` and `-c` / `--edit-config`, notes
  that `feeds.txt` supports comments and ships an example feed, and removes the `--config`
  instruction.
- **Tests**: `internal/config/bootstrap_test.go` is kept green by retaining the phrase
  "one RSS URL per line" in the new seed (or updated deliberately if the phrase moves).
  New tests cover flag parsing, path resolution, and editor invocation (via a
  marker-touching `EDITOR`).
- **Breaking**: anyone scripting `yerss -config <path>` must move their config to the XDG
  default path. The `[data] feeds_file` config option still allows `feeds.txt` relocation.

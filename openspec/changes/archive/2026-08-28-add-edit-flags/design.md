## Context

See `proposal.md` for motivation. Today `cmd/yerss/main.go` uses the standard
library `flag` package with a single `--config` flag; startup is linear:
`flag.Parse → config.Load → store.Open → runStartupGate → tea.NewProgram`.
The feed loader (`internal/feed/feeds.go`) already skips blank and `#`-prefixed
lines. The first-run `feeds.txt` seed is a one-line constant
(`feedsSeed` in `internal/config/bootstrap.go`). The startup gate and edit
action both run before the TUI's `tea.WithAltScreen` takes over the terminal, so
an external editor can own the TTY cleanly.

## Goals / Non-Goals

**Goals:**

- Two CLI flags (`-e`/`--edit-feeds`, `-c`/`--edit-config`) that open their
  respective files in `$EDITOR`.
- Editor invocation robust to `EDITOR="code --wait"` (commands with arguments).
- `-e` honors the `[data] feeds_file` override; `-c` always targets the XDG
  default config path.
- A richer `feeds.txt` seed with a commented-out example feed URL.

**Non-Goals:**

- An in-app editor (all editing is delegated to an external `$EDITOR`).
- A built-in fallback editor when `$EDITOR`/`$VISUAL` are unset.
- Re-reading `config.toml` without a relaunch (config is read at load time; a
  relaunch is required for `-c` changes to take effect).
- Restoring a `--config` path-override flag (config.toml is always at the XDG
  default; feeds.txt relocation via `[data] feeds_file` remains supported).

## Decisions

### Use `spf13/pflag` for flag parsing

`pflag` provides native short/long alias support (`-e` and `--edit-feeds`
bound to the same value) and an automatic `-h`/`--help`. The standard library
`flag` package cannot bind a short and long name to one variable without
duplicate `BoolVar` registration that pollutes the usage output.

- **Alternative considered:** register both `"e"` and `"edit-feeds"` as
  `BoolVar` against the same `*bool` on stdlib `flag`. No new dependency, but
  the usage message lists both forms separately and `--help` is not automatic.
  Rejected because the project owner prefers clean shorthand ergonomics and a
  real help flag.

### Remove the `--config` flag

`config.toml` is always loaded from the XDG default path. This is a **breaking**
change for anyone scripting `yerss -config <path>`. Acceptable because
config-file relocation is uncommon and conventionally fixed at the XDG path.
`feeds.txt` relocation via the `[data] feeds_file` config option is unaffected.

- **Alternative considered:** keep `--config` alongside the new flags. Rejected
  by the project owner — only `--edit-config` is wanted, no path override.

### `-c` edits then exits; `-e` edits then continues into the TUI

`-c` exits 0 because theme, keybindings, and refresh options are read at
`config.Load` time and require a relaunch to apply. `-e` continues into normal
startup so the freshly-edited feeds are re-read and verified by the startup
gate — this is the ergonomic case (edit feeds, immediately read them). Both
flags together edit config then feeds and exit 0 (the user signaled "edit
files," not "read now").

- **Alternative considered:** both flags exit after editing (symmetric, matches
  `git config -e`). Rejected in favor of `-e` continuity for feed-editing
  ergonomics.
- **Alternative considered:** both flags continue into the TUI. Rejected because
  `-c` config changes silently wouldn't apply, which is misleading.

### Editor resolution: `$EDITOR` then `$VISUAL`, no fallback

If both are unset, print an error to stderr and exit with a distinct code
(`exitNoEditor = 4`), mirroring the startup gate's 2/3 exit-code pattern. No
built-in fallback editor (`vi`/`nano`) — the project owner chose to require an
explicit editor so behavior is never surprising.

### Invoke the editor via `sh -c`

The resolved editor string and the quoted target path are passed to
`sh -c "<editor> <quoted-path>"`. This handles `EDITOR="code --wait"` and any
editor with arguments, matching how `git` invokes the editor. The path is
single-quote-escaped inline (no extra dependency) to handle paths containing
spaces or shell metacharacters. The editor's exit code is propagated via
`cmd.Run()` and `ExitError`.

### New file `cmd/yerss/edit.go`, mirroring `gate.go`

A `runEdit(...)` function (or pair of functions) returns an exit code, keeping
`main.go` thin — the same structure as `runStartupGate`. Path resolution for
`-c` uses a small exported helper in `internal/config` (e.g.
`DefaultConfigPath()` / `DefaultFeedsPath()`) so config.toml's location is
resolved without parsing it (config.toml may be the broken thing the user wants
to fix). `-e` does a best-effort `config.Load` to honor `feeds_file`; on parse
failure it falls back to the default feeds path, edits it, then falls through to
the real `config.Load` (which exits 1 on a genuinely broken config, consistent
with today's behavior).

### Keep the `feeds.txt` seed test green

The new seed retains the phrase "one RSS URL per line" so
`TestBootstrapCreatesFiles` continues to pass without modification. The example
feed URL line is prefixed with `#`, so the loader skips it automatically.

## Risks / Trade-offs

- **[Breaking `--config` removal]** → Mitigation: README updated to remove the
  `-config` instruction; the `[data] feeds_file` option still covers feeds.txt
  relocation. Anyone affected can move config.toml to the XDG default path.
- **[New direct dependency on pflag]** → Mitigation: pflag is widely used and
  maintained; it is the standard choice for short/long flag aliases in Go CLIs.
  Bonus: a real `-h`/`--help` arrives automatically.
- **[`-e` with unparseable config]** → Mitigation: best-effort load falls back
  to the default feeds path for editing, then the real `config.Load` reports the
  parse error and exits — consistent with today's behavior on broken config.
- **[Editor fails to start or exits non-zero]** → Mitigation: the editor's exit
  code is propagated; a failure to spawn (e.g. `EDITOR` points at a missing
  binary) surfaces as a clear `sh` error on stderr with a non-zero exit.

## MODIFIED Requirements

### Requirement: TOML config file location

The system SHALL load configuration from a TOML file at the XDG default path `$XDG_CONFIG_HOME/yerss/config.toml` (fallback `~/.config/yerss/config.toml`). If the config file does not exist, the system SHALL start with built-in defaults.

#### Scenario: Config loaded from default path

- **WHEN** the application starts and the config file exists at the default path
- **THEN** the system loads configuration from that file

#### Scenario: Missing config uses defaults

- **WHEN** the application starts and no config file exists
- **THEN** the system starts with built-in default settings without error

### Requirement: First-run config bootstrap

On first run — when no `config.toml` exists at the resolved config path — the
system SHALL create the config directory and seed a default `config.toml`. The
seeded file SHALL contain every supported option as a commented-out line, grouped
by section, and SHALL begin with a header comment recording the current config
schema version and the build commit (falling back to `"dev"` when no commit is
embedded at build time). The system SHALL also seed a `feeds.txt` in the same
directory with a comment header stating that one RSS URL goes per line and that
lines starting with `#` are ignored. The seeded `feeds.txt` SHALL include a
commented-out example feed URL (`https://www.dropsitenews.com/feed`) that users
can uncomment to add an example feed. Files that already exist SHALL NOT be
overwritten or truncated; the bootstrap SHALL only create files that are absent.

#### Scenario: First run seeds config directory and files

- **WHEN** the application starts and neither `config.toml` nor `feeds.txt` exist in the resolved config directory
- **THEN** the system creates the config directory, writes a `config.toml` whose header records the current schema version and build commit and whose body contains every supported option as commented-out lines, and writes a `feeds.txt` containing only a comment header and a commented-out example feed URL

#### Scenario: Existing config.toml is preserved

- **WHEN** the application starts and a `config.toml` already exists in the resolved config directory
- **THEN** the system does not overwrite or modify that file during bootstrap

#### Scenario: Existing feeds.txt is preserved

- **WHEN** the application starts and a `feeds.txt` already exists in the resolved config directory
- **THEN** the system does not overwrite or modify that file during bootstrap

#### Scenario: Bootstrap is non-fatal on unwritable config directory

- **WHEN** the system cannot create the config directory or write the seed files
- **THEN** startup continues using built-in defaults rather than aborting, because built-in defaults require no config file to function

## ADDED Requirements

### Requirement: Edit configuration files via `$EDITOR`

The system SHALL provide a `-c` / `--edit-config` command-line flag that opens
`config.toml` at the XDG default path in the user's editor, and a
`-e` / `--edit-feeds` command-line flag that opens `feeds.txt` in the user's
editor. The editor SHALL be resolved from the `$EDITOR` environment variable,
falling back to `$VISUAL` when `$EDITOR` is unset. If neither variable is set,
the system SHALL print an error to stderr naming both variables and SHALL exit
with a distinct non-zero exit code without entering the TUI; the system SHALL
NOT fall back to any built-in editor. When the target file does not exist, the
system SHALL ensure it exists (creating it via the first-run bootstrap or an
equivalent step) before invoking the editor. When `-c` is given, the system
SHALL exit 0 after the editor closes, because configuration changes are read at
load time and require a relaunch to apply. When `-e` is given, the system SHALL
continue into normal startup after the editor closes, so the freshly-edited
feeds are re-read and verified by the startup gate. When both flags are given,
the system SHALL edit `config.toml` then `feeds.txt` sequentially and SHALL exit
0 without entering the TUI. An editor command that includes arguments (for
example `EDITOR="code --wait"`) SHALL be supported. The editor's non-zero exit
code SHALL be propagated as the process exit code.

#### Scenario: Edit feeds then continue into TUI

- **WHEN** the user runs `yerss -e` (or `yerss --edit-feeds`) and `$EDITOR` is set to a working editor
- **THEN** the system opens `feeds.txt` in the editor, and after the editor closes the system continues into normal startup, re-reading and verifying the feeds

#### Scenario: Edit config then exit

- **WHEN** the user runs `yerss -c` (or `yerss --edit-config`) and `$EDITOR` is set to a working editor
- **THEN** the system opens `config.toml` in the editor, and after the editor closes the system exits 0 without entering the TUI

#### Scenario: Both flags edit config then feeds then exit

- **WHEN** the user runs `yerss -e -c` and `$EDITOR` is set to a working editor
- **THEN** the system opens `config.toml` first, then `feeds.txt`, and after both editors close the system exits 0 without entering the TUI

#### Scenario: Edit feeds honors feeds_file override

- **WHEN** the user runs `yerss -e` and the TOML config specifies a custom `feeds_file` path
- **THEN** the system opens the feeds file at that custom path rather than the XDG default

#### Scenario: Edit feeds falls back to default when config is unparseable

- **WHEN** the user runs `yerss -e` and the TOML config exists but cannot be parsed
- **THEN** the system opens `feeds.txt` at the XDG default path and continues into normal startup, which reports the config parse error and exits without entering the TUI

#### Scenario: No editor set refuses to start

- **WHEN** the user runs `yerss -e` or `yerss -c` and neither `$EDITOR` nor `$VISUAL` is set
- **THEN** the system prints an error to stderr naming both environment variables and exits with a distinct non-zero exit code without entering the TUI

#### Scenario: Editor with arguments is supported

- **WHEN** `$EDITOR` is set to a command containing arguments (for example `code --wait`) and the user runs `yerss -e`
- **THEN** the system invokes the editor with its arguments against the feeds file rather than failing

#### Scenario: Editor non-zero exit is propagated

- **WHEN** the editor exits with a non-zero status code after being invoked by `yerss -e` or `yerss -c`
- **THEN** the process exits with that same non-zero status code and does not enter the TUI

#### Scenario: Help flag lists available flags

- **WHEN** the user runs `yerss -h` or `yerss --help`
- **THEN** the system prints usage listing `-e`/`--edit-feeds` and `-c`/`--edit-config` and exits 0 without entering the TUI

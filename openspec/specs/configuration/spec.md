## Purpose

Provides TOML-based configuration for keybindings, display preferences (padding, theme, ASCII fallback), and file path overrides, with sensible defaults that require no configuration file to start.

## Requirements

### Requirement: TOML config file location

The system SHALL load configuration from a TOML file at the XDG default path `$XDG_CONFIG_HOME/yerss/config.toml` (fallback `~/.config/yerss/config.toml`). If the config file does not exist, the system SHALL start with built-in defaults.

#### Scenario: Config loaded from default path

- **WHEN** the application starts and the config file exists at the default path
- **THEN** the system loads configuration from that file

#### Scenario: Missing config uses defaults

- **WHEN** the application starts and no config file exists
- **THEN** the system starts with built-in default settings without error

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time. Keybinding parsing SHALL be case-sensitive: a lowercase letter and its uppercase counterpart are distinct keys, so both MAY be bound to the same action without conflict. The default keybindings SHALL include `b` as an alias for the `back` action (alongside `esc`, `enter`, and `h`) and `o` as an alias for the `open_article` action in the list view (alongside `enter` and `l`). In the article view, `o` SHALL remain bound to the `open_url` action (opening the article URL in the browser); the per-view key resolution system ensures `o` resolves to `open_article` in the list view and `open_url` in the article view without conflict. Shifted-letter default bindings SHALL be reachable via both cases whenever the other case is not already bound to a different action: the tag popup SHALL be bound to both `T` and `t`, and refresh SHALL be bound to both `R` and `r`. The `bottom` action SHALL remain bound to uppercase `G` only, because lowercase `g` is already bound to `top`.

#### Scenario: Custom keybinding

- **WHEN** the config file maps the "quit" action to `["ctrl+x"]`
- **THEN** pressing Ctrl-X quits the application instead of the default `q`

#### Scenario: Conflicting keybindings rejected

- **WHEN** the config file binds the same key to two different actions
- **THEN** the system reports an error at load time and does not start

#### Scenario: Default b alias for back

- **WHEN** no custom keybindings are configured and the user presses `b` in the article view
- **THEN** the article view returns to the list view, identical to pressing `esc` or `h`

#### Scenario: Default o alias for open_article in list

- **WHEN** no custom keybindings are configured and the user presses `o` on an article row in the list view
- **THEN** the article reader view opens, identical to pressing `enter` or `l`

#### Scenario: Lowercase and uppercase both open the tag popup

- **WHEN** no custom keybindings are configured and the user presses either `t` or `T` in the list view
- **THEN** the tag popup opens

#### Scenario: Lowercase and uppercase both refresh

- **WHEN** no custom keybindings are configured and the user presses either `r` or `R` in the list view
- **THEN** a manual refresh is triggered

#### Scenario: Lowercase g stays bound to top

- **WHEN** no custom keybindings are configured and the user presses `g`
- **THEN** the cursor moves to the top, and `G` moves to the bottom (the two cases remain distinct)

### Requirement: Configurable article padding

The system SHALL allow the horizontal and vertical padding inside the article reader border to be configured via the TOML config file. Default horizontal padding SHALL be 2 spaces and default vertical padding SHALL be 1 space.

#### Scenario: Custom padding

- **WHEN** the config file sets padding_x to 4 and padding_y to 2
- **THEN** the article reader renders content with 4 spaces of horizontal padding and 2 spaces of vertical padding inside the border

### Requirement: Theme selection

The system SHALL support light, dark, and auto themes configurable via the TOML config file. The auto theme SHALL detect the terminal's preferred color scheme at startup. Default theme SHALL be auto.

#### Scenario: Dark theme

- **WHEN** the config file sets theme to "dark"
- **THEN** the UI renders with dark-mode color palette

#### Scenario: Auto theme detects terminal preference

- **WHEN** the config file sets theme to "auto" and the terminal reports a dark preference
- **THEN** the UI renders with the dark-mode color palette

### Requirement: ASCII border fallback

The system SHALL support a plain-ASCII border fallback mode. It SHALL be enabled
by the `ascii` option in the TOML config file, and the system SHALL also
automatically enable ASCII fallback if the terminal does not support Unicode
box-drawing characters. Additionally, the system SHALL accept a `-a` /
`--ascii` command-line flag that forces ASCII fallback mode, overriding both the
config setting and terminal detection for the duration of that run. When ASCII
fallback is in effect, box-drawing characters SHALL be replaced with ASCII
equivalents.

#### Scenario: ASCII fallback enabled in config

- **WHEN** the config file sets ascii to true
- **THEN** the article reader border uses ASCII characters (`-`, `|`, `+`) instead of Unicode box-drawing glyphs

#### Scenario: ASCII fallback auto-detected

- **WHEN** the terminal does not support Unicode box-drawing characters and ascii is not explicitly configured
- **THEN** the system automatically enables ASCII fallback for border rendering

#### Scenario: ASCII flag forces fallback and overrides config

- **WHEN** the user runs `yerss -a` (or `yerss --ascii`) and the config sets ascii to false and the terminal reports Unicode support
- **THEN** the system renders with ASCII fallback glyphs for that run, overriding both the config and terminal detection

#### Scenario: Help flag lists the ascii flag

- **WHEN** the user runs `yerss -h` or `yerss --help`
- **THEN** the usage lists `-a`/`--ascii` and exits 0 without entering the TUI

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

### Requirement: Config file self-update on schema changes

The system SHALL maintain a config schema version (a monotonically increasing
integer) and a registry of options introduced per schema version. When loading an
existing `config.toml` whose recorded header schema version is older than the
current one, the system SHALL, for each option introduced in a later version,
append that option to the file as a commented-out line beneath a "new in vN"
banner — but only if that option's key does not already appear anywhere in the
file (whether commented or live). After appending, the system SHALL update the
file's header schema-version comment to the current version. This self-update
SHALL be best-effort: any read, parse, or write failure SHALL be non-fatal and
SHALL NOT prevent the application from starting with built-in defaults.

#### Scenario: Newly-added options are appended on upgrade

- **WHEN** the application loads a `config.toml` whose header schema version is 1 and the current schema version is 2, which introduced a new option `startup_timeout`
- **THEN** the system appends `# startup_timeout = "30s"` (beneath a "new in v2" banner) to the file and updates the header schema version to 2, provided `startup_timeout` does not already appear in the file

#### Scenario: Append is idempotent across runs

- **WHEN** the application loads a `config.toml` whose header schema version already equals the current version
- **THEN** the system does not append any options and does not rewrite the file

#### Scenario: Manually-uncommented option is not re-appended

- **WHEN** an option introduced in a later schema version already appears as a live (uncommented) key in the file
- **THEN** the system does not append a commented copy of that option

#### Scenario: Self-update failure does not block startup

- **WHEN** the self-update step fails to read, parse, or write the config file
- **THEN** the application proceeds to start using built-in defaults without aborting

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
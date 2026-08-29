## Purpose

Provides TOML-based configuration for keybindings, display preferences (padding, theme, ASCII fallback), and file path overrides, with sensible defaults that require no configuration file to start.

## Requirements

### Requirement: TOML config file location

The system SHALL load configuration from a TOML file at `$XDG_CONFIG_HOME/yerss/config.toml` (fallback `~/.config/yerss/config.toml`). The path SHALL be overridable via a command-line flag. If the config file does not exist, the system SHALL start with built-in defaults.

#### Scenario: Config loaded from default path

- **WHEN** the application starts and the config file exists at the default path
- **THEN** the system loads configuration from that file

#### Scenario: Missing config uses defaults

- **WHEN** the application starts and no config file exists
- **THEN** the system starts with built-in default settings without error

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time.

#### Scenario: Custom keybinding

- **WHEN** the config file maps the "quit" action to `["ctrl+x"]`
- **THEN** pressing Ctrl-X quits the application instead of the default `q`

#### Scenario: Conflicting keybindings rejected

- **WHEN** the config file binds the same key to two different actions
- **THEN** the system reports an error at load time and does not start

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

The system SHALL support a plain-ASCII border fallback mode configurable via the TOML config file. When enabled, box-drawing characters SHALL be replaced with ASCII equivalents. The system SHALL also automatically enable ASCII fallback if the terminal does not support Unicode box-drawing characters.

#### Scenario: ASCII fallback enabled in config

- **WHEN** the config file sets ascii to true
- **THEN** the article reader border uses ASCII characters (`-`, `|`, `+`) instead of Unicode box-drawing glyphs

#### Scenario: ASCII fallback auto-detected

- **WHEN** the terminal does not support Unicode box-drawing characters and ascii is not explicitly configured
- **THEN** the system automatically enables ASCII fallback for border rendering

### Requirement: First-run config bootstrap

On first run — when no `config.toml` exists at the resolved config path — the
system SHALL create the config directory and seed a default `config.toml`. The
seeded file SHALL contain every supported option as a commented-out line, grouped
by section, and SHALL begin with a header comment recording the current config
schema version and the build commit (falling back to `"dev"` when no commit is
embedded at build time). The system SHALL also seed an empty `feeds.txt` in the
same directory with a comment header stating that one RSS URL goes per line.
Files that already exist SHALL NOT be overwritten or truncated; the bootstrap
SHALL only create files that are absent.

#### Scenario: First run seeds config directory and files

- **WHEN** the application starts and neither `config.toml` nor `feeds.txt` exist in the resolved config directory
- **THEN** the system creates the config directory, writes a `config.toml` whose header records the current schema version and build commit and whose body contains every supported option as commented-out lines, and writes a `feeds.txt` containing only a comment header

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
## Purpose

Provides TOML-based configuration for keybindings, display preferences (padding, theme, ASCII fallback), and file path overrides, with sensible defaults that require no configuration file to start.

## ADDED Requirements

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

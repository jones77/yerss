## MODIFIED Requirements

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

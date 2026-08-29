## MODIFIED Requirements

### Requirement: Feed URL source

The system SHALL read feed URLs from a plain-text file with one URL per line.
Blank lines and lines whose first non-whitespace character is `#` SHALL be
ignored, so the file may contain comments. The default path SHALL be
`$XDG_CONFIG_HOME/yerss/feeds.txt` (fallback `~/.config/yerss/feeds.txt`). The
path SHALL be configurable via the TOML config file.

#### Scenario: Feeds loaded from default path

- **WHEN** the application starts and no custom feeds path is configured
- **THEN** the system reads feed URLs from `$XDG_CONFIG_HOME/yerss/feeds.txt` or `~/.config/yerss/feeds.txt` as fallback

#### Scenario: Custom feeds path

- **WHEN** the TOML config specifies a custom feeds file path
- **THEN** the system reads feed URLs from that path instead of the default

#### Scenario: Comments and blank lines ignored

- **WHEN** the feeds file contains blank lines and lines beginning with `#` interspersed with URL lines
- **THEN** the system skips the blank and comment lines and loads only the URL lines

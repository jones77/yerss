## MODIFIED Requirements

### Requirement: ASCII border fallback

The system SHALL support a plain-ASCII fallback mode. It SHALL be enabled by the `ascii` option in the TOML config file, and the system SHALL also automatically enable ASCII fallback if the terminal does not support Unicode box-drawing characters. Additionally, the system SHALL accept a `-a` / `--ascii` command-line flag that forces ASCII fallback mode, overriding both the config setting and terminal detection for the duration of that run. When ASCII fallback is in effect, box-drawing characters SHALL be replaced with ASCII equivalents and article lead image rendering SHALL be disabled (halfblock image glyphs are Unicode and incompatible with ASCII-only mode); articles SHALL render with `[alt]`/`[image]` placeholders for inline content images and no lead image block.

#### Scenario: ASCII fallback enabled in config

- **WHEN** the config file sets ascii to true
- **THEN** the article reader border uses ASCII characters (`-`, `|`, `+`) instead of Unicode box-drawing glyphs and lead images are not rendered

#### Scenario: ASCII fallback auto-detected

- **WHEN** the terminal does not support Unicode box-drawing characters and ascii is not explicitly configured
- **THEN** the system automatically enables ASCII fallback for border rendering and disables lead image rendering

#### Scenario: ASCII flag forces fallback and overrides config

- **WHEN** the user runs `yerss -a` (or `yerss --ascii`) and the config sets ascii to false and the terminal reports Unicode support
- **THEN** the system renders with ASCII fallback glyphs and no lead images for that run, overriding both the config and terminal detection

#### Scenario: Help flag lists the ascii flag

- **WHEN** the user runs `yerss -h` or `yerss --help`
- **THEN** the usage lists `-a`/`--ascii` and exits 0 without entering the TUI

## ADDED Requirements

### Requirement: Image rendering configuration

The system SHALL support an `images` option under the `[display]` section of the TOML config file with values `auto`, `on`, or `off`. The default SHALL be `auto`. When set to `auto`, the system SHALL render article lead images when an image URL is present and the terminal supports the rendering path. When set to `on`, the system SHALL always attempt to render lead images. When set to `off`, the system SHALL never render lead images. The `-a` / `--ascii` flag SHALL force image rendering off regardless of the config value, taking precedence over both `auto` and `on`.

#### Scenario: Default auto renders images

- **WHEN** the config does not specify `images` and an article with a lead image URL is opened in a Unicode-capable terminal
- **THEN** the lead image is rendered inline

#### Scenario: Explicit off disables images

- **WHEN** the config sets `images = "off"` and an article with a lead image URL is opened
- **THEN** no lead image block is rendered

#### Scenario: ASCII flag overrides images config

- **WHEN** the config sets `images = "on"` and the user runs `yerss -a`
- **THEN** no lead image is rendered for that run despite the config setting

#### Scenario: Invalid images value falls back to default

- **WHEN** the config sets `images` to an unrecognized value
- **THEN** the system treats it as `auto` and renders images when supported

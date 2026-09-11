## ADDED Requirements

### Requirement: Configurable seconds display

The system SHALL support a `show_seconds` boolean option under the `[display]`
section of the TOML config file. The default SHALL be `false`. When
`show_seconds` is `false`, every displayed publication or refresh time SHALL
render at minute precision (`HH:MM`); when `true`, they SHALL render at second
precision (`HH:MM:SS`). The option SHALL apply to both the list view and the
article view.

#### Scenario: Seconds hidden by default

- **WHEN** the config does not specify `show_seconds` and the list or article view displays a time
- **THEN** the time renders at minute precision (`HH:MM`) without seconds

#### Scenario: Seconds shown when enabled

- **WHEN** the config sets `show_seconds = true` and the list or article view displays a time
- **THEN** the time renders at second precision (`HH:MM:SS`)

#### Scenario: Invalid value falls back to minute precision

- **WHEN** the config sets `show_seconds` to an unrecognized value
- **THEN** the system treats it as `false` and renders times at minute precision
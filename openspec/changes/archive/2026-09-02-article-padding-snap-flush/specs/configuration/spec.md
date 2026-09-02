# Delta: configuration (article-padding-snap-flush)

## MODIFIED Requirements

### Requirement: Configurable article padding

The system SHALL allow the horizontal and vertical padding inside the article reader to be configured via the TOML config file. Default horizontal padding SHALL be 2 spaces and default vertical padding SHALL be 1 space. The horizontal padding SHALL be applied on each side of the content inside the border. The vertical padding SHALL be applied below the article content as trailing blank content rows: the article's scrollable content SHALL end with that many blank rows, so at the article's end the final text line sits above that many blank rows before the bottom border. The article view SHALL NOT reserve a fixed strip of vertical padding below the content viewport; the content SHALL fill the interior between the top and bottom borders so image captions can reach the viewport's last line. The two options SHALL be independent: setting one SHALL NOT affect the other.

#### Scenario: Custom horizontal padding

- **WHEN** the config file sets padding_x to 4
- **THEN** the article reader renders content with 4 spaces of horizontal padding inside the border and the default 1 space of trailing vertical padding

#### Scenario: Custom vertical padding

- **WHEN** the config file sets padding_y to 2
- **THEN** the article's scrollable content ends with 2 blank rows, and horizontal padding stays at its default 2 spaces

#### Scenario: Independent overrides

- **WHEN** the config file sets padding_x to 0 and padding_y to 3
- **THEN** content renders with no horizontal padding and 3 trailing blank rows, each option applied independently
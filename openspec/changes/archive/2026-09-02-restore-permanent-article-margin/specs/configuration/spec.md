# Delta: configuration (restore-permanent-article-margin)

## MODIFIED Requirements

### Requirement: Configurable article padding

The system SHALL allow the horizontal and vertical padding inside the article reader border to be configured via the TOML config file. Default horizontal padding SHALL be 2 spaces and default vertical padding SHALL be 1 space. The vertical padding SHALL be applied as a permanent margin between the content viewport and both the top and bottom borders: the content SHALL sit one vertical-padding row below the top border and one vertical-padding row above the bottom border.

#### Scenario: Custom padding

- **WHEN** the config file sets padding_x to 4 and padding_y to 2
- **THEN** the article reader renders content with 4 spaces of horizontal padding and a 2-row vertical margin above and below the content inside the border

#### Scenario: Independent overrides

- **WHEN** the config file sets padding_x to 0 and padding_y to 3
- **THEN** content renders with no horizontal padding and a 3-row permanent margin above and below the viewport, each option applied independently
## ADDED Requirements

### Requirement: Render safety under degenerate dimensions

The list view and article view SHALL render without panicking when the terminal
width or height is zero or smaller than the rendered content. The model SHALL
initialize with non-zero default width and height so that the first frame renders
before any `WindowSizeMsg` is received. No render function SHALL index a slice
using an index derived from terminal dimensions without first guaranteeing the
index is in range. When the available height is too small to display both
content and the status bar, the system SHALL prioritize not crashing over
displaying every element.

#### Scenario: List view renders before WindowSizeMsg

- **WHEN** `View()` is called before any `WindowSizeMsg` has been processed, so the model width and height are still their default non-zero values
- **THEN** the list view renders a string without panicking

#### Scenario: List view renders at height zero

- **WHEN** the model height is set to 0 and `renderList` is called
- **THEN** the function returns a string without panicking (no `index out of range`)

#### Scenario: List view renders at height one

- **WHEN** the model height is set to 1 and the article list is non-empty
- **THEN** the function returns a single-line string without panicking

#### Scenario: Article view renders at zero height

- **WHEN** the model height is set to 0 and `renderArticle` is called
- **THEN** the function returns a string without panicking

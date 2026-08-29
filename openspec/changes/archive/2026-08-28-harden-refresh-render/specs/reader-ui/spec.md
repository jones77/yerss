## ADDED Requirements

### Requirement: Popup overlay composition

When a popup (tag popup or help popup) is displayed over the list or article view,
the system SHALL compose the popup over the base view by centering the popup within
the base's visible area. The horizontal and vertical centering SHALL be computed
from the **display width** of the base and popup lines, with ANSI styling escapes
excluded from the measurement, so that escape-sequence runes are not counted as
visible columns. The system SHALL preserve the base view's content outside the
region covered by the popup. Popup placement SHALL be rune-width aware so that
wide glyphs (such as CJK characters) do not misalign the centered position.

#### Scenario: Popup is centered over the base

- **WHEN** a tag popup or help popup is rendered over a base view whose visible width is greater than the popup's visible width
- **THEN** the popup is horizontally centered within the base view's visible width, not glued to the left edge

#### Scenario: Base content is preserved outside the popup

- **WHEN** a popup is composed over a base view
- **THEN** the base view's content in regions not covered by the popup remains intact and is not overwritten by the popup's styling

#### Scenario: ANSI escapes are not counted as width

- **WHEN** the base or popup lines contain ANSI color escape sequences
- **THEN** the centering calculation uses the visible display width (escapes excluded), so the popup is positioned by its visible extent rather than its raw rune count

## MODIFIED Requirements

### Requirement: Render safety under degenerate dimensions

The list view and article view SHALL render without panicking when the terminal
width or height is zero or smaller than the rendered content. The model SHALL
initialize with non-zero default width and height so that the first frame renders
before any `WindowSizeMsg` is received. No render function SHALL index a slice
using an index derived from terminal dimensions without first guaranteeing the
index is in range. No render function SHALL emit more lines than the available
terminal height; when an interior region (such as the article viewport) is floored
at its minimum height under a degenerate terminal combined with non-zero padding,
the rendered frame SHALL be clamped to the available height. When the available
height is too small to display both content and the status bar, the system SHALL
prioritize not crashing over displaying every element.

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

#### Scenario: Article border does not exceed terminal height at degenerate size

- **WHEN** the article view is rendered with a very small terminal height and non-zero vertical padding that floors the viewport height at its minimum
- **THEN** the rendered border frame produces at most as many lines as the available terminal height, never more

## MODIFIED Requirements

### Requirement: Wrapped link URLs are fully hidden in compact mode

The system SHALL hide the entire visible URL of a body hyperlink in compact mode
even when the URL is longer than the content width and wraps across multiple
rendered lines. When a link's URL wraps, the compact rendering SHALL show only
the link's name and SHALL NOT render any part of the URL text — neither the first
line nor any wrapped continuation line. A link name that itself wraps across
lines SHALL render in full. The link name SHALL remain wrapped in OSC 8 escape
sequences carrying the full URL, so the link stays clickable. Hiding a wrapped
URL SHALL NOT leave a short, unpadded line in the surrounding paragraph: the
paragraph SHALL re-wrap to the content width so every line fills the full width
and no visible gap appears where the URL text was removed.

#### Scenario: Long URL wrapping across lines is fully hidden

- **WHEN** an article body contains a link whose URL is longer than the content width so the URL wraps across multiple rendered lines, and the reader is in compact mode
- **THEN** the rendered text shows only the link name, and no part of the URL text — including the wrapped continuation lines — is visible

#### Scenario: Wrapped link name is preserved

- **WHEN** a compact-mode link's name is long enough to wrap across multiple rendered lines
- **THEN** the full link name renders and the URL remains hidden

#### Scenario: Adjacent links to the same URL each hide their href

- **WHEN** an article body contains two links to the same URL in compact mode
- **THEN** both link names render and neither URL text is visible

#### Scenario: Paragraph re-wraps after a wrapped URL is hidden

- **WHEN** a compact-mode link's URL is long enough to wrap across lines, so hiding the URL would otherwise leave the surrounding paragraph line shorter than the content width
- **THEN** the paragraph re-wraps to the full content width and no line is left visibly short with a trailing blank gap
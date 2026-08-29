## ADDED Requirements

### Requirement: Article top border title truncation

The system SHALL render the article's date and title inline with the top border
with the date always fully visible and the title left-aligned after the date.
When the title does not fit the available width, the system SHALL truncate the
title and end it with an ellipsis glyph — `…` in Unicode mode and `...` in ASCII
fallback mode — and SHALL place a single horizontal dash adjacent to the right
corner glyph so the right end mirrors the left (`┌─` on the left, `─╖` on the
right). When the title fits with room to spare, the system SHALL fill the
leftover space with horizontal dashes as before and SHALL NOT append an
ellipsis.

#### Scenario: Title fits is dash-filled

- **WHEN** the article title fits the top border with room to spare
- **THEN** the top border renders the full date and title followed by fill dashes and the right corner, with no ellipsis

#### Scenario: Long title is truncated with an ellipsis

- **WHEN** the article title does not fit the top border
- **THEN** the date remains fully visible, the title is truncated and ends with `…`, and a single dash precedes the right corner (for example `┌─ 2026-01-02 · Hello world artic… ─╖`)

#### Scenario: ASCII fallback ellipsis

- **WHEN** ASCII fallback mode is enabled and the article title does not fit
- **THEN** the truncated title ends with `...` instead of `…`

#### Scenario: Date stays visible for a very long title

- **WHEN** the article title is longer than the entire top-border content width
- **THEN** the date is still rendered in full and only the title is truncated with an ellipsis

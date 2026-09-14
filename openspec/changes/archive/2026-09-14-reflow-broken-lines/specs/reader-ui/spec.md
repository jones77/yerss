## ADDED Requirements

### Requirement: Blockquote closing punctuation stays on the quote bar

The system SHALL render every line of a blockquote with the quote marker,
including the line that holds the blockquote's closing punctuation when the
quoted text's final word wraps. When a blockquote's last word ends in closing
punctuation (such as a closing curly quote) and the word wraps across lines, the
closing punctuation SHALL render on a new line that still carries the quote
marker, and SHALL NOT be stranded as a bare line without the marker.

#### Scenario: Closing curly quote wraps onto a quote-bar line

- **WHEN** a blockquote's quoted text is long enough that its final word ending in a closing curly quote (e.g. `on.”`) does not fit on the blockquote's last full line
- **THEN** the closing curly quote renders on a new line that still shows the quote marker, and no bare punctuation-only line appears without the marker
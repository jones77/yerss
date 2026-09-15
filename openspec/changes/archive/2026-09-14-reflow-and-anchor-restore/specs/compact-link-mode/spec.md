## MODIFIED Requirements

### Requirement: Expanded link display mode

The system SHALL provide an expanded link display mode in the article reader in
which each body hyperlink renders the styled link name followed by the full URL
(`text url`), preserving the visible URL for readers who want to inspect link
targets. The URL text SHALL be wrapped in OSC 8 sequences carrying the same URL
as the name. Expanded mode SHALL NOT alter clickability: both the name and the
URL open the link. When a URL does not fit in the space remaining on the link
name's line, the surrounding paragraph SHALL re-flow so the URL starts on the
name's line and wraps at the content width instead of being pushed onto its own
line below a half-empty name line, and every wrapped line of the URL SHALL
remain wrapped in OSC 8 sequences carrying the full URL.

#### Scenario: Expanded link shows name and URL

- **WHEN** the reader is in expanded mode and an article body contains `<a href="https://example.com/post">world</a>`
- **THEN** the rendered text shows `world https://example.com/post` and both are clickable

#### Scenario: URL shown is the link target

- **WHEN** the reader is in expanded mode and a link renders with its URL text
- **THEN** the displayed URL text matches the link's actual target URL

#### Scenario: Long URL re-flows onto the name's line

- **WHEN** the reader is in expanded mode and a link's URL is longer than the
  space remaining on the link name's line, so the un-re-flowed render would
  push the URL onto its own line
- **THEN** the paragraph re-flows so the URL starts on the name's line and
  wraps at the content width, and no half-empty name line appears above a bare
  URL line

#### Scenario: Each wrapped URL line stays clickable

- **WHEN** the reader is in expanded mode and a re-flowed URL wraps across
  multiple lines
- **THEN** every wrapped line of the URL is wrapped in OSC 8 sequences carrying
  the full URL, so clicking any wrapped segment opens the link

## ADDED Requirements

### Requirement: Hiding a wrapped URL leaves no whitespace-only residual line

The system SHALL hide the entire visible URL of a body hyperlink in compact
mode without leaving a whitespace-only residual line behind. When a wrapped
URL's closing bracket lands on a continuation line and nothing but styled
whitespace (the fill padding and the wrap newline) separates it from the end of
that line, hiding the URL SHALL also remove that trailing whitespace, so the
paragraph re-wraps as a single flow with no blank or whitespace-only line
splitting it where the wrapped href ended.

#### Scenario: No whitespace-only line after hiding a wrapped href

- **WHEN** compact mode hides a URL that wraps across lines, so the wrapped
  href's close is followed only by fill padding and a wrap newline
- **THEN** the paragraph re-wraps as one continuous flow and no blank or
  whitespace-only line appears where the wrapped href ended

#### Scenario: Text after the wrapped href continues on the same flow

- **WHEN** compact mode hides a wrapped URL and the sentence continues after
  the link
- **THEN** the continuing text re-wraps into the same paragraph flow, with no
  blank line separating it from the paragraph that held the link
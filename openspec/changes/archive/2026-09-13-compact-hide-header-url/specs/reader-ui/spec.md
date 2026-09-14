## MODIFIED Requirements

### Requirement: Clickable article URL in reader header

The reader view SHALL open, in expanded link mode, with the article's own URL as
the very first content line, followed by a blank line, the article title rendered
bold with the author line (`by <author>`) directly beneath it when the article
has an author. In compact link mode the reader view SHALL NOT render the header
URL line; the article title SHALL instead be the first content line, with the
author line directly beneath it when the article has an author. In expanded mode
the URL SHALL be rendered exactly once — the system SHALL NOT repeat it as
trailing href text after the link — and SHALL be wrapped in an OSC 8 terminal
hyperlink, so that Cmd/Ctrl-clicking it opens the URL in the system browser,
mirroring the `o` key behavior. In expanded mode the URL line SHALL render in the
status-bar blue role and SHALL wrap at the content width so a long URL folds onto
subsequent lines rather than overflowing the viewport. This single-render rule
applies only to the header URL; hyperlinks in the article body SHALL render
according to the compact or expanded link mode.

#### Scenario: Header URL is the first line and rendered once in expanded mode

- **WHEN** an article with a link URL is displayed in the reader view in expanded link mode
- **THEN** the first content line is the article URL, and the URL text appears exactly once on that line

#### Scenario: Header URL is clickable in expanded mode

- **WHEN** an article with a link URL is displayed in expanded mode and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser

#### Scenario: Header URL renders in the blue role in expanded mode

- **WHEN** an article with a link URL is displayed in expanded mode
- **THEN** the URL line renders in the status-bar blue role, matching the status bar and title

#### Scenario: Header URL wraps at the content width in expanded mode

- **WHEN** an article URL is longer than the content width and is displayed in expanded mode
- **THEN** the URL line folds onto subsequent lines within the content width instead of overflowing the viewport

#### Scenario: Header order is URL, blank line, bold title, author in expanded mode

- **WHEN** an article with a title and author is displayed in expanded mode
- **THEN** the URL line is followed by a blank line, then the title rendered bold, then the `by <author>` line directly beneath the title with no blank line between them

#### Scenario: Header order without author in expanded mode

- **WHEN** an article has a title but no author and is displayed in expanded mode
- **THEN** the header shows the URL line, a blank line, and the bold title with no author line

#### Scenario: Header URL is hidden in compact mode

- **WHEN** an article with a link URL is displayed in compact link mode
- **THEN** the header URL line is not rendered and the article title is the first content line

#### Scenario: Compact header starts with the title then author

- **WHEN** an article with a title and author is displayed in compact mode
- **THEN** the title renders as the first content line, with the `by <author>` line directly beneath it, and no URL line appears

#### Scenario: Body links follow the link display mode

- **WHEN** an article body contains `<a href="https://example.com/post">world</a>` in compact mode
- **THEN** the body renders the styled link name `world` in the status-bar blue role without the URL; in expanded mode it renders `world https://example.com/post`

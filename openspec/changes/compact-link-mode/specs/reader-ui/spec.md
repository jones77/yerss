## MODIFIED Requirements

### Requirement: Clickable links in article content

The system SHALL render hyperlinks in article HTML content as clickable OSC 8
terminal hyperlinks. The rendered link SHALL display the styled link text
wrapped in OSC 8 escape sequences carrying the link's URL, so that
Cmd/Ctrl-clicking it in a supporting terminal opens the URL in the system
browser, and the application's left-click hit-testing opens it as well. By
default the link SHALL render in compact form — the styled link name only, in
the status-bar blue role, without the visible URL text — so a readable,
uncluttered fallback exists for terminals that do not support OSC 8 while the
click target is preserved. An expanded mode SHALL render the link name followed
by the visible URL (`text url`) in the same blue role. OSC 8 hyperlinks SHALL
function in ASCII fallback mode, since OSC 8 is an escape sequence and not a
Unicode glyph.

#### Scenario: Inline link rendered as clickable styled text without the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view in compact mode
- **THEN** the link text "world" is styled in the status-bar blue role and wrapped in OSC 8 escape sequences carrying the URL `https://example.com/post`, and the rendered text shows only `world`

#### Scenario: Cmd/Ctrl-click on link opens browser

- **WHEN** the user Cmd/Ctrl-clicks on the link text of an OSC 8 hyperlink in the article reader view in a supporting terminal
- **THEN** the terminal opens the link URL in the system browser

#### Scenario: Left-click on link opens browser

- **WHEN** the user left-clicks the rendered link name in the article reader view
- **THEN** the application opens the link URL in the system browser without starting a text selection

#### Scenario: Link URL preserved in expanded mode

- **WHEN** the reader is in expanded mode and an article with a link is displayed
- **THEN** the link text and the URL (`text url`) are displayed, identical to the pre-compact behavior

#### Scenario: OSC 8 links work in ASCII mode

- **WHEN** ASCII fallback mode is enabled and an article with links is displayed
- **THEN** the link text is wrapped in OSC 8 escape sequences and the markdown links are clickable in supporting terminals

### Requirement: Clickable article URL in reader header

The reader view SHALL open with the article's own URL as the very first content
line, followed by a blank line, the article title rendered bold with the author
line (`by <author>`) directly beneath it when the article has an author. The
URL SHALL be rendered exactly once — the system SHALL NOT repeat it as trailing
href text after the link — and SHALL be wrapped in an OSC 8 terminal hyperlink,
so that Cmd/Ctrl-clicking it opens the URL in the system browser, mirroring the
`o` key behavior. The URL line SHALL render in the status-bar blue role and
SHALL wrap at the content width so a long URL folds onto subsequent lines
rather than overflowing the viewport. This single-render rule applies only to
the header URL; hyperlinks in the article body SHALL render according to the
compact or expanded link mode.

#### Scenario: Header URL is the first line and rendered once

- **WHEN** an article with a link URL is displayed in the reader view
- **THEN** the first content line is the article URL, and the URL text appears exactly once on that line

#### Scenario: Header URL is clickable

- **WHEN** an article with a link URL is displayed in the reader view and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser

#### Scenario: Header URL renders in the blue role

- **WHEN** an article with a link URL is displayed in the reader view
- **THEN** the URL line renders in the status-bar blue role, matching the status bar and title

#### Scenario: Header URL wraps at the content width

- **WHEN** an article URL is longer than the content width
- **THEN** the URL line folds onto subsequent lines within the content width instead of overflowing the viewport

#### Scenario: Header order is URL, blank line, bold title, author

- **WHEN** an article with a title and author is displayed in the reader view
- **THEN** the URL line is followed by a blank line, then the title rendered bold, then the `by <author>` line directly beneath the title with no blank line between them

#### Scenario: Header order without author

- **WHEN** an article has a title but no author
- **THEN** the header shows the URL line, a blank line, and the bold title with no author line

#### Scenario: Body links follow the link display mode

- **WHEN** an article body contains `<a href="https://example.com/post">world</a>` in compact mode
- **THEN** the body renders the styled link name `world` in the status-bar blue role without the URL; in expanded mode it renders `world https://example.com/post`
## MODIFIED Requirements

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation
(PgUp/PgDn, Ctrl-F/Ctrl-B) SHALL scroll by a full viewport page. Half-page
navigation (Ctrl-D/Ctrl-U, Space) SHALL scroll by half a viewport. `g` or
Ctrl-Up SHALL move to the top; `G` or Ctrl-Down SHALL move to the bottom. In the
list view, up/down moves selection across visible rows (day headers and expanded
articles); in the article view, up/down scrolls the article by one line. `Enter`,
`l`, or `o` opens the selected article from the list; `Esc`, `Enter`, `h`, or
`b` returns from the article view to the list, reselecting the previously viewed
article. In the list view, `Esc`, `h`, and `b` clear an active tag filter or are
no-ops when no filter is active.

When the list cursor is on a day header, fold keys SHALL act on that day:
`h` SHALL collapse it, `l` SHALL expand it, and `Enter`, Space, or Tab SHALL
toggle its expansion; the article-open behavior of `l`/Enter/`o` SHALL NOT apply
on a header row. When the cursor is on an article row, `Tab` SHALL toggle the
expansion of that article's day group, and the existing bindings (`h`/`b` clear
the filter or are a no-op, `l`/Enter/`o` open the article, Space half-pages)
SHALL apply unchanged.

#### Scenario: Open article from list

- **WHEN** the user presses Enter, `l`, or `o` on a selected article row in the list view
- **THEN** the article reader view opens displaying that article

#### Scenario: Fold keys on a day header

- **WHEN** the list cursor is on a day header and the user presses `h`, `l`, Enter, Space, or Tab
- **THEN** `h` collapses that day, `l` expands it, and Enter/Space/Tab toggle its expansion; no article is opened

#### Scenario: Tab toggles the current day group from an article row

- **WHEN** the list cursor is on an article row and the user presses Tab
- **THEN** the day group containing that article toggles its expansion

#### Scenario: Return to list from article

- **WHEN** the user presses Esc, Enter, `h`, or `b` in the article view
- **THEN** the list view is displayed with the previously viewed article selected

#### Scenario: Clear tag filter from list

- **WHEN** a tag filter is active on the list view and the user presses Esc, `h`, `b`, or left arrow on an article row
- **THEN** the tag filter is removed and all articles are shown

#### Scenario: Esc on unfiltered list is a no-op

- **WHEN** no tag filter is active on the list view and the user presses Esc, `h`, or `b` on an article row
- **THEN** nothing happens

## ADDED Requirements

### Requirement: Mouse navigation

The system SHALL enable mouse cell-motion reporting and handle mouse events in
both the list and article views. In the list view, a single left-click on an
article row SHALL set the cursor to that row and open the article. A left-click
on a day-header row SHALL toggle that day group's collapse/expand state. A click
on the status bar (the bottom line) SHALL be ignored. The mouse wheel SHALL move
the list cursor: wheel down moves the cursor down by one row, wheel up moves it
up by one row. In the article view, the mouse wheel SHALL scroll the article
viewport: wheel down scrolls down by one line, wheel up scrolls up by one line.
Mouse clicks in the article view SHALL NOT be interpreted by the application —
clickable links are handled natively by the terminal via OSC 8 hyperlinks.

#### Scenario: Click opens article from list

- **WHEN** the user single-clicks on an article row in the list view
- **THEN** the cursor moves to that row and the article reader view opens displaying that article

#### Scenario: Click on day header toggles collapse

- **WHEN** the user single-clicks on a day-header row in the list view
- **THEN** that day group toggles between collapsed and expanded

#### Scenario: Click on status bar is ignored

- **WHEN** the user clicks on the bottom line of the list view (the status bar)
- **THEN** nothing happens

#### Scenario: Wheel scrolls list cursor

- **WHEN** the user scrolls the mouse wheel down in the list view
- **THEN** the list cursor moves down by one row

- **WHEN** the user scrolls the mouse wheel up in the list view
- **THEN** the list cursor moves up by one row

#### Scenario: Wheel scrolls article viewport

- **WHEN** the user scrolls the mouse wheel down in the article view
- **THEN** the article viewport scrolls down by one line

- **WHEN** the user scrolls the mouse wheel up in the article view
- **THEN** the article viewport scrolls up by one line

### Requirement: Clickable links in article content

The system SHALL render hyperlinks in article HTML content as clickable OSC 8
terminal hyperlinks. The link text SHALL be wrapped in OSC 8 escape sequences
that carry the link's URL, so that Cmd/Ctrl-clicking the link text in a
supporting terminal opens the URL in the system browser. The link URL SHALL
also be displayed as plain text alongside the clickable link text, preserving
the existing behavior for terminals that do not support OSC 8. OSC 8 hyperlinks
SHALL function in ASCII fallback mode, since OSC 8 is an escape sequence and not
a Unicode glyph.

#### Scenario: Inline link rendered as OSC 8 hyperlink

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the link text "world" is wrapped in OSC 8 escape sequences carrying the URL `https://example.com/post`, and the URL is also displayed as plain text

#### Scenario: Cmd/Ctrl-click on link opens browser

- **WHEN** the user Cmd/Ctrl-clicks on an OSC 8 hyperlink in the article reader view in a supporting terminal
- **THEN** the terminal opens the link URL in the system browser

#### Scenario: Link URL preserved as fallback text

- **WHEN** an article with a link is displayed in a terminal that does not support OSC 8
- **THEN** the link text and URL are displayed as plain text, identical to the behavior without OSC 8

#### Scenario: OSC 8 links work in ASCII mode

- **WHEN** ASCII fallback mode is enabled and an article with links is displayed
- **THEN** the link text is wrapped in OSC 8 escape sequences and the links are clickable in supporting terminals

### Requirement: Clickable article URL in reader header

The system SHALL render the article's own URL (displayed in the reader view
header) as a clickable OSC 8 terminal hyperlink, so that Cmd/Ctrl-clicking it
opens the URL in the system browser. This mirrors the existing `o` key behavior
for opening the article URL.

#### Scenario: Article header URL is clickable

- **WHEN** an article with a link URL is displayed in the reader view and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser

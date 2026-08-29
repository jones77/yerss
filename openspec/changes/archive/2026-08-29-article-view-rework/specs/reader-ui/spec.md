## MODIFIED Requirements

### Requirement: Clickable article URL in reader header

The reader view SHALL open with the article's own URL as the very first content
line, followed by a blank line, the article title rendered bold with the author
line (`by <author>`) directly beneath it when the article has an author. The
URL SHALL be rendered exactly once — the system SHALL NOT repeat it as trailing
href text after the link — and SHALL be wrapped in an OSC 8 terminal hyperlink,
so that Cmd/Ctrl-clicking it opens the URL in the system browser, mirroring the
`o` key behavior. The header SHALL be rendered through the same markdown
renderer as the article body so its styling matches the configured display
theme. This single-render rule applies only to the header URL; hyperlinks in
the article body SHALL continue to render as the styled link text followed by
the URL (`text url`).

#### Scenario: Header URL is the first line and rendered once

- **WHEN** an article with a link URL is displayed in the reader view
- **THEN** the first content line is the article URL, and the URL text appears exactly once on that line

#### Scenario: Header URL is clickable

- **WHEN** an article with a link URL is displayed in the reader view and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser

#### Scenario: Header order is URL, blank line, bold title, author

- **WHEN** an article with a title and author is displayed in the reader view
- **THEN** the URL line is followed by a blank line, then the title rendered bold, then the `by <author>` line directly beneath the title with no blank line between them

#### Scenario: Header order without author

- **WHEN** an article has a title but no author
- **THEN** the header shows the URL line, a blank line, and the bold title with no author line

#### Scenario: Body links keep the text-plus-URL rendering

- **WHEN** an article body contains `<a href="https://example.com/post">world</a>`
- **THEN** the body renders the styled link text `world` followed by the URL `https://example.com/post`

### Requirement: Article lead image rendering

When image rendering is enabled and an article has a publisher-attached image
URL (captured from `<enclosure>` or `media:thumbnail` during feed parsing), the
system SHALL render the image inline below the reader header (the URL, title,
and author lines) and above the rendered article content, as a block of content
that scrolls with the article text. The image SHALL be rendered using a Unicode
halfblocks representation (half-block characters with ANSI color) sized to fit
the content width while preserving aspect ratio, and SHALL be horizontally
centered within the content width. Directly beneath the image, with no blank
line between them, the system SHALL render a one-line photo attribution
constrained to the photo's own width (truncated to that width when the credit
is longer) and centered beneath the photo; a blank line SHALL follow the
attribution before the article content resumes. The attribution text SHALL be
the photo's credit extracted from the article's stored HTML when present — a
`figcaption` associated with the figure containing the lead image, otherwise an
element whose class indicates a credit line — and SHALL fall back to
`photo: <source>`, where `<source>` is derived by the same source-identifier
rules used for the list view's organization column. The combined image-plus-
attribution block height SHALL be capped so that the full block plus at least
one line of body text fits within the viewport height; when the block is taller
than this cap, it SHALL be scaled down to fit. When image rendering is disabled
(config `off` or `--ascii` mode), no image block and no attribution SHALL be
rendered and the article SHALL render exactly as it does without an image URL.

#### Scenario: Lead image rendered below the header and above the body

- **WHEN** image rendering is enabled and an article with an enclosure image URL is opened
- **THEN** the image is rendered as a halfblocks block positioned below the URL/title/author header lines and above the body text, sized to the content width with aspect ratio preserved

#### Scenario: Image is horizontally centered

- **WHEN** a rendered lead image occupies fewer columns than the content width
- **THEN** the image block is centered with equal left and right padding within the content area

#### Scenario: Attribution centered directly beneath the image

- **WHEN** a lead image is rendered with an attribution
- **THEN** the attribution line is centered within the photo's own width and there is no blank line between the image and the attribution

#### Scenario: Attribution truncated to the photo width

- **WHEN** the attribution text is longer than the rendered photo's width (for example after the photo is scaled down by the height cap)
- **THEN** the attribution is truncated to the photo's width and stays centered beneath the photo

#### Scenario: Attribution taken from the article HTML

- **WHEN** the article's HTML contains a `figcaption` credit associated with the lead image's figure
- **THEN** the attribution line shows that credit text, centered beneath the photo

#### Scenario: Attribution falls back to the source organization

- **WHEN** the article's HTML contains no credit for the lead image and the article link host is `theintercept.com`
- **THEN** the attribution line shows `photo: intercept`

#### Scenario: Blank line after the attribution

- **WHEN** a lead image with attribution is rendered
- **THEN** exactly one blank line separates the attribution from the first body line

#### Scenario: Lead image scrolls with content

- **WHEN** a rendered lead image is visible and the user scrolls down
- **THEN** the image block scrolls upward with the body text as a single scrolling content region

#### Scenario: Lead image capped to viewport

- **WHEN** an article's enclosure image, when aspect-fit to the content width, would make the combined image-plus-attribution block taller than the available viewport height minus one line
- **THEN** the image is scaled down so that the full block and at least one line of body text are visible without scrolling

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

### Requirement: Atomic image scroll behavior

When a lead image block is rendered inline, the system SHALL ensure the image
block — the image lines plus its attribution line, treated as one unit — is
either fully visible within the viewport or fully scrolled out of view; it
SHALL never display a partially clipped image or attribution. When a downward
scroll would move the viewport offset into the interior of the image block
(the range strictly between the block's first and last line), the system SHALL
snap the offset to the line immediately after the block, causing the image and
attribution to vanish in a single keystroke. When an upward scroll would move
the offset into the image interior, the system SHALL snap the offset to the
block's first line, revealing the full image and its attribution. This
snapping SHALL apply to all scroll operations: line, page, half-page, and
goto-bottom.

#### Scenario: Downward scroll skips past image

- **WHEN** the image block occupies content lines 4 through 10 (image lines 4 through 9 plus its attribution line 10), the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 11 (the line after the block) and the image and attribution are no longer visible

#### Scenario: Upward scroll reveals full image with attribution

- **WHEN** the image block occupies content lines 4 through 10, the viewport offset is 11, and the user scrolls up one line
- **THEN** the viewport offset snaps to 4 (the block's first line) and the full image and its attribution are visible

#### Scenario: Page-down skips past image

- **WHEN** the image block is partially within the viewport after a page-down operation
- **THEN** the offset snaps past the image block so the image and attribution are fully scrolled out

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below the image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally

### Requirement: Mouse navigation

The system SHALL enable mouse cell-motion reporting and handle mouse events in
both the list and article views. In the list view, a single left-click on an
article row SHALL set the cursor to that row and open the article. A left-click
on a day-header row SHALL toggle that day group's collapse/expand state. A click
on the status bar (the bottom line) SHALL be ignored. The mouse wheel SHALL move
the list cursor: wheel down moves the cursor down by one row, wheel up moves it
up by one row. In the article view, the mouse wheel SHALL scroll the article
viewport: wheel down scrolls down by one line, wheel up scrolls up by one line.
In the article view, an unmodified left-button press on the content area SHALL
begin a text selection (see "In-view text selection" below); a left-button press
carrying a Ctrl or Cmd modifier SHALL NOT be interpreted by the application, so
the terminal can handle clickable OSC 8 hyperlinks natively. In the article
view, an unmodified right-button press SHALL return to the list view with the
previously viewed article selected, exactly like the Back keys; a right-button
press carrying a Ctrl/Alt/Cmd modifier SHALL NOT be interpreted.

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

#### Scenario: Modifier-click in article is not interpreted

- **WHEN** the user Cmd/Ctrl-clicks inside the article content area
- **THEN** no text selection begins and the application does not handle the click

#### Scenario: Right-click returns to the list

- **WHEN** the user presses the right mouse button (no modifiers) in the article view
- **THEN** the list view is displayed with the previously viewed article selected

#### Scenario: Right-click does not start a text selection

- **WHEN** the user presses the right mouse button in the article view
- **THEN** no text selection is anchored or highlighted

## ADDED Requirements

### Requirement: Article links popup

The system SHALL provide a links popup in the article view, opened by pressing
`l` or the right arrow key. The popup SHALL list every hyperlink rendered in
the article — the header link and the body links — in document order,
deduplicated by URL. Each row SHALL show the link text followed by the URL
rendered in the dim style and truncated to the popup width. Navigation SHALL
use `j`/`k` and the arrow keys and SHALL wrap around at both ends. Enter SHALL
open the selected URL in the system browser using the same non-shell
single-argument opener used for other URL opens. Esc SHALL close the popup
without navigating. When the article contains no links, the popup SHALL open
with an empty list and Enter SHALL do nothing. While the popup is closed,
Enter SHALL continue to close the article view and return to the list.

#### Scenario: Open links popup from article

- **WHEN** the user presses `l` or the right arrow key in the article view
- **THEN** a popup opens listing the article's links in document order

#### Scenario: Row shows link text and dimmed URL

- **WHEN** the article contains `<a href="https://example.com/post">world</a>` and the links popup is open
- **THEN** the popup shows a row with the text `world` followed by the URL `https://example.com/post` in the dim style, truncated to the popup width

#### Scenario: Header link appears in the popup

- **WHEN** an article with a link URL is open and the links popup is opened
- **THEN** the article's own URL appears as a row in the popup

#### Scenario: Duplicate URLs are listed once

- **WHEN** the same URL is rendered twice in the article (for example a link whose text wraps across lines)
- **THEN** the popup lists that URL exactly once

#### Scenario: Navigation wraps around

- **WHEN** the user presses `j` (or down) while the last link row is selected
- **THEN** selection moves to the first row

#### Scenario: Enter opens the selected link

- **WHEN** the user presses Enter on a selected link row in the popup
- **THEN** the selected URL is opened in the system browser via a non-shell single-argument opener

#### Scenario: Esc closes the popup without navigating

- **WHEN** the user presses Esc while the links popup is open
- **THEN** the popup closes, the article view remains, and no URL is opened

#### Scenario: Empty article shows an empty popup

- **WHEN** the article contains no links and the user opens the links popup
- **THEN** the popup opens with an empty list and Enter does nothing

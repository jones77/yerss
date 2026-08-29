## Purpose

Provides the terminal user interface for browsing, reading, and filtering RSS articles — comprising the article list view, the article reader view with a scroll-progress border, and the tag popup for category-based filtering.

## Requirements

### Requirement: Article list rendering

The system SHALL display articles grouped under a collapsible header for each
day, with the most recent day first and articles within a day sorted reverse
chronologically. The list SHALL be rendered as a continuous tree rail down the
left edge: the first row of the whole visible-row list SHALL be prefixed with
`┌` (ASCII `+`), the last row with `└` (ASCII `+`), and every other row with
`├` (ASCII `+`). Article rows SHALL carry a horizontal dash after the corner
glyph (`├─`, ASCII `+-`); day-header rows SHALL carry a single space. Day
headers SHALL NOT display fold markers. Each day header SHALL display the long
date in the form `Weekday Day-ordinal Month, Year` (for example
`Saturday 20th August, 2025`), computed in the user's local timezone, where the
day is suffixed with an English ordinal (`1st`, `2nd`, `3rd`, `4th`, …). The
header for the current local calendar day SHALL be prefixed with `today, ` and
the header for the previous local calendar day SHALL be prefixed with
`yesterday, `; older days and the `Undated` group SHALL be unprefixed. Each
article row SHALL show the publication time as `HH:MM` in the user's local
timezone (`--:--` when undated) after the tree glyph, followed by the article
title; the title SHALL NOT be preceded or followed by bullet points. The
source-domain identifier SHALL be right-aligned at the content edge and SHALL
be the only field on the right; it SHALL be derived from the article's link
host by stripping any leading `www.`, taking the organization label of the
registrable domain (the public suffix plus one, so multi-part public suffixes
such as `co.uk` are removed whole), and truncating it to at most 12 characters
(for example `newrepublic.com` → `newrepublic`,
`reallylongnewspaperdomainname.net` → `reallylongne`). When the article has no
link, the system SHALL derive the identifier from the feed URL host instead;
hosts that have no registrable domain (single-label hosts, IP literals) fall
back to removing the part after the final dot.
When a title would run into the right-aligned identifier, the system SHALL
leave at least one space between the title and the identifier, and a truncated
title SHALL end with an ellipsis — `…` in Unicode mode and `...` in ASCII
fallback mode. The source identifier and the publication time SHALL always be
rendered in the dim/grey style. Only the title SHALL change with read state:
unread titles SHALL be bold and bright, and read titles SHALL be dim.
Selection SHALL be indicated by a full-row background highlight covering the
tree glyph and the row's content, on day-header rows and article rows alike;
the system SHALL NOT render a cursor-gutter marker. A collapsed day header
SHALL hide its article rows; an expanded day header SHALL show them. Selection
SHALL wrap around when moving past the first or last visible row (headers plus
expanded articles).

#### Scenario: Tree rail bookends the list

- **WHEN** the article list is displayed with more than one row
- **THEN** the first row is prefixed with `┌`, the last row with `└`, and every row between with `├`

#### Scenario: Article rows carry the dash, headers a space

- **WHEN** the list is displayed
- **THEN** each article row is prefixed `├─` (or `┌─`/`└─` at the bookends) and each day header is prefixed `┌`, `├`, or `└` followed by a single space

#### Scenario: Tree glyphs are stable while scrolling

- **WHEN** the visible window scrolls so the list's first and last rows are off screen
- **THEN** the interior rows are all prefixed `├` and the bookend glyphs do not move to the window edges

#### Scenario: ASCII fallback tree glyphs

- **WHEN** ASCII fallback mode is enabled
- **THEN** the tree glyphs render as `+` (corners and tees) and `-` (article-row dash)

#### Scenario: Day header shows the long local date

- **WHEN** a day header is rendered for a date with day-of-month 20 in August 2025 that is neither today nor yesterday
- **THEN** the header displays the long date with weekday, an ordinal day number, month, and year (for example `Wednesday 20th August, 2025`) in the user's local timezone

#### Scenario: Current day is prefixed with today

- **WHEN** a day header is rendered for the current local calendar day
- **THEN** the label is `today, ` followed by the long date (for example `today, Saturday 20th August, 2025`)

#### Scenario: Previous day is prefixed with yesterday

- **WHEN** a day header is rendered for the previous local calendar day
- **THEN** the label is `yesterday, ` followed by the long date (for example `yesterday, Friday 19th August, 2025`)

#### Scenario: Articles grouped under a day header

- **WHEN** the article list is displayed and articles exist on more than one day
- **THEN** each day's articles are listed beneath a header for that day, with the most recent day first and articles reverse chronological within the day

#### Scenario: Article row shows time and title without bullets

- **WHEN** an article row is rendered for a story published at 15:04 local time
- **THEN** the row shows `HH:MM` after the tree glyph followed by the title, with no bullet markers anywhere in the row

#### Scenario: Source identifier right-aligned at the content edge

- **WHEN** an article row is rendered for a story linked from `newrepublic.com`
- **THEN** the identifier `newrepublic` is right-aligned at the content edge and no other field (such as the time) appears on the right

#### Scenario: Source identifier truncated to twelve characters

- **WHEN** an article row is rendered for a story linked from `reallylongnewspaperdomainname.net`
- **THEN** the source identifier is `reallylongne` (the first twelve characters of `reallylongnewspaperdomainname`)

#### Scenario: Source identifier falls back to feed host

- **WHEN** an article has no link and its feed URL is `https://nytimes.com/rss`
- **THEN** the source identifier is `nytimes`

#### Scenario: Source identifier strips multi-part public suffixes

- **WHEN** an article row is rendered for a story linked from `tribunemag.co.uk`
- **THEN** the source identifier is `tribunemag`, not the truncated remainder of the final-dot strip

#### Scenario: Truncated title ends with an ellipsis and keeps its gap

- **WHEN** a title is long enough to run into the right-aligned identifier
- **THEN** the title is truncated with `…` and at least one space remains between the ellipsis and the identifier

#### Scenario: Unread article title is bold

- **WHEN** the article list is displayed and an article has read status false
- **THEN** that article's title is rendered in bold bright text while its source identifier and time are dim

#### Scenario: Read article title is dim

- **WHEN** the article list is displayed and an article has read status true
- **THEN** that article's title, source identifier, and time are all rendered in the dim style

#### Scenario: Selection is a full-row highlight

- **WHEN** the cursor is on any visible row (day header or article)
- **THEN** the entire row, including the tree glyph column, is rendered with the selection background and no `> ` marker is shown

#### Scenario: Collapsed day hides its articles

- **WHEN** a day header is collapsed
- **THEN** that day's article rows are not rendered, and the header alone is shown

#### Scenario: Selection wraps from bottom to top

- **WHEN** the user presses down arrow or `j` while the last visible row is selected
- **THEN** selection moves to the first visible row

#### Scenario: Selection wraps from top to bottom

- **WHEN** the user presses up arrow or `k` while the first visible row is selected
- **THEN** selection moves to the last visible row

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view showing the
article count as `<n>/<total>`, where `n` is the number of articles currently
shown (after any active tag filter) and `total` is the total number of articles
in the database, followed by the date the feeds were last refreshed. The status
bar SHALL also display the on-disk size of the article database, formatted as a
human-readable value with binary unit suffixes (B, KB, MB, GB), on the same side
of the status bar as the last-refresh date. The reported size SHALL include the
WAL sidecar file when one is present, so it reflects the actual on-disk
footprint. When a tag filter is active, the status bar SHALL also display the
active filter name.

#### Scenario: Status bar shows filtered/total count, size, and refresh date

- **WHEN** the list view is displayed
- **THEN** the status bar shows the article count as `<n>/<total>` (shown over total), the database size as a human-readable value, and the last-refreshed date

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the `<n>/<total>` count, the database size, and the date

#### Scenario: Database size includes WAL sidecar

- **WHEN** the database has a WAL sidecar file containing uncheckpointed data and the list view is displayed
- **THEN** the status bar reports a size that includes both the main database file and the WAL sidecar file

#### Scenario: Database size formatted with binary unit suffix

- **WHEN** the database file is 1_500_000 bytes
- **THEN** the status bar displays the size as `1.4 MB`

#### Scenario: Database size refreshed after a refresh

- **WHEN** a refresh completes and new articles are stored
- **THEN** the status bar database size reflects the post-refresh on-disk footprint on the next list load

### Requirement: Article reader border rendering

The system SHALL render the article view inside a thin-line border. The article's
date and title SHALL appear inline with the top border. The left vertical border
SHALL be rendered in the same grey style as the top and bottom borders, while the
content text inside the border SHALL keep its own (bright) styling. The right
edge SHALL act as a scrollbar: a contiguous thumb segment rendered as a bright
double-line glyph represents the currently visible portion of the article and is
positioned along the track to reflect the current scroll offset, while the rest
of the track is rendered as a grey single-line glyph. The thumb SHALL be present
from the first frame when a scrollable article is opened, sitting at the top of
the track when the scroll offset is zero. The thumb height SHALL be proportional
to the fraction of the article that is visible, clamped to a minimum of one row
and a maximum of one row less than the track height (the right border's interior
height between the top and bottom borders). The thumb SHALL never fill the entire
track on a scrollable article, so it always reads as a movable indicator. The
bottom border SHALL display a left-aligned permanent help hint `o: open in
browser` and, on the right, a position indicator in the form
`<percent>% · <bottomLine>/<totalLines>`, where `<percent>` is the scroll
percentage, `<bottomLine>` is the line number of the last visible viewport line
clamped to the total, and `<totalLines>` is the article's total line count;
horizontal dashes fill the space between the hint and the indicator (for example
at the very bottom: `o: open in browser ───── 100% · 120/120`). The bullet used
between the percent and the line ratio SHALL be the same middle-dot bullet used
in the top border (`·`, ASCII `.`). The content area SHALL have two spaces of
horizontal padding on each side and one space of vertical padding at the top and
bottom. When the entire article fits within the viewport and no scrolling is
possible, the right border SHALL be fully filled with the bright double-line
glyph and the bottom border SHALL display `100%` with `<bottomLine>` equal to
`<totalLines>`.

#### Scenario: Short article fully visible

- **WHEN** an article is opened that fits entirely within the viewport
- **THEN** the top border displays the date and title, the right border is fully filled with the bright double-line glyph, and the bottom border displays the help hint on the left and `100% · <n>/<n>` on the right where `<n>` is the total line count

#### Scenario: Scrollbar visible on open

- **WHEN** a scrollable article (content taller than the viewport) is opened and the scroll offset is zero
- **THEN** the right border renders a bright double-line thumb at least one row tall at the top of the track, with the remainder of the track as a grey single line, and the bottom border displays `0% · <viewportHeight>/<totalLines>`

#### Scenario: Thumb moves with scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to 42% of the scrollable range
- **THEN** the thumb's top row is at 42% of the thumb's travel range along the track (rounded), rendered as a bright double-line, with the rest of the track as a grey single line, and the bottom border displays `42%` and a `<bottomLine>/<totalLines>` ratio consistent with that offset

#### Scenario: Thumb height stays within bounds

- **WHEN** a scrollable article is rendered with any combination of total content height and viewport height
- **THEN** the thumb height is at least one row and at most one row less than the right border's track height

#### Scenario: Thumb reaches the bottom at full scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to the very bottom (100% of the scrollable range)
- **THEN** the thumb sits at the bottom of the track as a bright double-line, with the rest of the track as a grey single line, and the bottom border displays `100% · <totalLines>/<totalLines>`

#### Scenario: Left border uses the grey border style

- **WHEN** the article view is rendered
- **THEN** the left vertical border is drawn in the same grey style as the top and bottom borders, and the content text retains its own bright styling

#### Scenario: Bottom border shows help hint and line indicator

- **WHEN** the article view is rendered
- **THEN** the bottom border shows `o: open in browser` left-aligned and `<percent>% · <bottomLine>/<totalLines>` right-aligned, with fill dashes between them

#### Scenario: ASCII fallback border

- **WHEN** ASCII fallback mode is enabled or the terminal lacks box-drawing support
- **THEN** single-line border glyphs are replaced with ASCII equivalents (`-`, `+`), the left and right vertical borders are both `:` rendered in the grey style, and the scrollbar thumb is `|` rendered in the bright/white style (the unfilled track stays `:` in grey)

### Requirement: Article top border title truncation

The system SHALL render the article's date and title inline with the top border
with the date always fully visible and the title left-aligned after the date.
The date prefix SHALL be `YYYY-MM-DD HH:MM:SS` showing the publication date and
time in the user's local timezone, followed by the middle-dot bullet (`·`, ASCII
`.`) and the title. When the title does not fit the available width, the system
SHALL truncate the title and end it with an ellipsis glyph — `…` in Unicode mode
and `...` in ASCII fallback mode — and SHALL place a single horizontal dash
adjacent to the right corner glyph so the right end mirrors the left (`┌─` on
the left, `─╖` on the right). When the title fits with room to spare, the system
SHALL fill the leftover space with horizontal dashes as before and SHALL NOT
append an ellipsis.

#### Scenario: Title fits is dash-filled

- **WHEN** the article title fits the top border with room to spare
- **THEN** the top border renders the full date-and-time prefix, the bullet, the full title, fill dashes, and the right corner, with no ellipsis

#### Scenario: Long title is truncated with an ellipsis

- **WHEN** the article title does not fit the top border
- **THEN** the date-and-time prefix remains fully visible, the title is truncated and ends with `…`, and a single dash precedes the right corner (for example `┌─ 2026-01-02 15:04:05 · Hello world artic… ─╖`)

#### Scenario: ASCII fallback ellipsis

- **WHEN** ASCII fallback mode is enabled and the article title does not fit
- **THEN** the truncated title ends with `...` instead of `…`

#### Scenario: Date and time stay visible for a very long title

- **WHEN** the article title is longer than the entire top-border content width
- **THEN** the `YYYY-MM-DD HH:MM:SS` date prefix is still rendered in full and only the title is truncated with an ellipsis

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation
(PgUp/PgDn, Ctrl-F/Ctrl-B, Space) SHALL scroll by a full viewport page. Half-page
navigation (Ctrl-D/Ctrl-U) SHALL scroll by half a viewport. `g` or
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

### Requirement: Clickable links in article content

The system SHALL render hyperlinks in article HTML content as clickable OSC 8
terminal hyperlinks. The rendered link SHALL display the styled link text
followed by the URL (`text url`), with the link text wrapped in
OSC 8 escape sequences carrying the link's URL, so that Cmd/Ctrl-clicking it in
a supporting terminal opens the URL in the system browser. The URL SHALL also be
displayed as visible fallback text following the link text, preserving a readable
fallback for terminals that do not support OSC 8. OSC 8 hyperlinks SHALL
function in ASCII fallback mode, since OSC 8 is an escape sequence and not a
Unicode glyph.

#### Scenario: Inline link rendered as clickable styled text with the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the link text "world" is styled and wrapped in OSC 8 escape sequences carrying the URL `https://example.com/post`, and the rendered text shows `world https://example.com/post`

#### Scenario: Cmd/Ctrl-click on link opens browser

- **WHEN** the user Cmd/Ctrl-clicks on the link text of an OSC 8 hyperlink in the article reader view in a supporting terminal
- **THEN** the terminal opens the link URL in the system browser

#### Scenario: Link URL preserved as fallback text

- **WHEN** an article with a link is displayed in a terminal that does not support OSC 8
- **THEN** the link text and the URL (`text url`) are displayed as plain text, identical to the behavior without OSC 8

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

### Requirement: Read-status on open and scroll

The system SHALL mark an article as read when it is opened and the entire article content fits within the viewport. If the article does not fit entirely within the viewport, the system SHALL NOT mark it read upon opening; instead, any single downward scroll action (down arrow, `j`, PgDn, Ctrl-F, Ctrl-D, Space) SHALL mark the article as read instantly.

#### Scenario: Short article marked read on open

- **WHEN** the user opens an article whose full content fits in the viewport
- **THEN** the article's read status is set to true immediately

#### Scenario: Long article not marked read on open

- **WHEN** the user opens an article whose content exceeds the viewport height
- **THEN** the article's read status remains false

#### Scenario: Long article marked read on first downward scroll

- **WHEN** a partially viewed article is displayed and the user presses down arrow, `j`, PgDn, Ctrl-F, Ctrl-D, or Space
- **THEN** the article's read status is set to true immediately

### Requirement: Tag popup on list view

The system SHALL open a tag popup occupying 80% of the screen when the user presses `T` on the list view. The popup SHALL list all tags sorted by popularity (total article count descending). Each tag SHALL display its unread/total counts in parentheses. When read count is zero (all unread) or unread count is zero (all read), the counts SHALL collapse to a single total number. The unread count SHALL always be rendered in bold. Navigation within the popup SHALL use arrow keys and hjkl and SHALL wrap around.

#### Scenario: Open tag popup from list

- **WHEN** the user presses `T` on the list view
- **THEN** a popup covering 80% of the screen appears listing all tags sorted by popularity

#### Scenario: All-unread tag collapses to total

- **WHEN** a tag has 0 read articles and 60 unread articles
- **THEN** the tag displays as `tagname (60)` with 60 in bold

#### Scenario: All-read tag collapses to total

- **WHEN** a tag has 60 read articles and 0 unread articles
- **THEN** the tag displays as `tagname (60)` with 60 in plain text

#### Scenario: Mixed tag shows unread/total

- **WHEN** a tag has 45 read articles and 15 unread articles out of 60 total
- **THEN** the tag displays as `tagname (15/60)` with 15 in bold and 60 in plain text

#### Scenario: Selecting a tag filters the list

- **WHEN** the user selects a tag in the popup and confirms
- **THEN** the list view closes the popup and displays only articles with that tag

### Requirement: Tag popup on article view

The system SHALL open a tag popup occupying 40% of the screen when the user presses `T` on the article view. The popup SHALL list only the categories assigned to the current article, sorted by popularity. Count display and navigation SHALL follow the same rules as the list-view tag popup. Selecting a tag SHALL filter the list view and navigate back to it.

#### Scenario: Open tag popup from article

- **WHEN** the user presses `T` on the article view
- **THEN** a popup covering 40% of the screen appears listing only the current article's tags

#### Scenario: Selecting a tag from article popup filters list

- **WHEN** the user selects a tag in the article-view popup
- **THEN** the system returns to the list view filtered by that tag

### Requirement: Help popup

The system SHALL display a help popup listing each action and its bound keys
when the user presses `?`. Each action SHALL be labeled with its display name
derived from the action identifier by replacing underscores with spaces and
rendering the result in Title Case (for example `half_page_down` →
`Half Page Down`). The action label SHALL appear on the left
and the bound key strings SHALL appear on the right.

#### Scenario: Action labels are Title Case without underscores

- **WHEN** the help popup is displayed
- **THEN** each action is labeled with underscores replaced by spaces and each word capitalized (for example `Open Article`, `Half Page Down`)

#### Scenario: Help popup lists current bindings

- **WHEN** the user presses `?` in any view
- **THEN** a popup appears listing each action's display name on the left and its currently bound keys on the right

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

### Requirement: Open-article bounds safety

When the user opens an article from the list view, the system SHALL verify the
current cursor index is within the bounds of the loaded article list before
accessing it. When the list is empty or the cursor is otherwise out of range
(such as a stale cursor left over from a list that has since shrunk), the
open-article action SHALL be a no-op and SHALL NOT panic. The bounds check SHALL
not depend on any prior clamping having occurred.

#### Scenario: Open article on empty list is a no-op

- **WHEN** the article list is empty and the user presses the open-article key
- **THEN** nothing happens and the application does not panic

#### Scenario: Open article with a stale cursor does not panic

- **WHEN** the cursor index is greater than or equal to the current length of the loaded article list (for example after a refresh shrank the list) and the user presses the open-article key
- **THEN** the action is a no-op and the application does not panic

### Requirement: Quit and utility keys

The system SHALL quit on `q` or Ctrl-C from any view. The system SHALL provide `m` to toggle read/unread status, `a` to mark all articles in the current view as read, `o` to open the article URL in the system browser, `c` to copy the article URL to the clipboard, `C` to copy the article text to the clipboard, and `?` to display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in any view
- **THEN** the application exits

#### Scenario: Toggle read status

- **WHEN** the user presses `m` on an article in the list view or article view
- **THEN** the article's read status is toggled between read and unread

#### Scenario: Mark all as read

- **WHEN** the user presses `a` in the list view
- **THEN** all articles currently displayed are marked as read

#### Scenario: Copy article text to clipboard

- **WHEN** the user presses `C` in the article view
- **THEN** the full article text is copied to the system clipboard

### Requirement: URL open and copy argument safety

When the user opens an article URL in the system browser or copies it to the
clipboard, the system SHALL pass the URL to the platform opener or clipboard helper
as a single exec argument without routing it through a command shell. The system
SHALL NOT invoke a shell (`cmd.exe /c` or equivalent) with the raw, feed-controlled
URL, so that metacharacters in an article's link (such as `&` or `|`) cannot inject
additional commands. Clipboard copy SHALL feed the URL to the clipboard helper via
its standard input, not via a shell command string.

#### Scenario: Open URL uses a non-shell single-argument opener on Windows

- **WHEN** the user opens an article URL on Windows and the link contains shell metacharacters such as `&` or `|`
- **THEN** the system opens the URL via a non-shell opener that receives the URL as a single argument, and no additional command is executed

#### Scenario: Open URL on Unix uses a single exec argument

- **WHEN** the user opens an article URL on macOS or Linux
- **THEN** the system passes the URL as a single exec argument to the platform opener without a shell

#### Scenario: Copy URL feeds stdin without a shell

- **WHEN** the user copies an article URL to the clipboard
- **THEN** the system pipes the URL to the clipboard helper's standard input as a single value, not via a shell command string

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown and then render that
markdown to styled terminal text using a theme-aware markdown renderer that
wraps its output to the content width. The renderer SHALL follow the configured
display theme: in `dark` mode it SHALL use the dark theme, in `light` mode the
light theme, and in `auto` mode the theme SHALL be chosen from the terminal's
preferred background. Inline `<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available; the alt text
SHALL be rendered literally, so markdown special characters in the alt text
(such as `*`, `_`, `[`, or `]`) SHALL NOT be interpreted as styling or link
syntax. A publisher-attached lead image (from `<enclosure>` or
`media:thumbnail`) is rendered separately as an inline halfblocks image block
above the rendered article content and is not affected by this placeholder rule.
`<strong>` and `<b>` SHALL render with the theme's bold styling and
`<em>`/`<i>` with its italic styling, not as literal markdown markers.
Hyperlinks SHALL render as the styled link text followed by the URL
(`text url`) and SHALL be wrapped in OSC 8 hyperlink escape sequences carrying
the URL. Structured HTML SHALL render as styled markdown: ordered and unordered
lists (including nested lists) with correct bullet markers and indentation,
blockquotes, code blocks, and tables. When a link's rendered text does not fit
on the current line, the renderer SHALL wrap it onto a new line. When a URL is
too long to fit on a single line by itself, the renderer SHALL wrap it across
lines; it SHALL NOT truncate the displayed URL. Non-link text SHALL be
unaffected. When HTML-to-markdown conversion fails, the system SHALL render the
article body as plain text with HTML markup removed rather than displaying the
raw HTML source. When markdown rendering fails, the system SHALL display the
article text without markdown formatting markers (no literal `**` or
`[text](url)` syntax) rather than the raw markdown source.

#### Scenario: HTML content rendered as styled markdown with paragraph breaks

- **WHEN** an article with HTML content is displayed in the reader view
- **THEN** the content is rendered as styled terminal markdown text with paragraph breaks preserved

#### Scenario: Bold text rendered with theme styling

- **WHEN** an article contains `<p><strong>bold text</strong> and <b>more</b> plain</p>`
- **THEN** "bold text" and "more" are rendered with the theme's bold styling and no literal markdown asterisks are displayed

#### Scenario: Bold inside a link stays styled

- **WHEN** an article contains `<a href="https://example.com/x"><strong>bold link</strong></a>`
- **THEN** the link text "bold link" keeps its bold styling inside the rendered link, with no literal markdown asterisks

#### Scenario: Markdown theme follows the display theme

- **WHEN** the configured display theme is `dark` (or `light`, or `auto` resolving to the terminal background)
- **THEN** the article content is rendered with glamour's matching dark (or light) theme

#### Scenario: Nested list renders with correct indentation

- **WHEN** an article contains a `<ul>` with a nested `<ul>` inside one of its `<li>` elements
- **THEN** top-level items render with a bullet marker at the leftmost column and nested items render indented under their parent with their own bullet markers

#### Scenario: List item containing a paragraph renders on one line

- **WHEN** an article contains `<ul><li><p>first point</p></li></ul>` (as emitted by Substack and similar publishers)
- **THEN** the bullet marker and the item text appear on the same line, with no orphaned bullet marker on its own line

#### Scenario: Blockquote rendered as a styled blockquote

- **WHEN** an article contains a `<blockquote>` element
- **THEN** the quoted text is rendered with the theme's blockquote styling and a quote marker

#### Scenario: Image rendered as alt text

- **WHEN** an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

#### Scenario: Image alt text with markdown characters renders literally

- **WHEN** an article contains an `<img>` element with alt text "a *b* [c]"
- **THEN** the reader displays `[a *b* [c]]` literally, with the `*` and `[`/`]` shown as plain characters and no styled or link formatting applied

#### Scenario: Link rendered as styled text followed by the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the reader displays "world" styled as a link followed by `https://example.com/post`, with the link wrapped in OSC 8 hyperlink escapes

#### Scenario: Long URL wraps onto a new line

- **WHEN** a link's rendered text does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL wraps across lines

- **WHEN** a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is wrapped across multiple lines and is not truncated or elided

#### Scenario: Conversion failure renders plain text, not raw HTML

- **WHEN** HTML-to-markdown conversion of the article content fails
- **THEN** the reader displays the article's plain text with HTML markup removed, and no `<p>`/`<a>`/`<div>` tag text is shown

#### Scenario: Markdown render failure shows text without markdown markers

- **WHEN** the markdown renderer fails to render the article content
- **THEN** the reader displays the article text without literal markdown formatting markers such as `**` or `[text](url)`
### Requirement: Article lead image rendering

When image rendering is enabled and an article has a publisher-attached image
URL (captured from `<enclosure>` or `media:thumbnail` during feed parsing), the
system SHALL render the image inline below the reader header (the URL, title,
and author lines) and above the rendered article content, as a block of content
that scrolls with the article text. The image SHALL be rendered using a Unicode
halfblocks representation (half-block characters with ANSI color) sized to fit
the content width while preserving aspect ratio, and SHALL be horizontally
centered within the content width. Directly beneath the image, with no blank
line between them, the system SHALL render a photo attribution constrained to
the photo's own width (wrapped onto additional centered lines when the credit
is longer) and centered beneath the photo; a blank line SHALL follow the
attribution before the article content resumes. The attribution text SHALL be
the photo's credit extracted from the article's stored HTML when present — a
`figcaption` associated with the figure containing the lead image, otherwise an
element whose class indicates a credit line — and SHALL fall back to
`photo: <source>`, where `<source>` is derived by the same source-identifier
rules used for the list view's organization column. The combined image-plus-
attribution block height SHALL be capped so that the full block plus at least
one line of body text fits within the viewport height; when the block is taller
than this cap — for example because a wrapped attribution needs more lines —
it SHALL be scaled down to fit. When image rendering is disabled
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

#### Scenario: Attribution wraps under the photo

- **WHEN** the attribution text is longer than the rendered photo's width (for example after the photo is scaled down by the height cap)
- **THEN** the attribution wraps onto additional lines within the photo's width, each centered beneath the photo, and the photo is re-rendered smaller so the whole block plus at least one line of body text still fits the viewport

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

### Requirement: Async image load lifecycle

The system SHALL fetch article lead images asynchronously so that opening an article never blocks the UI on a network request. Opening an article with an image URL SHALL first attempt to load the image from the cache hierarchy (in-memory decoded image, then stored database bytes); only on a cache miss SHALL it issue a network fetch command. On a successful load (from cache or network), the system decodes the image and produces a load message that triggers a re-render with the image block inserted; when the load came from a network fetch, the raw image bytes SHALL be persisted to the database. On failure (network error, timeout, non-image content type, or exceeded byte limit), the system SHALL produce a failure message and render the article without an image block. The decoded image SHALL be cached in memory by URL for the session so that resizing the viewport does not re-fetch or re-decode from the database. When the terminal is resized while an article with a loaded image is open, the system SHALL re-render the image block to the new content width using the cached decoded image without re-fetching. When an image loads after the user has scrolled past the insertion point, the system SHALL preserve the user's reading position by adjusting the viewport offset by the number of inserted image-block lines.

#### Scenario: Article opens with image loading asynchronously

- **WHEN** the user opens an article that has an enclosure image URL and image rendering is enabled
- **THEN** the article text renders immediately and the image fetch runs in the background

#### Scenario: Image loads and appears inline

- **WHEN** the background image fetch succeeds while the article is open and the user is at the top of the article
- **THEN** the image block is inserted at the top of the content and becomes visible

#### Scenario: Image fetch fails gracefully

- **WHEN** the background image fetch fails or times out
- **THEN** the article renders without an image block and no error is surfaced to the user

#### Scenario: Resize re-renders image without re-fetch

- **WHEN** the terminal is resized while an article with a loaded image is open
- **THEN** the image block is re-rendered to the new content width using the cached image and no network request is made

#### Scenario: Scroll position preserved on late image load

- **WHEN** an image loads after the user has scrolled into the body text past the image insertion point
- **THEN** the viewport offset is increased by the number of image-block lines so the user's reading position is unchanged

#### Scenario: Re-opened article uses cached image

- **WHEN** the user closes and re-opens an article whose image was previously fetched
- **THEN** the image block renders immediately from the in-memory cache without a network request

#### Scenario: Restarted article uses stored image bytes

- **WHEN** the application is restarted and the user opens an article whose image was previously fetched and stored in the database
- **THEN** the image block renders from the stored bytes without making a network request to the image host

### Requirement: Restore reader state on start

The system SHALL persist the reader's selection when the application quits and
restore it when the application starts again. The persisted selection SHALL
record the current view (list or article), the row under the cursor (the
article by its ID, or the day group by its calendar-day key when the cursor is
on a day header), and, when the article view is open, the article's scroll
offset. On startup the system SHALL move the list cursor to the row matching
the saved selection. When the saved view is the article view and the saved
article still exists, the system SHALL open that article and restore its scroll
offset, so that returning to the list (via `b`, `esc`, or `h`) leaves the same
article selected. When the saved selection no longer exists — for example the
article was pruned or the day group changed — the system SHALL fall back to the
default clamped cursor without error.

#### Scenario: List cursor restored on start

- **WHEN** the user quits with the cursor on an article row in the list view and restarts
- **THEN** the list cursor is on that same article's row

#### Scenario: Day header selection restored on start

- **WHEN** the user quits with the cursor on a day-header row and restarts
- **THEN** the list cursor is on that day group's header row

#### Scenario: Article view reopened on start

- **WHEN** the user quits while reading an article and restarts
- **THEN** the reader reopens that article, and pressing `b`, `esc`, or `h` returns to the list with that article still selected

#### Scenario: Article scroll position restored on start

- **WHEN** the user quits while reading an article scrolled partway and restarts
- **THEN** the reopened article is scrolled to the same position

#### Scenario: Fallback when the saved article no longer exists

- **WHEN** the saved article was removed from the store and the user restarts
- **THEN** the list opens with the default cursor and the application does not error

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

### Requirement: In-view text selection

The system SHALL let the user select article text with the mouse in the article
view and automatically copy the selected text to the system clipboard on
release. An unmodified left-button press on the article content area SHALL
anchor the selection at that cell. While the button is held, dragging SHALL
extend the selection to the current cell. Releasing the button SHALL copy the
selected text to the system clipboard and clear the selection highlight. The
selection SHALL span the rendered viewport cells, so it tracks the article's
current scroll position and wrapping. The system SHALL render the selected
cells with an inverted (highlighted) style distinct from unselected text.
Copied text SHALL be the plain text characters within the selected region,
preserving the line breaks of the wrapped rendering, with all ANSI styling and
OSC 8 hyperlink escape sequences stripped. A press with an empty or zero-width
selection SHALL copy nothing. Starting a new selection SHALL replace the
previous selection. Leaving the article view SHALL clear the selection.

#### Scenario: Press anchors selection

- **WHEN** the user presses the left button on a cell in the article content area
- **THEN** a selection is anchored at that cell with no highlight yet

#### Scenario: Drag extends selection

- **WHEN** the user holds the left button and drags across cells in the article content area
- **THEN** the selection spans from the anchor cell to the current cell and those cells are rendered inverted

#### Scenario: Release copies selection

- **WHEN** the user releases the left button after dragging a selection
- **THEN** the selected text is copied to the system clipboard and the selection highlight is cleared

#### Scenario: Selected text is plain with escapes stripped

- **WHEN** the selected region includes text rendered with ANSI styling or OSC 8 hyperlink sequences
- **THEN** the copied text contains the visible characters with no ANSI or OSC 8 escape bytes

#### Scenario: Selection tracks scroll

- **WHEN** the user drags a selection over a scrollable article
- **THEN** the highlighted cells correspond to the visible viewport content at the current scroll offset

#### Scenario: New selection replaces previous

- **WHEN** the user starts a second drag selection after a first selection exists
- **THEN** the second selection replaces the first and the first's highlight is cleared as the drag proceeds

#### Scenario: Zero-width selection copies nothing

- **WHEN** the user presses and releases the left button without moving
- **THEN** nothing is copied to the clipboard

#### Scenario: Leaving the article clears the selection

- **WHEN** the user returns from the article view to the list view with a selection active
- **THEN** the selection and its highlight are cleared

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
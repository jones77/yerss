## Purpose

Provides the terminal user interface for browsing, reading, and filtering RSS articles — comprising the article list view, the article reader view with a scroll-progress border, and the tag popup for category-based filtering.

## Requirements

### Requirement: Article list rendering

The system SHALL display articles grouped under a collapsible header for each
day, with the most recent day first and articles within a day sorted reverse
chronologically. Each day header SHALL display the long date in the form
`Weekday Day-ordinal Month, Year` (for example `Saturday 20th August, 2025`),
computed in the user's local timezone, where the day is suffixed with an English
ordinal (`1st`, `2nd`, `3rd`, `4th`, …). The system SHALL NOT render a top title
banner; the most recent day header is the first content at the top of the list.
Each article row SHALL show three fields separated by a middle-dot bullet (`·`,
ASCII fallback `.`): the article title on the left, a source-domain identifier,
and the publication time as `HH:MM` on the right, in the user's local timezone.
The source-domain identifier SHALL be derived from the article's link host by
stripping any leading `www.`, removing the top-level domain (the part after the
final dot), and truncating the remainder to at most 12 characters (for example
`newrepublic.com` → `newrepublic`, `sueddeutsche.de` → `sueddeutsche`,
`reallylongnewspaperdomainname.net` → `reallylongne`). When the article has no
link, the system SHALL derive the identifier from the feed URL host instead.
The source identifier and the publication time SHALL always be rendered in the
dim/grey style. Only the title SHALL change with read state: unread titles SHALL
be bold and bright, and read titles SHALL be dim. A collapsed day header SHALL
hide its article rows; an expanded day header SHALL show them. Selection SHALL
wrap around when moving past the first or last visible row (headers plus expanded
articles).

#### Scenario: Unread article title is bold

- **WHEN** the article list is displayed and an article has read status false
- **THEN** that article's title is rendered in bold bright text while its source identifier and time are dim

#### Scenario: Read article title is dim

- **WHEN** the article list is displayed and an article has read status true
- **THEN** that article's title, source identifier, and time are all rendered in the dim style

#### Scenario: Articles grouped under a day header

- **WHEN** the article list is displayed and articles exist on more than one day
- **THEN** each day's articles are listed beneath a header for that day, with the most recent day first and articles reverse chronological within the day

#### Scenario: Day header shows the long local date

- **WHEN** a day header is rendered for a date with day-of-month 20 in August 2025
- **THEN** the header displays the long date with weekday, an ordinal day number, month, and year (for example `Wednesday 20th August, 2025`) in the user's local timezone

#### Scenario: Article row shows title, source, and HH:MM separated by bullets

- **WHEN** an article row is rendered for a story linked from `newrepublic.com` published at 15:04 local time
- **THEN** the row shows the title, followed by a `·` bullet, `newrepublic`, a `·` bullet, and `15:04` on the right

#### Scenario: Source identifier truncated to twelve characters

- **WHEN** an article row is rendered for a story linked from `reallylongnewspaperdomainname.net`
- **THEN** the source identifier is `reallylongne` (the first twelve characters of `reallylongnewspaperdomainname`)

#### Scenario: Source identifier falls back to feed host

- **WHEN** an article has no link and its feed URL is `https://nytimes.com/rss`
- **THEN** the source identifier is `nytimes`

#### Scenario: Collapsed day hides its articles

- **WHEN** a day header is collapsed
- **THEN** that day's article rows are not rendered, and the header alone is shown

#### Scenario: No title banner at the top

- **WHEN** the article list is displayed
- **THEN** no `yerss` title or branding line is rendered; the first rendered row is the most recent day header (or an empty-state message when there are no articles)

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
terminal hyperlinks. The link text (the portion inside the markdown square
brackets) SHALL be wrapped in OSC 8 escape sequences that carry the link's URL,
so that Cmd/Ctrl-clicking the link text in a supporting terminal opens the URL in
the system browser. The URL SHALL also be displayed as plain text inside the
markdown parentheses, preserving a readable fallback for terminals that do not
support OSC 8. OSC 8 hyperlinks SHALL function in ASCII fallback mode, since OSC
8 is an escape sequence and not a Unicode glyph.

#### Scenario: Inline link rendered as a clickable markdown link

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the link text "world" is wrapped in OSC 8 escape sequences carrying the URL `https://example.com/post`, and the markdown `[world](https://example.com/post)` is displayed

#### Scenario: Cmd/Ctrl-click on link opens browser

- **WHEN** the user Cmd/Ctrl-clicks on the link text of an OSC 8 hyperlink in the article reader view in a supporting terminal
- **THEN** the terminal opens the link URL in the system browser

#### Scenario: Link URL preserved as fallback text

- **WHEN** an article with a link is displayed in a terminal that does not support OSC 8
- **THEN** the markdown `[text](url)` is displayed as plain text, identical to the behavior without OSC 8

#### Scenario: OSC 8 links work in ASCII mode

- **WHEN** ASCII fallback mode is enabled and an article with links is displayed
- **THEN** the link text is wrapped in OSC 8 escape sequences and the markdown links are clickable in supporting terminals

### Requirement: Clickable article URL in reader header

The system SHALL render the article's own URL (displayed in the reader view
header) as a clickable OSC 8 terminal hyperlink, so that Cmd/Ctrl-clicking it
opens the URL in the system browser. This mirrors the existing `o` key behavior
for opening the article URL.

#### Scenario: Article header URL is clickable

- **WHEN** an article with a link URL is displayed in the reader view and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser

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
rendering the result in Title Case (for example `open_article` → `Open Article`,
`mark_all_read` → `Mark All Read`). The action label SHALL appear on the left
and the bound key strings SHALL appear on the right.

#### Scenario: Action labels are Title Case without underscores

- **WHEN** the help popup is displayed
- **THEN** each action is labeled with underscores replaced by spaces and each word capitalized (for example `Open Article`, `Mark All Read`, `Toggle Read`)

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

The system SHALL quit on `q` or Ctrl-C from any view. The system SHALL provide `m` to toggle read/unread status, `a` to mark all articles in the current view as read, `o` to open the article URL in the system browser, `c` to copy the article URL to the clipboard, and `?` to display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in any view
- **THEN** the application exits

#### Scenario: Toggle read status

- **WHEN** the user presses `m` on an article in the list view or article view
- **THEN** the article's read status is toggled between read and unread

#### Scenario: Mark all as read

- **WHEN** the user presses `a` in the list view
- **THEN** all articles currently displayed are marked as read

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

The system SHALL render article HTML content as plain text using an HTML-to-text
conversion that preserves paragraph breaks. Images SHALL be rendered as `[alt]`
using the image's alt text, or `[image]` if no alt text is available. Hyperlinks
SHALL be rendered in markdown style as `[text](url)`: the link text inside the
square brackets, followed by the URL in parentheses. When a link's markdown
rendering does not fit on the current line, the system SHALL break the URL onto a
new line. When a URL is too long to fit on a single line by itself, the system
SHALL truncate the displayed URL with an ellipsis glyph (`…` in Unicode mode,
`...` in ASCII fallback mode) so it fits on one line. Non-link text SHALL be
unaffected.

#### Scenario: HTML content converted to plain text

- **WHEN** an article with HTML content is displayed in the reader view
- **THEN** the content is rendered as plain text with paragraph breaks preserved

#### Scenario: Image rendered as alt text

- **WHEN** an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

#### Scenario: Link rendered as markdown

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the reader displays `[world](https://example.com/post)`

#### Scenario: Long URL breaks onto a new line

- **WHEN** a link's markdown rendering does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL is truncated with an ellipsis

- **WHEN** a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is truncated to fit one line and ends with `…` (or `...` in ASCII fallback mode)

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
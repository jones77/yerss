## Purpose

Provides the terminal user interface for browsing, reading, and filtering RSS articles — comprising the article list view, the article reader view with a scroll-progress border, and the tag popup for category-based filtering.

## Requirements

### Requirement: Article list rendering

The system SHALL display articles grouped under a collapsible header for each
day, with the most recent day first and articles within a day sorted reverse
chronologically. The day-group headers SHALL be rendered with a tree rail down
the left edge: the first day group SHALL be prefixed with `┌` (ASCII `+`), the
last day group with `└` (ASCII `+`), and every other day group with `├`
(ASCII `+`), each followed by four horizontal dashes (`────`, ASCII `----`) and
a single space. Article rows SHALL NOT carry a
tree glyph; the publication time SHALL start the row. Day
headers SHALL NOT display fold markers. Each day header SHALL display the long
date in the form `Weekday Day-ordinal Month, Year` (for example
`Saturday 20th August, 2025`), computed in the user's local timezone, where the
day is suffixed with an English ordinal (`1st`, `2nd`, `3rd`, `4th`, …). The
header for the current local calendar day SHALL be prefixed with `today, ` and
the header for the previous local calendar day SHALL be prefixed with
`yesterday, `; older days and the `Undated` group SHALL be unprefixed. Day
headers SHALL render entirely in the teal role: the rail corner, the dash
connector, and the date label. Each
article row SHALL show the publication time at the start of the row in the
user's local timezone (`--:--` when undated), as `HH:MM` by default and as
`HH:MM:SS` when the `show_seconds` display option is enabled, followed by the
article title; the title SHALL NOT be preceded or followed by bullet points. The
source-domain identifier SHALL be right-aligned at the content edge and SHALL
be the only field on the right; it SHALL be derived from the article's link
host by stripping any leading `www.`, taking the organization label of the
registrable domain (the public suffix plus one, so multi-part public suffixes
such as `co.uk` are removed whole), and middle-eliding it to at most 15
characters when longer (for example `newrepublic.com` → `newrepublic`,
`reallylongnewspaperdomainname.net` → `reallyl…ainname`). When the article has no
link, the system SHALL derive the identifier from the feed URL host instead;
hosts that have no registrable domain (single-label hosts, IP literals) fall
back to removing the part after the final dot.
When a title would run into the right-aligned identifier, the system SHALL
leave at least one space between the title and the identifier, and a truncated
title SHALL end with an ellipsis — `…` in Unicode mode and `...` in ASCII
fallback mode. The source identifier and the publication time SHALL always be
rendered in the text (white) style. Only the title SHALL change with read
state: unread titles SHALL be bold and bright, and read titles SHALL be white.
Selection SHALL be indicated by a full-row background highlight covering the
row's content — including the rail glyph on day-header rows — on day-header
rows and article rows alike;
the system SHALL NOT render a cursor-gutter marker. A collapsed day header
SHALL hide its article rows; an expanded day header SHALL show them. Selection
SHALL wrap around when moving past the first or last visible row (headers plus
expanded articles).

#### Scenario: Tree rail bookends the list

- **WHEN** the article list is displayed with more than one day group
- **THEN** the first day header is prefixed with `┌────`, the last day header with `└────`, and every day header between with `├────`

#### Scenario: Day headers carry the rail, article rows do not

- **WHEN** the list is displayed
- **THEN** each day header is prefixed `┌`, `├`, or `└` followed by four horizontal dashes (`────`) and a single space, and each article row carries no tree glyph, starting with its publication time

#### Scenario: Tree glyphs are stable while scrolling

- **WHEN** the visible window scrolls so the list's first and last day groups are off screen
- **THEN** the visible day headers are all prefixed `├────` and the bookend glyphs stay on the first and last day groups rather than moving to the window edges

#### Scenario: ASCII fallback tree glyphs

- **WHEN** ASCII fallback mode is enabled
- **THEN** the day-header tree glyphs render as `+` (corners and tees) with `----` dash connectors and article rows carry no tree glyph

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

#### Scenario: Day header renders in the teal role

- **WHEN** a day header is rendered in dark mode
- **THEN** the rail corner, the dash connector, and the date label render in ANSI color 14 (bright cyan), and in ANSI color 6 (cyan) in light mode

#### Scenario: Article row shows time and title without bullets

- **WHEN** an article row is rendered for a story published at 15:04 local time
- **THEN** the row shows `HH:MM` at the start of the row followed by the title, with no bullet markers anywhere in the row

#### Scenario: Article row shows seconds when enabled

- **WHEN** the `show_seconds` display option is enabled and an article row is rendered for a story published at 15:04:05 local time
- **THEN** the row shows `15:04:05` at the start of the row followed by the title

#### Scenario: Source identifier right-aligned at the content edge

- **WHEN** an article row is rendered for a story linked from `newrepublic.com`
- **THEN** the identifier `newrepublic` is right-aligned at the content edge and no other field (such as the time) appears on the right

#### Scenario: Source identifier middle-elided to fifteen characters

- **WHEN** an article row is rendered for a story linked from `reallylongnewspaperdomainname.net`
- **THEN** the source identifier is `reallyl…ainname`, middle-elided to 15 characters

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
- **THEN** that article's title is rendered in bold bright text while its source identifier and time are white

#### Scenario: Read article title is white

- **WHEN** the article list is displayed and an article has read status true
- **THEN** that article's title, source identifier, and time are all rendered in the text (white) style

#### Scenario: Selection is a full-row highlight

- **WHEN** the cursor is on any visible row (day header or article)
- **THEN** the entire row, including the rail glyph on day-header rows, is rendered with the selection background and no `> ` marker is shown

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

The system SHALL display a status bar at the bottom of the list view. The
left-most elements SHALL be the time and long-form date the feeds were last
refreshed, formatted as `HH:MM Weekday Day-ordinal Month, Year last refresh`
(for example `14:30 Saturday 29th August, 2026 last refresh`), or
`HH:MM:SS Weekday Day-ordinal Month, Year last refresh` when the `show_seconds`
display option is enabled, with the
`last refresh` label rendered in the dim role (#707070), or the text
`never refreshed` when no refresh has occurred. The right-most elements SHALL
be a literal `?: help` hint, followed by the bullet and the on-disk size of the
article database as `DB <n>MB`, the scroll position percentage, and the article
count as `<n>/<total>` — where the database size is rendered as a whole number
of megabytes (floored to an integer, with no space between the number and `MB`)
and `<total>` is the total number of articles in the database — joined by the
bullet (`•`, ASCII `.`) (for example
`?: help • DB 320MB • 100% • 2/2`). The `DB` value SHALL be the on-disk database
size including the WAL sidecar file when one is present. Values below one
megabyte SHALL render as `0MB`. The bullets and the `last refresh` label SHALL
render in the dim role (#707070); the remaining status bar text SHALL render in
the chrome role. When a tag filter is active, the status bar SHALL also display
the active filter name.

#### Scenario: Status bar leads the right side with the help hint

- **WHEN** the list view is displayed
- **THEN** the status bar's right side reads `?: help • DB <n>MB • <percent>% •
  <n>/<total>`, with `?: help` as the first right-aligned element

#### Scenario: Status bar shows refresh time, size, percentage, and position

- **WHEN** the list view is displayed
- **THEN** the status bar shows the last-refresh time and date on the left as
  `HH:MM Weekday Day-ordinal Month, Year last refresh` and, on the right, the
  database size as `DB <n>MB`, the percentage, and the `<n>/<total>` position
  joined by ` • `

#### Scenario: Status bar shows seconds when enabled

- **WHEN** the `show_seconds` display option is enabled and the list view is displayed
- **THEN** the status bar shows the last-refresh time on the left as `HH:MM:SS Weekday Day-ordinal Month, Year last refresh`

#### Scenario: Status bar uses dim bullets and dim refresh label

- **WHEN** the list view is displayed
- **THEN** the bullets between the database size, percentage, and position
  render in the dim role (#707070), the `last refresh` label renders in the dim
  role, and the elements between the bullets render in the chrome role

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the refresh
  time, the database size, the percentage, and the `<n>/<total>` position

#### Scenario: Database size includes WAL sidecar

- **WHEN** the database has a WAL sidecar file containing uncheckpointed data
  and the list view is displayed
- **THEN** the `DB` value reports a size that includes both the main database
  file and the WAL sidecar file

#### Scenario: Database size is a whole megabyte

- **WHEN** the database file is 1_500_000 bytes
- **THEN** the status bar displays `DB 1MB` as a floored whole number with no
  space between the number and `MB`

#### Scenario: Database size refreshed after a refresh

- **WHEN** a refresh completes and new articles are stored
- **THEN** the status bar `DB` value reflects the post-refresh on-disk
  footprint on the next list load

### Requirement: Article reader border rendering

The system SHALL render the article view inside a thin-line border. The article's
date and title SHALL appear inline with the top border. The left vertical border
SHALL be rendered in the same grey style as the top and bottom borders, while the
content text inside the border SHALL keep its own (bright) styling. The right
edge SHALL act as a scrollbar: a contiguous thumb segment rendered as a bright
single-line glyph represents the currently visible portion of the article and is
positioned along the track to reflect the current scroll offset, while the rest
of the track is rendered as a grey single-line glyph. The scrollbar style SHALL
be configurable via the `scrollbar` display option: the default `single` renders
the thumb as one line (`│` Unicode, `|` ASCII), and `double` renders it as a
double line (`║`, Unicode only; ASCII fallback has no double-line glyph). The
thumb SHALL be present
from the first frame when a scrollable article is opened, sitting at the top of
the track when the scroll offset is zero. The thumb height SHALL be proportional
to the fraction of the article that is visible, clamped to a minimum of one row
and a maximum of one row less than the track height (the right border's interior
height between the top and bottom borders). The thumb SHALL never fill the entire
track on a scrollable article, so it always reads as a movable indicator. The
bottom border SHALL display a left-aligned permanent help hint `o: open
article in browser` and, on the right, a position indicator led by a literal
`?: help` hint in the form
`?: help • <percent>% • <bottomLine>/<totalLines>`, where `<percent>` is the
scroll percentage, `<bottomLine>` is the line number of the last visible
viewport line clamped to the total, and `<totalLines>` is the article's total
line count; horizontal dashes fill the space between the hint and the indicator
(for example at the very bottom:
`o: open article in browser ───── ?: help • 100% • 120/120`). The bullets used
between the help affordance and the percent and between the percent and the
line ratio SHALL be the same bullet used in the top border (`•`,
ASCII `.`). The content area SHALL have two spaces of
horizontal padding on each side and one space of vertical padding above and
below the content, so a permanent one-row margin separates the content
viewport from both the top and the bottom borders. When the entire article fits within the viewport
and no scrolling is
possible, the right border SHALL be fully filled with the bright single-line
glyph and the bottom border SHALL display `100%` with `<bottomLine>` equal to
`<totalLines>`.

#### Scenario: Short article fully visible

- **WHEN** an article is opened that fits entirely within the viewport
- **THEN** the top border displays the date and title, the right border is fully filled with the bright single-line glyph, and the bottom border displays the help hint on the left and `100% • <n>/<n>` on the right where `<n>` is the total line count

#### Scenario: Scrollbar visible on open

- **WHEN** a scrollable article (content taller than the viewport) is opened and the scroll offset is zero
- **THEN** the right border renders a bright single-line thumb at least one row tall at the top of the track, with the remainder of the track as a grey single line, and the bottom border displays `0% • <viewportHeight>/<totalLines>`

#### Scenario: Thumb moves with scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to 42% of the scrollable range
- **THEN** the thumb's top row is at 42% of the thumb's travel range along the track (rounded), rendered as a bright single-line, with the rest of the track as a grey single line, and the bottom border displays `42%` and a `<bottomLine>/<totalLines>` ratio consistent with that offset

#### Scenario: Thumb height stays within bounds

- **WHEN** a scrollable article is rendered with any combination of total content height and viewport height
- **THEN** the thumb height is at least one row and at most one row less than the right border's track height

#### Scenario: Thumb reaches the bottom at full scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to the very bottom (100% of the scrollable range)
- **THEN** the thumb sits at the bottom of the track as a bright single-line, with the rest of the track as a grey single line, and the bottom border displays `100% • <totalLines>/<totalLines>`

#### Scenario: Configurable scrollbar style

- **WHEN** the `scrollbar` display option is set to `double` and an article is displayed in Unicode mode
- **THEN** the thumb renders as the bright double-line glyph `║`; with the default `single` it renders as the single line `│`

#### Scenario: Left border uses the grey border style

- **WHEN** the article view is rendered
- **THEN** the left vertical border is drawn in the same grey style as the top and bottom borders, and the content text retains its own bright styling

#### Scenario: Bottom border leads the right side with the help hint

- **WHEN** the article view is rendered
- **THEN** the bottom border's right side reads `?: help • <percent>% •
  <bottomLine>/<totalLines>`, with `?: help` as the first right-aligned element

#### Scenario: Bottom border shows help hint and line indicator

- **WHEN** the article view is rendered
- **THEN** the bottom border shows `o: open article in browser` left-aligned and `<percent>% • <bottomLine>/<totalLines>` right-aligned, with fill dashes between them

#### Scenario: ASCII fallback border

- **WHEN** ASCII fallback mode is enabled or the terminal lacks box-drawing support
- **THEN** single-line border glyphs are replaced with ASCII equivalents (`-`, `+`), the left and right vertical borders are both `:` rendered in the grey style, and the scrollbar thumb is `|` rendered in the bright/white style (the unfilled track stays `:` in grey)

### Requirement: Article top border title truncation

The system SHALL render the article's date and title inline with the top border
with the date always fully visible and the title left-aligned after the date.
The date prefix SHALL be `YYYY-MM-DD HH:MM` showing the publication date and
time in the user's local timezone by default, or `YYYY-MM-DD HH:MM:SS` when the
`show_seconds` display option is enabled, followed by the bullet
(`•`, ASCII
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
- **THEN** the date-and-time prefix remains fully visible, the title is truncated and ends with `…`, and a single dash precedes the right corner (for example `┌─ 2026-01-02 15:04 • Hello world artic… ─╖`)

#### Scenario: Top border shows seconds when enabled

- **WHEN** the `show_seconds` display option is enabled and an article is displayed
- **THEN** the top border date prefix renders as `YYYY-MM-DD HH:MM:SS` (for example `┌─ 2026-01-02 15:04:05 • Hello world artic… ─╖`)

#### Scenario: ASCII fallback ellipsis

- **WHEN** ASCII fallback mode is enabled and the article title does not fit
- **THEN** the truncated title ends with `...` instead of `…`

#### Scenario: Date and time stay visible for a very long title

- **WHEN** the article title is longer than the entire top-border content width
- **THEN** the `YYYY-MM-DD HH:MM` (or `YYYY-MM-DD HH:MM:SS` when seconds are enabled) date prefix is still rendered in full and only the title is truncated with an ellipsis

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation
(PgUp/PgDn, Ctrl-F/Ctrl-B, Space) SHALL scroll by a full viewport page. Half-page
navigation (Ctrl-D/Ctrl-U) SHALL scroll by half a viewport. `1`, `g`, or
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
The system SHALL provide an `x` key in the list view that toggles all day groups
at once: when every day group is expanded it SHALL collapse them all, and
otherwise (some or all collapsed, including a mixed state) it SHALL expand them
all.

#### Scenario: Open article from list

- **WHEN** the user presses Enter, `l`, or `o` on a selected article row in the list view
- **THEN** the article reader view opens displaying that article

#### Scenario: Fold keys on a day header

- **WHEN** the list cursor is on a day header and the user presses `h`, `l`, Enter, Space, or Tab
- **THEN** `h` collapses that day, `l` expands it, and Enter/Space/Tab toggle its expansion; no article is opened

#### Scenario: Tab toggles the current day group from an article row

- **WHEN** the list cursor is on an article row and the user presses Tab
- **THEN** the day group containing that article toggles its expansion

#### Scenario: Toggle all collapses an all-expanded list

- **WHEN** every day group is expanded in the list view and the user presses `x`
- **THEN** every day group collapses and only the day headers remain visible

#### Scenario: Toggle all expands a fully collapsed list

- **WHEN** every day group is collapsed in the list view and the user presses `x`
- **THEN** every day group expands and all articles become visible

#### Scenario: Toggle all expands a mixed list

- **WHEN** some day groups are expanded and others are collapsed in the list
  view and the user presses `x`
- **THEN** every day group expands (a mixed state defaults to expand all)

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

### Requirement: Read-status on open and movement

The system SHALL mark an article as read when the user opens it and the entire
article content fits within the viewport, since no scrolling is possible. If
the article content does not fit entirely within the viewport, the system SHALL
mark it read upon the user's first scroll or jump action through the article —
a line, page, or half-page scroll in either direction, a jump to the top or
bottom, or a mouse-wheel scroll in either direction — regardless of whether the
action actually changes the scroll position. Once marked read, further movement
SHALL NOT change the read status. Restoring a previously saved scroll offset
when an article is re-opened SHALL NOT count as movement.

#### Scenario: Short article marked read on open

- **WHEN** the user opens an article whose full content fits in the viewport
- **THEN** the article's read status is set to true immediately

#### Scenario: Long article not marked read on open

- **WHEN** the user opens an article whose content exceeds the viewport height
- **THEN** the article's read status remains false

#### Scenario: Long article marked read on first movement

- **WHEN** a long article is open and the user scrolls in either direction, or
  jumps to the top or bottom
- **THEN** the article's read status is set to true immediately

#### Scenario: A scroll action marks read regardless of position change

- **WHEN** the user issues a scroll action while reading a long article
- **THEN** the article is marked read even if the scroll position does not
  actually change

#### Scenario: Restoring a saved offset is not movement

- **WHEN** the app re-opens an article at a previously saved scroll offset
- **THEN** the article is not marked read until the user scrolls or jumps

#### Scenario: Further movement does not change read status

- **WHEN** an article is already read and the user scrolls it
- **THEN** the article remains read

### Requirement: Tag popup on list view

The system SHALL open a tag popup when the user presses `T` on the list view.
The popup SHALL cover the screen except for a margin of 12 columns on the left
and right and 2 rows at the top and bottom; on a terminal too small to honor
the margins, the popup SHALL fall back to a minimal usable box within the
terminal bounds rather than a zero-sized or negative-sized popup. The popup's
title SHALL be `Tags & Sources`, inlined into the center of the top border
rather than rendered
as a body line, and the box SHALL keep a single space of padding on the left
and right with no top or bottom padding. The bottom border SHALL right-align a
position indicator in the form `<percent>% • <nth>/<total>` — the selected
tag's place among all tags — mirroring the article view's bottom border and
the list view's status bar; the bottom border line and its bullet SHALL
render in the same chrome role as the rest of the popup border. The popup
SHALL lay out all tags in a multi-column grid in Windows Explorer list style:
source-domain (news organization) tags SHALL be grouped first, then category
tags, each group ordered by popularity (total article count descending) with
equal-popularity tags ordered alphabetically (case-insensitive, so `apple`
sorts before `Zoo`). The
source-domain group SHALL be introduced by a non-selectable `Sources` subtitle
and the category group by a non-selectable `Tags` subtitle, each shown only
when its group is non-empty, rendered in the status-bar teal role (non-bold),
and occupying a single in-column grid cell, centered within its column, laid
out in the same column-major
flow as the tags so a subtitle never carries the selection cursor and the
cursor never lands on one; tags
fill the
first column top to bottom, continuing into the next column when the bottom of
the grid is reached. Each column SHALL be sized to fit its own widest entry
(tag name and counts), capped at 23 columns, so that one long
entry does not widen every column; an entry wider than its column SHALL have
its middle elided with an ellipsis — `…` in Unicode mode and `...` in ASCII
fallback mode — keeping the tag's start and end readable. Tags SHALL carry no
cursor gutter: each tag name SHALL sit one column from the popup's content
edge. Multiple columns
SHALL be visible side by side whenever they fit, with two spaces between
columns. When the natural grid is narrower than the popup's interior, the spare
space on the right SHALL widen the columns — growing them to one equal width
that fills the interior, or, when a single long entry forces a column wider
than that equal width, growing every column by an equal share of the spare
instead — and any leftover space that cannot be split into whole columns SHALL
be removed by pulling the popup's right border in so the grid never leaves
space to spare. When the
full grid is wider than the popup's interior, the popup SHALL scroll
horizontally as the cursor moves toward the least popular tags, keeping the
selected tag visible; the window SHALL advance right only when the cursor
passes the window's right edge and left only when it passes the window's left
edge, staying put while the cursor moves within it (symmetric in both
directions); when the cursor wraps from the last column back to the
first, the popup SHALL scroll back to the start. The visible window SHALL
likewise fill the popup's interior — its columns widened equally, with any
leftover space removed by pulling the right border in — so no trailing space
appears between the right-most visible column and the border. Each tag SHALL
display its
unread/total counts as `X/Y` (no parentheses), right-aligned to the tag cell's
right edge with the tag name left-aligned and any leftover column space padding
between them. When read count is zero (all unread) or
unread count is zero (all read), the counts SHALL collapse to a single total
number. The unread count SHALL always be rendered in bold bright and the total
in the text (white) role, matching the list view's read/unread colors. The
selected tag SHALL be highlighted with the list view's full-cell background
highlight; the popup SHALL NOT render a `>` cursor marker. Navigation SHALL
be four-directional:
`j`/`k` and the up/down arrow keys SHALL move along the column-major order,
wrapping from the last tag to the first tag and from the first tag to the last
tag; `h`/`l` and the left/right arrow keys SHALL move between columns on the
same row, wrapping from the first column to the last column and from the last
column to the first column; when a horizontal wrap enters a final column that
holds fewer rows than the full columns, the cursor SHALL clamp to that column's
last occupied row. `1` or `g` SHALL move to the first tag (the most popular,
top of the first column) and `G` SHALL move to the last tag (the least
popular, last occupied row of the last column).

#### Scenario: Open tag popup from list

- **WHEN** the user presses `T` on the list view
- **THEN** a popup appears covering the screen except for a 12-column margin on the left and right and a 2-row margin at the top and bottom, listing all tags with source-domain tags grouped first

#### Scenario: Source-domain tags grouped ahead of categories

- **WHEN** the tag popup is opened and stored articles come from `https://www.nytimes.com` and `https://www.theguardian.com` alongside category tags
- **THEN** the `nytimes` and `theguardian` tags are listed first, ahead of the category tags, each with its article count

#### Scenario: Title is inline in the top border

- **WHEN** the tag popup is displayed
- **THEN** the title `Tags & Sources` is centered in the top border, no title line appears in the popup body, and the first grid row sits directly beneath the top border with no vertical padding

#### Scenario: Sources subtitle groups the news organizations

- **WHEN** the tag popup displays source-domain tags alongside category tags
- **THEN** a `Sources` subtitle renders above the source-domain tags and a `Tags` subtitle above the category tags, both in the non-bold status-bar teal role, centered in a single in-column grid cell

#### Scenario: Subtitle only shown for a non-empty group

- **WHEN** the tag popup displays only category tags (no source-domain tags)
- **THEN** the `Tags` subtitle renders above them and no `Sources` subtitle appears

#### Scenario: Section subtitles are not selectable

- **WHEN** the tag popup is displayed and the user navigates
- **THEN** the selection cursor never lands on a `Sources` or `Tags` subtitle, and the subtitles render without the selection background highlight

#### Scenario: Bottom border shows the position indicator

- **WHEN** the tag popup is displayed with the 20th tag of 40 selected
- **THEN** the bottom border right-aligns `50% • 20/40`, the selected tag's place among all tags, with the border line and bullet rendered in the same chrome role as the rest of the popup border

#### Scenario: Selection is a background highlight

- **WHEN** a tag is selected in the tag popup
- **THEN** the selected cell is highlighted with the list view's selection background across the full column width — the width of the column's longest entry including its counts — regardless of the selected tag's own name length, and no `>` marker is shown

#### Scenario: Count colors match the list view

- **WHEN** a tag displays a mixed unread/total count
- **THEN** the unread number renders bold bright and the total in the text (white) role, matching the list view's read/unread colors

#### Scenario: Grid fills top to bottom then continues to the next column

- **WHEN** the tag popup displays more tags than fit in one column
- **THEN** the tags fill the first column top to bottom in order and the next column continues from the top with the following tags

#### Scenario: Long tag names are middle-elided at the column cap

- **WHEN** the tag popup displays a tag entry wider than its column's width
- **THEN** the tag's middle is elided with an ellipsis, keeping its start and end readable

#### Scenario: ASCII fallback ellipsis in elided tags

- **WHEN** ASCII fallback mode is enabled and a tag entry is middle-elided
- **THEN** the elided entry shows `...` instead of `…`

#### Scenario: Multiple columns are visible side by side

- **WHEN** the tag popup displays more than one column whose combined widths fit the popup interior
- **THEN** two or more columns render side by side, each sized to its own widest entry

#### Scenario: Spare space widens the columns to fill the interior

- **WHEN** the natural grid is narrower than the popup's interior
- **THEN** the columns grow to equal widths that fill the interior, keeping two spaces between columns, so the grid leaves no space on the right

#### Scenario: Leftover space pulls the right border in

- **WHEN** the widened grid cannot consume all the spare space in whole columns (for example a remainder of one)
- **THEN** the popup's right border pulls in by the leftover amount so the grid still fills the interior with no trailing space

#### Scenario: Scrolling window also fills the interior

- **WHEN** the full grid is wider than the popup's interior and a horizontal window of columns is visible
- **THEN** the visible columns are widened equally to span the interior, with any leftover space pulling the right border in, so no trailing space appears between the right-most visible column and the border

#### Scenario: Grid scrolls right to keep the selected tag visible

- **WHEN** the tag grid is wider than the popup's interior and the user moves the cursor into a column beyond the right edge of the visible area
- **THEN** the popup scrolls right so the selected column is visible, and when the cursor wraps from the last column back to the first the popup scrolls back to the start

#### Scenario: Window only scrolls left when the cursor passes its left edge

- **WHEN** the tag grid is wider than the popup's interior and the user moves the cursor left from the last column
- **THEN** the window stays put while the cursor moves within it, scrolling left only when the cursor passes the window's left edge — symmetric with the rightward behavior

#### Scenario: Down wraps from the last tag to the first

- **WHEN** the user presses down arrow or `j` while the last tag (last occupied row of the last column) is selected
- **THEN** selection moves to the first tag (top of the first column)

#### Scenario: Up wraps from the first tag to the last

- **WHEN** the user presses up arrow or `k` while the first tag (top of the first column) is selected
- **THEN** selection moves to the last tag (last occupied row of the last column)

#### Scenario: Down continues into the top of the next column

- **WHEN** the user presses down arrow or `j` while the bottom tag of a full column is selected
- **THEN** selection moves to the top of the next column

#### Scenario: Left wraps from the first column to the last

- **WHEN** the user presses the left arrow key or `h` while a tag in the first column is selected
- **THEN** selection moves to the same row of the last column

#### Scenario: Right wraps from the last column to the first

- **WHEN** the user presses the right arrow key or `l` while a tag in the last column is selected
- **THEN** selection moves to the same row of the first column

#### Scenario: Vim left and right keys move between columns

- **WHEN** the user presses `h` or `l` in the tag popup
- **THEN** `h` moves the cursor one column left and `l` moves it one column right, wrapping and clamping exactly like the arrow keys

#### Scenario: Horizontal wrap clamps into a partial final column

- **WHEN** a horizontal wrap targets a row that does not exist in the destination column because that column holds fewer tags than the full columns
- **THEN** selection lands on the last occupied row of that destination column

#### Scenario: First and last tag jumps

- **WHEN** the user presses `1` or `g` in the tag popup
- **THEN** the first tag (most popular, top of the first column) is selected
- **WHEN** the user presses `G` in the tag popup
- **THEN** the last tag (least popular, last occupied row of the last column) is selected

#### Scenario: Tiny terminal falls back to a minimal popup

- **WHEN** the terminal is smaller than the margins require (for example 20 columns by 8 rows) and the tag popup is opened
- **THEN** the popup renders as a minimal usable box within the terminal bounds and the application does not panic

#### Scenario: All-unread tag collapses to total

- **WHEN** a tag has 0 read articles and 60 unread articles
- **THEN** the tag displays as `tagname 60` with 60 in bold

#### Scenario: All-read tag collapses to total

- **WHEN** a tag has 60 read articles and 0 unread articles
- **THEN** the tag displays as `tagname 60` with 60 in plain text

#### Scenario: Mixed tag shows unread/total

- **WHEN** a tag has 45 read articles and 15 unread articles out of 60 total
- **THEN** the tag displays as `tagname 15/60` with 15 in bold and 60 in plain text

#### Scenario: Selecting a tag filters the list

- **WHEN** the user selects a tag in the popup and confirms
- **THEN** the list view closes the popup and displays only articles with that tag

### Requirement: Tag popup on article view

The system SHALL open a tag popup when the user presses `T` on the article
view, using the same geometry, title-in-border presentation (`Tags & Sources`),
padding, count colors, selection highlight, multi-column grid layout,
source-first grouping with non-selectable `Sources`/`Tags` section subtitles,
horizontal scrolling, and four-direction wrap-around navigation as the
list-view tag popup: the popup covers the screen except for a 12-column margin
on the left and right and a 2-row margin at the top and bottom, tags fill
columns top to bottom, names wider than their column are middle-elided at the
23-column cap, and `h`/`j`/`k`/`l`, the arrow keys, and `1`/`g`/`G` navigate
as in the list-view popup. The popup SHALL list only the
categories assigned to the current article, ordered by popularity (with
equal-popularity tags ordered alphabetically, case-insensitive). Selecting a
tag SHALL filter the list view and navigate back to it.

#### Scenario: Open tag popup from article

- **WHEN** the user presses `T` on the article view
- **THEN** a popup appears covering the screen except for a 12-column margin on the left and right and a 2-row margin at the top and bottom, listing only the current article's tags

#### Scenario: Article popup grid and navigation follow the list popup

- **WHEN** the article-view tag popup is displayed
- **THEN** its tags fill columns top to bottom with source-domain tags grouped first under a `Sources` subtitle and category tags under a `Tags` subtitle (both non-selectable), names wider than their column are middle-elided at the 23-column cap, the title `Tags & Sources` is inline in the top border, selection is a background highlight, and `h`/`j`/`k`/`l`, the arrow keys, and `1`/`g`/`G` navigate with the same wrap-around and horizontal-scroll behavior as the list-view tag popup

#### Scenario: Selecting a tag from article popup filters list

- **WHEN** the user selects a tag in the article-view popup
- **THEN** the system returns to the list view filtered by that tag

### Requirement: Help popup

The system SHALL display a help popup listing each action and its bound keys
when the user presses `?`. Each action SHALL be labeled with its display name
derived from the action identifier by replacing underscores with spaces and
rendering the result in Title Case (for example `half_page_down` →
`Half Page Down`). The action label SHALL appear on the left
and the bound key strings SHALL appear on the right. The popup SHALL present
the bindings grouped under three headings in this order: **Global**, **List
view**, and **Article view**. Actions not specific to a single view — those
valid in all views and those valid in both the list and article views — SHALL
be listed under Global. Actions valid only in the list view SHALL be listed
under List view, and actions valid only in the article view SHALL be listed
under Article view. Within each section the actions SHALL appear in catalog
order, and each section heading SHALL be rendered in the teal role. The List
view and Article view sections SHALL be laid out in a second column to the
right of the Global section, and the popup including its border SHALL NOT
exceed 70 columns wide.

#### Scenario: Action labels are Title Case without underscores

- **WHEN** the help popup is displayed
- **THEN** each action is labeled with underscores replaced by spaces and each word capitalized (for example `Open Article`, `Half Page Down`)

#### Scenario: Help popup lists current bindings

- **WHEN** the user presses `?` in any view
- **THEN** a popup appears listing each action's display name on the left and its currently bound keys on the right

#### Scenario: Help popup groups bindings by scope

- **WHEN** the user presses `?` in any view
- **THEN** the popup lists Quit, Back, Move Up, Page Down, and Tag Popup under Global, Refresh and Open Article under List view, and Close Article, Link Popup, Open URL, Copy URL, and Copy Article Text under Article view

#### Scenario: Global includes all-view and shared bindings

- **WHEN** the Global section is rendered
- **THEN** it lists every action that is not specific to a single view, including actions valid in all views (such as Quit, Back, Move Up, Move Down, Help) and navigation shared by the list and article views (such as Page Up, Page Down, Top, Bottom, Tag Popup)

#### Scenario: Section headings use the teal role

- **WHEN** a section heading is rendered
- **THEN** it renders in the teal role (ANSI 14 in dark mode, ANSI 6 in light mode)

#### Scenario: List and article sections form a second column

- **WHEN** the help popup is displayed
- **THEN** the List view and Article view sections render in a second column to the right of the Global section

#### Scenario: Help popup stays within seventy columns

- **WHEN** the help popup is rendered with its default keybindings
- **THEN** the popup including its border is at most 70 columns wide

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

The system SHALL quit on `q` or Ctrl-C from the list and article views. When a
popup is open, `q` SHALL close the popup and return to the underlying view
instead of quitting, while Ctrl-C still quits from inside a popup. The system SHALL provide `m` to toggle read/unread status, `a` to mark all articles in the current view as read, `o` to open the article URL in the system browser, `c` to copy the article URL to the clipboard, `C` to copy the article text to the clipboard, and `?` to display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in the list or article view
- **THEN** the application exits

#### Scenario: Q closes an open popup

- **WHEN** the user presses `q` while a popup is open over the list or article view
- **THEN** the popup closes, the underlying view remains, and the application does not exit

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
preferred background. The renderer SHALL render base (non-styled) body text in
ANSI color 7 (white) in dark mode and ANSI color 0 (black) in light mode,
overriding the theme's default base foreground, while styled elements (bold,
italics, headings, links, blockquotes) keep the theme's own styling. Inline
`<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available; the alt text
SHALL be rendered literally, so markdown special characters in the alt text
(such as `*`, `_`, `[`, or `]`) SHALL NOT be interpreted as styling or link
syntax. A publisher-attached lead image (from `<enclosure>` or
`media:thumbnail`) is rendered separately as an inline halfblocks image block
above the rendered article content and is not affected by this placeholder rule.
`<strong>` and `<b>` SHALL render with the theme's bold styling and
`<em>`/`<i>` with its italic styling, not as literal markdown markers.
Hyperlink rendering SHALL follow the active link display mode, as specified by
the `compact-link-mode` capability. Structured HTML SHALL render
as styled markdown: ordered and unordered
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

#### Scenario: Body text uses the standard white/black base color

- **WHEN** an article is displayed in the reader view in dark (or light) mode
- **THEN** the base body text renders in ANSI color 7 (white) in dark mode and ANSI color 0 (black) in light mode, while bold text, links, and headings keep their theme styling

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

#### Scenario: Expanded mode renders link as styled text followed by the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view in expanded mode
- **THEN** the reader displays "world" styled as a link followed by `https://example.com/post`, with the link wrapped in OSC 8 hyperlink escapes

#### Scenario: Compact mode renders only the link name

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view in compact mode
- **THEN** the reader displays "world" styled as a link without the URL text, with the link wrapped in OSC 8 hyperlink escapes carrying the URL

#### Scenario: Long URL wraps onto a new line in expanded mode

- **WHEN** in expanded mode a link's rendered text does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL wraps across lines in expanded mode

- **WHEN** in expanded mode a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is wrapped across multiple lines and is not truncated or elided

#### Scenario: Conversion failure renders plain text, not raw HTML

- **WHEN** HTML-to-markdown conversion of the article content fails
- **THEN** the reader displays the article's plain text with HTML markup removed, and no `<p>`/`<a>`/`<div>` tag text is shown

#### Scenario: Markdown render failure shows text without markdown markers

- **WHEN** the markdown renderer fails to render the article content
- **THEN** the reader displays the article text without literal markdown formatting markers such as `**` or `[text](url)`

The system SHALL remove U+00AD soft hyphens from the article content before
rendering: the app counts them as zero-width but the terminal renders each as a
visible cell, so a line padded to the content width would overflow the frame by
one column per soft hyphen and wrap the right border (scrollbar) onto the next
line's left.

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

### Requirement: Restore reader state on start

The system SHALL persist the reader's selection when the application quits and
restore it when the application starts again. The persisted selection SHALL
record the current view (list or article), the row under the cursor (the
article by its ID, or the day group by its calendar-day key when the cursor is
on a day header), and, when the article view is open, the article's scroll
offset together with a content anchor for that offset. On startup the system
SHALL move the list cursor to the row matching the saved selection. When the
saved view is the article view and the saved article still exists, the system
SHALL open that article and restore its scroll position, so that returning to
the list (via `b`, `esc`, or `h`) leaves the same article selected.

The content anchor SHALL identify the reading position independently of the
rendered line count, so it survives a restart even when image blocks have not
materialized yet and the initial composition happens at a different size than
the real terminal. When the first visible line of the viewport lies inside a
composed image block, the anchor SHALL be that block; otherwise the anchor
SHALL be the first text line's leading content together with the number of
lines between that line and the viewport top (the gap). On restore, the system
SHALL re-locate the anchored line in the freshly composed article and position
the viewport so the anchored line sits at its saved gap from the top; when the
anchor is an image block, the system SHALL position the viewport on the block's
placeholder row. When the restored anchor was chosen at a size that differs
from the real terminal size, the system SHALL re-apply the anchor once the true
terminal size is known, so the reading position lands on the same line the user
was reading. When no usable anchor can be located, the system SHALL fall back
to the saved raw offset. When the saved selection no longer exists — for
example the article was pruned or the day group changed — the system SHALL fall
back to the default clamped cursor without error.

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

#### Scenario: Restore lands on the anchored line before blocks materialize

- **WHEN** the user quits while reading partway down an article that contains
  image blocks whose stored blocks do not match the restart content width, and
  restarts
- **THEN** the reopened article lands on the same content line the user was
  reading, even though the image blocks render as placeholders initially

#### Scenario: Restore re-applies the anchor at the real terminal size

- **WHEN** the reader restores an article at a size that differs from the real
  terminal size and the first real window-size message then arrives
- **THEN** the anchor is re-applied at the real size so the reading position
  stays on the line the user was reading

#### Scenario: Restore falls back to the raw offset without an anchor

- **WHEN** the saved article contains no text line or image block that the
  anchor can locate
- **THEN** the system restores the saved raw scroll offset

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
cells with an inverted (highlighted) style distinct from unselected text, and
the inverted style SHALL remain continuous across hyperlinks and styled text
within the selection rather than resetting at the first link on a line. Copied
text SHALL be the plain text characters within the selected region,
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

#### Scenario: Selection highlights continuously across a link

- **WHEN** the selected region spans a hyperlink and the surrounding text on the same line
- **THEN** the inverted highlight covers the entire selected region including the link, with no un-highlighted gap and no early reset at the link boundary

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
the article body (the article's own URL is not listed) in document order,
deduplicated by URL. Each row SHALL show the link text followed by the URL
rendered in the dim style and truncated to the popup width. Navigation SHALL
use `j`/`k` and the arrow keys and SHALL wrap around at both ends. Enter SHALL
open the selected URL in the system browser using the same non-shell
single-argument opener used for other URL opens. While the links popup is open,
`o` SHALL still open the article's own URL in the system browser without closing
the popup. Esc or `q` SHALL close the popup
without navigating. When the article contains no links, the popup SHALL open
with an empty list and Enter SHALL do nothing. While the popup is closed,
Enter SHALL continue to close the article view and return to the list.

#### Scenario: Open links popup from article

- **WHEN** the user presses `l` or the right arrow key in the article view
- **THEN** a popup opens listing the article's links in document order

#### Scenario: Row shows link text and dimmed URL

- **WHEN** the article contains `<a href="https://example.com/post">world</a>` and the links popup is open
- **THEN** the popup shows a row with the text `world` followed by the URL `https://example.com/post` in the dim style, truncated to the popup width

#### Scenario: Article's own URL is not listed

- **WHEN** an article with a link URL is open and the links popup is opened
- **THEN** the article's own URL does not appear as a row in the popup; only the article body's hyperlinks are listed

#### Scenario: Duplicate URLs are listed once

- **WHEN** the same URL is rendered twice in the article (for example a link whose text wraps across lines)
- **THEN** the popup lists that URL exactly once

#### Scenario: Navigation wraps around

- **WHEN** the user presses `j` (or down) while the last link row is selected
- **THEN** selection moves to the first row

#### Scenario: Enter opens the selected link

- **WHEN** the user presses Enter on a selected link row in the popup
- **THEN** the selected URL is opened in the system browser via a non-shell single-argument opener

#### Scenario: O opens the article URL while the popup is open

- **WHEN** the links popup is open and the user presses `o`
- **THEN** the article's own URL is opened in the system browser and the popup remains open

#### Scenario: Esc closes the popup without navigating

- **WHEN** the user presses Esc while the links popup is open
- **THEN** the popup closes, the article view remains, and no URL is opened

#### Scenario: Empty article shows an empty popup

- **WHEN** the article contains no links and the user opens the links popup
- **THEN** the popup opens with an empty list and Enter does nothing

### Requirement: Standard terminal color roles

The system SHALL derive all UI foreground colors from the standard terminal
(ANSI) palette instead of bespoke RGB values. The following color roles SHALL
apply across the list, reader, and popup views:

- **grey/dim role** — the colour `#707070` in both dark and light modes:
  the article border, rails, bullets, scrollbar track, image
  attribution, and muted popup text (such as plain tag counts and link URLs).
- **teal role** — ANSI color 14 (bright cyan) in dark mode and ANSI color 6
  (cyan) in light mode: the status bar, day headers, the inline text in the
  article border (date, title, help hint, position indicator), the scrollbar
  thumb, and popup titles and popup borders.
- **text role** — ANSI color 7 (white) in dark mode and ANSI color 0 (black)
  in light mode: the article body text and the list view's article rows (read
  article titles, timestamps, and source identifiers).
- **bright role** — ANSI color 15 (bright white) in dark mode and ANSI color 0
  (black) with bold weight in light mode: unread article titles and bold tag
  counts.

The selected-row background highlight SHALL be theme-aware: `#242424` in dark
mode and `#d3d3d3` in light mode, so selected text remains readable in both
modes (white on `#242424`, black on `#d3d3d3`).

#### Scenario: Border and muted text share the grey role

- **WHEN** the reader or list view is rendered
- **THEN** the article border, rails, bullets, scrollbar track,
  image attribution, and popup URLs render in `#707070`

#### Scenario: Status, day headers, inline border text, scrollbar thumb, and popup chrome use the teal role

- **WHEN** the status bar, a day header, the article border's inline text, the scrollbar
  thumb, or a popup's title or border is rendered in dark mode
- **THEN** it renders in ANSI color 14 (bright cyan), and in ANSI color 6
  (cyan) in light mode

#### Scenario: Unread titles use the bright role

- **WHEN** an unread article title is rendered in the list view in dark mode
- **THEN** the title renders in ANSI color 15 (bright white) with bold weight,
  and in light mode it renders in ANSI color 0 (black) with bold weight

#### Scenario: Article text uses the text role

- **WHEN** article body text or a list article row (read title, timestamp, or
  source identifier) is rendered in dark mode
- **THEN** it renders in ANSI color 7 (white), and in ANSI color 0 (black) in
  light mode

#### Scenario: Selection highlight is theme-aware

- **WHEN** a row is selected in the list view in dark mode
- **THEN** the selection highlight renders with the `#242424` background, and
  with the `#d3d3d3` background in light mode
### Requirement: No-op article scrolls do not re-render

The system SHALL memoize the article reader's rendered frame and SHALL NOT
re-render it when a scroll action produces no change in scroll position. A
no-op scroll is one that leaves the viewport offset unchanged — for example
scrolling the mouse wheel up while already at the top of the article, or
scrolling down while already at the bottom — so the article's viewport
composition, native-image re-show and clipped-transmit transforms, and border
rendering are not re-run for that input. The frame the system produces after a
no-op scroll SHALL be byte-identical to the frame it produced before the
no-op. A scroll that changes the scroll position, or any other change to the
state the article view renders (a resize, an image-load recomposition, a
selection change, a glyph-mode change, a link-mode change, or a native-image
transmit transition), SHALL re-render the frame and update the cached frame.

#### Scenario: Wheel up at the top does not re-render

- **WHEN** the article view is scrolled to the top (offset zero) and the user
  scrolls the mouse wheel up
- **THEN** the scroll is a no-op, the article frame is not re-rendered, and the
  rendered output is unchanged

#### Scenario: Wheel down at the bottom does not re-render

- **WHEN** the article view is scrolled to the bottom and the user scrolls the
  mouse wheel down
- **THEN** the scroll is a no-op, the article frame is not re-rendered, and the
  rendered output is unchanged

#### Scenario: No-op frame is byte-identical

- **WHEN** the article view renders once and then receives a no-op scroll
- **THEN** the frame rendered after the no-op is byte-identical to the frame
  rendered before it

#### Scenario: A real scroll re-renders

- **WHEN** the article view is scrolled away from a boundary and the user
  scrolls the mouse wheel in a direction that changes the scroll position
- **THEN** the article frame is re-rendered and the cached frame is updated to
  reflect the new scroll offset

#### Scenario: Image recomposition re-renders

- **WHEN** an image load lands and re-composes the open article while the
  reader is at a boundary scroll position
- **THEN** the article frame is re-rendered to include the newly available
  image block rather than served from the stale cache

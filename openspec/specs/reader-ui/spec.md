## Purpose

Provides the terminal user interface for browsing, reading, and filtering RSS articles — comprising the article list view, the article reader view with a scroll-progress border, and the tag popup for category-based filtering.

## Requirements

### Requirement: Article list rendering

The system SHALL display articles grouped under a collapsible header for each
day, with the most recent day first and articles within a day sorted reverse
chronologically. The day-group headers SHALL be rendered with a tree rail down
the left edge: the first day group SHALL be prefixed with `┌` (ASCII `+`), the
last day group with `└` (ASCII `+`), and every other day group with `├`
(ASCII `+`), each followed by a single space. Article rows SHALL NOT carry a
tree glyph; the publication time SHALL start the row. Day
headers SHALL NOT display fold markers. Each day header SHALL display the long
date in the form `Weekday Day-ordinal Month, Year` (for example
`Saturday 20th August, 2025`), computed in the user's local timezone, where the
day is suffixed with an English ordinal (`1st`, `2nd`, `3rd`, `4th`, …). The
header for the current local calendar day SHALL be prefixed with `today, ` and
the header for the previous local calendar day SHALL be prefixed with
`yesterday, `; older days and the `Undated` group SHALL be unprefixed. Each
article row SHALL show the publication time as `HH:MM` in the user's local
timezone (`--:--` when undated) at the start of the row, followed by the
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
- **THEN** the first day header is prefixed with `┌`, the last day header with `└`, and every day header between with `├`

#### Scenario: Day headers carry the rail, article rows do not

- **WHEN** the list is displayed
- **THEN** each day header is prefixed `┌`, `├`, or `└` followed by a single space, and each article row carries no tree glyph, starting with its publication time

#### Scenario: Tree glyphs are stable while scrolling

- **WHEN** the visible window scrolls so the list's first and last day groups are off screen
- **THEN** the visible day headers are all prefixed `├` and the bookend glyphs stay on the first and last day groups rather than moving to the window edges

#### Scenario: ASCII fallback tree glyphs

- **WHEN** ASCII fallback mode is enabled
- **THEN** the day-header tree glyphs render as `+` (corners and tees) and article rows carry no tree glyph

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
- **THEN** the row shows `HH:MM` at the start of the row followed by the title, with no bullet markers anywhere in the row

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
(for example `14:30 Saturday 29th August, 2026 last refresh`), with the
`last refresh` label rendered in the dim role (#707070), or the text
`never refreshed` when no refresh has occurred. The right-most elements SHALL
be a literal `?: help` hint, followed by the bullet and the on-disk size of the
article database as `DB <n>MB`, the scroll position percentage, and the article
count as `<n>/<total>` — where the database size is rendered as a whole number
of megabytes (floored to an integer, with no space between the number and `MB`)
and `<total>` is the total number of articles in the database — joined by the
middle-dot bullet (`·`, ASCII `.`) (for example
`?: help · DB 320MB · 100% · 2/2`). The `DB` value SHALL be the on-disk database
size including the WAL sidecar file when one is present. Values below one
megabyte SHALL render as `0MB`. The bullets and the `last refresh` label SHALL
render in the dim role (#707070); the remaining status bar text SHALL render in
the chrome role. When a tag filter is active, the status bar SHALL also display
the active filter name.

#### Scenario: Status bar leads the right side with the help hint

- **WHEN** the list view is displayed
- **THEN** the status bar's right side reads `?: help · DB <n>MB · <percent>% ·
  <n>/<total>`, with `?: help` as the first right-aligned element

#### Scenario: Status bar shows refresh time, size, percentage, and position

- **WHEN** the list view is displayed
- **THEN** the status bar shows the last-refresh time and date on the left as
  `HH:MM Weekday Day-ordinal Month, Year last refresh` and, on the right, the
  database size as `DB <n>MB`, the percentage, and the `<n>/<total>` position
  joined by ` · `

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
`?: help · <percent>% · <bottomLine>/<totalLines>`, where `<percent>` is the
scroll percentage, `<bottomLine>` is the line number of the last visible
viewport line clamped to the total, and `<totalLines>` is the article's total
line count; horizontal dashes fill the space between the hint and the indicator
(for example at the very bottom:
`o: open article in browser ───── ?: help · 100% · 120/120`). The bullets used
between the help affordance and the percent and between the percent and the
line ratio SHALL be the same middle-dot bullet used in the top border (`·`,
ASCII `.`). The content area SHALL have two spaces of
horizontal padding on each side and one space of vertical padding below the
content and none above it, so the first content line (the article URL) sits
directly beneath the top border. When the entire article fits within the viewport
and no scrolling is
possible, the right border SHALL be fully filled with the bright single-line
glyph and the bottom border SHALL display `100%` with `<bottomLine>` equal to
`<totalLines>`.

#### Scenario: Short article fully visible

- **WHEN** an article is opened that fits entirely within the viewport
- **THEN** the top border displays the date and title, the right border is fully filled with the bright single-line glyph, and the bottom border displays the help hint on the left and `100% · <n>/<n>` on the right where `<n>` is the total line count

#### Scenario: Scrollbar visible on open

- **WHEN** a scrollable article (content taller than the viewport) is opened and the scroll offset is zero
- **THEN** the right border renders a bright single-line thumb at least one row tall at the top of the track, with the remainder of the track as a grey single line, and the bottom border displays `0% · <viewportHeight>/<totalLines>`

#### Scenario: Thumb moves with scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to 42% of the scrollable range
- **THEN** the thumb's top row is at 42% of the thumb's travel range along the track (rounded), rendered as a bright single-line, with the rest of the track as a grey single line, and the bottom border displays `42%` and a `<bottomLine>/<totalLines>` ratio consistent with that offset

#### Scenario: Thumb height stays within bounds

- **WHEN** a scrollable article is rendered with any combination of total content height and viewport height
- **THEN** the thumb height is at least one row and at most one row less than the right border's track height

#### Scenario: Thumb reaches the bottom at full scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to the very bottom (100% of the scrollable range)
- **THEN** the thumb sits at the bottom of the track as a bright single-line, with the rest of the track as a grey single line, and the bottom border displays `100% · <totalLines>/<totalLines>`

#### Scenario: Configurable scrollbar style

- **WHEN** the `scrollbar` display option is set to `double` and an article is displayed in Unicode mode
- **THEN** the thumb renders as the bright double-line glyph `║`; with the default `single` it renders as the single line `│`

#### Scenario: Left border uses the grey border style

- **WHEN** the article view is rendered
- **THEN** the left vertical border is drawn in the same grey style as the top and bottom borders, and the content text retains its own bright styling

#### Scenario: Bottom border leads the right side with the help hint

- **WHEN** the article view is rendered
- **THEN** the bottom border's right side reads `?: help · <percent>% ·
  <bottomLine>/<totalLines>`, with `?: help` as the first right-aligned element

#### Scenario: Bottom border shows help hint and line indicator

- **WHEN** the article view is rendered
- **THEN** the bottom border shows `o: open article in browser` left-aligned and `<percent>% · <bottomLine>/<totalLines>` right-aligned, with fill dashes between them

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

The system SHALL open a tag popup when the user presses `T` on the list view.
The popup SHALL cover the screen except for a margin of 12 columns on the left
and right and 2 rows at the top and bottom; on a terminal too small to honor
the margins, the popup SHALL fall back to a minimal usable box within the
terminal bounds rather than a zero-sized or negative-sized popup. The popup's
title SHALL be `Tags & Sources`, inlined into the center of the top border
rather than rendered
as a body line, and the box SHALL keep a single space of padding on the left
and right with no top or bottom padding. The bottom border SHALL right-align a
position indicator in the form `<percent>% · <nth>/<total>` — the selected
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
when its group is non-empty, rendered in the status-bar blue role (non-bold),
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
- **THEN** a `Sources` subtitle renders above the source-domain tags and a `Tags` subtitle above the category tags, both in the non-bold status-bar blue role, centered in a single in-column grid cell

#### Scenario: Subtitle only shown for a non-empty group

- **WHEN** the tag popup displays only category tags (no source-domain tags)
- **THEN** the `Tags` subtitle renders above them and no `Sources` subtitle appears

#### Scenario: Section subtitles are not selectable

- **WHEN** the tag popup is displayed and the user navigates
- **THEN** the selection cursor never lands on a `Sources` or `Tags` subtitle, and the subtitles render without the selection background highlight

#### Scenario: Bottom border shows the position indicator

- **WHEN** the tag popup is displayed with the 20th tag of 40 selected
- **THEN** the bottom border right-aligns `50% · 20/40`, the selected tag's place among all tags, with the border line and bullet rendered in the same chrome role as the rest of the popup border

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
order, and each section heading SHALL be rendered in the blue role. The List
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

#### Scenario: Section headings use the blue role

- **WHEN** a section heading is rendered
- **THEN** it renders in the blue role (ANSI 12 in dark mode, ANSI 4 in light mode)

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
that scrolls with the article text. On a terminal without native image support
the image SHALL be rendered using a Unicode halfblocks representation
(half-block characters with ANSI color) sized to fit the content width while
preserving aspect ratio, and SHALL be horizontally centered within the content
width. The halfblock block SHALL be served from the stored text block when one
exists at the current content width, re-rendered from the in-memory decoded
image when one is cached, or re-fetched and re-rendered when neither is
available. On a full-image-capable terminal the halfblock block SHALL be shown
first as an immediate placeholder and SHALL be replaced by a native inline
photo render once the real photo loads. Directly beneath the image or photo,
with no blank line between them, the system SHALL render a photo attribution
constrained to a standard caption width — a fixed fraction of the content width
(such as 70%) that does not depend on the photo's own width — wrapped onto
additional centered lines when the credit is longer and centered beneath it; a
blank line SHALL follow the attribution before the article content resumes. The
attribution text SHALL be the photo's credit extracted from the article's
stored HTML when present — a `figcaption` associated with the figure containing
the lead image, used only when the lead image is the figure's last `<img>` in
document order (a figure's caption belongs to its last image), otherwise an
element whose class indicates a credit line — and SHALL fall back to `photo:
<source>`, where `<source>` is derived by the same source-identifier rules used
for the list view's organization column; when the lead image sits inside a
captioned figure whose caption belongs to a later image, no attribution SHALL
render at all, with no `photo: <source>` fallback. The
combined image-plus-attribution block height SHALL be capped so that the full
block plus at least one line of body text fits within the viewport height; the
height fit SHALL reserve the wrapped line count of the caption the block
actually composes with, so a long caption cannot overflow the reserved space and
inflate the block past the viewport; when the block is taller than this cap it
SHALL be scaled down to fit. When image rendering is disabled
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
- **THEN** the attribution line is centered beneath the image at the standard caption width and there is no blank line between the image and the attribution

#### Scenario: Attribution wraps at the standard caption width

- **WHEN** the attribution text is longer than the standard caption width (for example a long credit beneath a narrow photo)
- **THEN** the attribution wraps onto additional centered lines within the standard caption width — a fixed fraction of the content width that does not depend on the photo's own width — and the photo is re-rendered smaller so the whole block plus at least one line of body text still fits the viewport

#### Scenario: Attribution rendered at a standard caption width

- **WHEN** a lead image narrower than the content width renders with a long credit
- **THEN** the attribution wraps to the standard caption width (a fixed fraction of the content width), centered beneath the image, rather than wrapping to the photo's own width

#### Scenario: Long caption does not inflate the block past the viewport

- **WHEN** an image's credit is long enough that wrapping it at the standard caption width takes several lines
- **THEN** the image is scaled down so the image plus the full wrapped attribution and one line of body text fit within the viewport height

#### Scenario: Attribution taken from the article HTML

- **WHEN** the article's HTML contains a `figcaption` credit associated with the lead image's figure
- **THEN** the attribution line shows that credit text, centered beneath the photo

#### Scenario: Lead image in a shared captioned figure renders without a caption

- **WHEN** the lead image is not the last `<img>` of a figure whose `figcaption` labels a later image in the same figure
- **THEN** no attribution renders beneath the lead image, with no `photo: <source>` fallback

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

#### Scenario: Stored block renders on open

- **WHEN** image rendering is enabled and an article with a stored image block at the current content width is opened
- **THEN** the stored block renders immediately as the lead image without a network request

#### Scenario: Native terminal swaps block for the photo

- **WHEN** a full-image-capable terminal opens an article and the fetched photo loads
- **THEN** the halfblock placeholder block is replaced by a native inline photo render with the attribution still beneath

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

### Requirement: Atomic image scroll behavior

When an image block (a lead image or an inline image) is rendered, the system
SHALL ensure the photo's lines are either fully visible within the viewport or
fully scrolled out of view; it SHALL never display a partially clipped photo.
Each image block SHALL apply these rules independently, using its own caption.
The snapping applies to single-line moves (the line keys and the mouse wheel);
page, half-page, and goto-bottom moves land wherever they land, even mid-photo,
so paging never skips the text between images.

Each image block SHALL define two snap boundaries. The top boundary is the
offset at which the image's first line sits at the viewport's first row; the
bottom boundary is the offset at which the block's last line — the last line of
the wrapped caption — sits at the viewport's last row. Snapping is defined by
these boundaries and the transitions between them. A block whose bottom
boundary equals its top boundary (it fills the viewport exactly) has a single
snap position.

For the lead image, a downward single-line scroll that would move the viewport
offset into the photo's line range SHALL snap the offset to the photo's
attribution line when one is rendered — landing on the caption so it can be
read — and to the line immediately after the block when no attribution is
rendered, skipping the photo in a single keystroke; a downward scroll that
leaves the lead photo fully visible SHALL likewise snap the offset to the first
non-photo line. The lead image's boundaries are pre-consumed — it was shown on
open — so it has no entry stages: a downward scroll never catches it at the
bottom boundary first.

For an inline image, a downward single-line scroll follows the boundary
transitions: a scroll that leaves the image partially visible in the window —
its top inside the window and its wrapped caption's last line below the fold,
whether it entered from below or a skip of the preceding image left its top
already inside the window — SHALL snap the block's bottom boundary, so the image
and its full wrapped caption are visible with the caption's last line at the
viewport bottom; the next downward scroll SHALL snap the image's first line to
the viewport top (the top boundary); and a scroll landing in its photo range —
including the top-boundary position, where the whole image and caption are on
screen — skips the entire block, landing on the line immediately after the
wrapped caption, so the caption is never left as a standalone position at the
viewport top: the image and its caption are one snap unit. A
downward scroll that would move the offset into the photo's range from above
SHALL snap to the photo's first line (the top boundary); one starting at or
inside it skips the block past its caption onto the line after it. An inline
image whose block is shorter than
the viewport SHALL snap to the bottom boundary (its first line at or inside the
window with the block's last line at the viewport bottom) as soon as a downward
single-line scroll brings its top into the window from below, so a short photo
entered from below is aligned flush rather than scrolled through line-by-line —
the image and its full wrapped caption appear at the viewport bottom; the next
downward scroll SHALL rise it to the top boundary, and the following downward
scroll then skips the whole block onto the line after its caption. The
fully-visible skip SHALL NOT
apply to inline images, so an inline image is never skipped past without first
being shown. Upward single-line scrolls mirror the transitions: an image
entering from above snaps its top to the viewport top (the top boundary) — the
entry snap fires as soon as the block's last line enters the window from above,
even when it appears at the window's first row, so the whole image and its
caption appear at once rather than a caption-only frame — the
next upward scroll snaps its bottom boundary so the caption's last line is at
the viewport bottom, a scroll landing in its photo range reveals it (snaps to
the photo's first line), and an upward scroll that cuts the block's last line
below the fold SHALL snap it fully below the fold on the next move so it scrolls
off the bottom edge cleanly. When an inline image's block fills the viewport
exactly (its image plus wrapped caption), its bottom and top boundaries
coincide, so there is no distinct bottom stage: the upward scroll from the
top-aligned position SHALL scroll the image fully out of view rather than
looping on the same offset. When an exit or scroll-off position would land
inside a preceding image block's photo range (image blocks spaced closer than
the viewport height), the system SHALL snap to the immediately preceding image's
first line instead, revealing it, so consecutive images are each shown in
sequence rather than a nearer image being skipped. The wrapped caption is part
of the image block — displayed with the image at the top and bottom boundaries —
and is not a standalone scroll position. On a full-image-capable terminal, when
the lead photo is fully visible and the user scrolls upward with the offset past
the article top, the system SHALL snap the offset to the article top in a single
step.

When an image load re-composes the article, the system SHALL NOT leave the
viewport offset inside an image block's photo range or with an image's top
materialized mid-window; it SHALL snap the offset to that image's first line.
When the offset is already on the article's last screen (the user moved to the
bottom with `G` or scrolled to the end), re-composition SHALL keep the offset at
the bottom so the article's end stays visible: an image load SHALL NOT snap it
up to a photo top, even when the bottom offset falls inside a photo's range
whose caption is cut off below the fold. On a native-image terminal, a photo's
placement SHALL be drawn only when its rows are fully contained in the viewport
window; a partially visible photo's placement is deleted and its transmit
suppressed so the image never paints over the article border.

#### Scenario: Downward scroll skips the lead photo onto its caption

- **WHEN** the lead image block occupies content lines 4 through 10 (photo lines 4 through 7 plus attribution lines 8 through 10), the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 8 (the attribution's first line): the photo is fully scrolled out and the caption is visible at the top of the viewport

#### Scenario: Inline image snaps to the viewport bottom when its top enters from below

- **WHEN** an inline image block occupies content lines 20 through 24 (photo lines 20 through 22 plus attribution lines 23 through 24), the viewport height is 15, and a downward scroll brings the image's top into view from below
- **THEN** the viewport offset snaps to 10 (the block's bottom boundary, its last line at the viewport bottom) so the image and its full caption are visible at the bottom of the screen

#### Scenario: Wrapped caption with its last line below the fold snaps to the bottom boundary

- **WHEN** an inline image has a caption that wraps to three lines (photo lines 20 through 22, caption lines 23 through 25), the viewport height is 15, and a downward scroll leaves the image's top inside the window with the caption's last line (25) below the fold while its first line (23) is already visible
- **THEN** the viewport offset snaps to 11 (the block's bottom boundary) so the caption's last line (25) sits at the viewport bottom and the full wrapped caption is visible

#### Scenario: Inline image already partially visible snaps to the viewport bottom

- **WHEN** a downward scroll leaves an inline image partially visible in the window — its top inside the window, its last line below the fold — because a skip of the preceding image left its top already inside the window (consecutive tall images are spaced closer than the viewport height)
- **THEN** the viewport offset snaps to put the block's last line at the viewport bottom so the image and caption are fully visible, rather than leaving it partially clipped

#### Scenario: Short inline image entered from below snaps to the viewport bottom

- **WHEN** a downward single-line scroll brings a short inline image's top into the window from below (its block is shorter than the viewport, so it is never partially clipped), the viewport height is 15, and the image's top line is 10
- **THEN** the viewport offset snaps to the block's bottom boundary so the image and its full caption are visible at the bottom of the screen, and the next downward scroll snaps the image's first line to the viewport top (the top boundary)

#### Scenario: Inline image snaps to the viewport top on the next downward scroll

- **WHEN** an inline image is fully visible with its last line at the viewport bottom (at its bottom boundary) and the user scrolls down one line
- **THEN** the viewport offset snaps to the image's first line so the image moves to the top of the viewport

#### Scenario: Inline image skips past its caption when a scroll lands in its photo

- **WHEN** the viewport offset lands inside an inline image's photo range after a downward scroll that started at or inside the image (including the top-boundary position, where the image and its full caption are on screen)
- **THEN** the viewport offset snaps to the line immediately after the wrapped caption, skipping the whole image-and-caption block out of view as one unit

#### Scenario: Upward scroll reveals the whole block at once, never a caption-only frame

- **WHEN** an upward single-line scroll brings an inline image block's last line into the window from above at the window's first row (an image with a caption whose last line would otherwise appear alone at the top of the screen)
- **THEN** the viewport offset snaps to the image's first line so the whole image and its full caption are visible with the image at the top of the screen, rather than showing only the caption's last line

#### Scenario: Upward scroll reveals full inline image

- **WHEN** the viewport offset lands inside an inline image's photo range and the user scrolls up one line
- **THEN** the viewport offset snaps to the photo's first line and the full image and its attribution are visible

#### Scenario: Upward scroll moves a top-aligned inline image to the bottom

- **WHEN** an inline image is fully visible with its top at the viewport top (at its top boundary) and the user scrolls up one line
- **THEN** the viewport offset snaps so the image's bottom boundary holds: the image and its wrapped caption are visible at the bottom of the screen with the caption's last line at the viewport bottom

#### Scenario: Upward scroll snaps a partially-scrolled-off inline image below the fold

- **WHEN** an upward scroll cuts an inline image's last line below the fold, leaving it partially visible at the bottom edge
- **THEN** the next upward scroll snaps the offset so the image's top sits at the fold, scrolling the image fully out of view; a frame where the image pokes only partially into the window renders its halfblock preview rather than a blank strip

#### Scenario: Exit or scroll-off reveals the immediately preceding image

- **WHEN** image blocks are spaced closer than the viewport height and a scroll-off position for one image (its top at the fold) would land inside a preceding image's photo range
- **THEN** the viewport offset snaps to the immediately preceding image's first line so it is revealed rather than left partially clipped, and each image is shown in sequence rather than a nearer one being skipped

#### Scenario: Page-down lands naturally even mid-photo

- **WHEN** the user pages down and the resulting viewport offset lands inside an image block's photo range
- **THEN** no snapping occurs; the offset stays where the page-down landed, even with a partially clipped photo

#### Scenario: Image load does not leave the offset inside a photo

- **WHEN** an inline image finishes loading and the article re-composes while the viewport offset falls inside the image's photo range or with its top within the window
- **THEN** the viewport offset snaps to the image's first line so the image is shown, never left partially clipped

#### Scenario: Image load at the article bottom keeps the end visible

- **WHEN** an image finishes loading and the article re-composes while the user is on the article's last screen (they moved to the bottom with `G` or scrolled to the end), and the bottom offset falls inside a photo's range whose caption is cut off below the fold
- **THEN** the offset stays at the article bottom; it does not snap up to the photo's first line, so the article's end remains visible

#### Scenario: Downward scroll without attribution skips the lead block

- **WHEN** an image block has no attribution, the photo occupies lines 4 through 9, the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 10 (the line after the block) and the photo is fully scrolled out

#### Scenario: Downward scroll skips a fully visible lead photo

- **WHEN** the lead image block occupies content lines 4 through 10, the viewport offset is 2 so the full block is on screen, and the user scrolls down one line
- **THEN** the viewport offset snaps to 8 (the first non-photo line) and the photo is fully scrolled out in one step

#### Scenario: Upward scroll through the caption is normal

- **WHEN** an image block occupies content lines 4 through 10, the viewport offset is 11 (just below the block), and the user scrolls up one line
- **THEN** the viewport offset becomes 10 (the last caption line) rather than snapping to the photo; further upward scrolls advance through the caption before revealing the full image

#### Scenario: Upward scroll from a fully visible lead photo skips to the top

- **WHEN** on a full-image-capable terminal the lead image block occupies content lines 4 through 10, the viewport offset is 2 so the full block is on screen with the header above, and the user scrolls up one line
- **THEN** the viewport offset snaps to 0 (the article top) in a single step

#### Scenario: Native photo never paints over the border

- **WHEN** on a native-image terminal an image block's rows are only partially within the viewport window (its top above the fold or its bottom below it)
- **THEN** the image's placement is deleted and its transmit suppressed, so no image rows draw over the article border, and the visible rows render the image's halfblock preview rather than a blank strip

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below an image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally

### Requirement: Async image load lifecycle

The system SHALL fetch article lead images asynchronously so that opening an
article never blocks the UI on a network request. Opening an article with an
image URL SHALL first attempt to load the block from the cache hierarchy
(in-memory decoded image, then the stored database block when its width matches
the content width); only on a miss SHALL it issue a network fetch. On a
successful block load, the article re-renders with the block inserted. On a
network success, the image is decoded, the rendered block and the full photo
bytes are persisted to the database, and the decoded image is cached in memory.
On a full-image-capable terminal the fetched photo bytes SHALL additionally be
cached in memory and a native photo render SHALL replace the block. On failure
(network error, timeout, or non-image content type), the system SHALL produce a
failure message and render the article without an image block. The decoded image
SHALL be cached in memory by URL for the session so that resizing the viewport
does not re-fetch or re-decode from the database. When the terminal is resized
while an article with a loaded image is open, the system SHALL re-render the
image block to the new content width using the cached decoded image (or the
cached photo on a native terminal) without re-fetching. When an image loads
after the user has scrolled past the insertion point, the system SHALL preserve
the user's reading position by adjusting the viewport offset by the number of
inserted image-block lines.

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

#### Scenario: Restarted article uses stored block

- **WHEN** the application is restarted and the user opens an article whose block was previously rendered and stored in the database at the current content width
- **THEN** the image block renders from the stored block without making a network request to the image host

#### Scenario: Restarted article uses stored photo

- **WHEN** the application is restarted and the user opens an article whose photo was previously stored at a different content width
- **THEN** the image is re-rendered at the current width from the stored photo bytes without making a network request to the image host

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
the article body (the article's own URL is not listed) in document order,
deduplicated by URL. Each row SHALL show the link text followed by the URL
rendered in the dim style and truncated to the popup width. Navigation SHALL
use `j`/`k` and the arrow keys and SHALL wrap around at both ends. Enter SHALL
open the selected URL in the system browser using the same non-shell
single-argument opener used for other URL opens. While the links popup is open,
`o` SHALL still open the article's own URL in the system browser without closing
the popup. Esc SHALL close the popup
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
  the article border, rails, bullets, scrollbar track, day headers, image
  attribution, and muted popup text (such as plain tag counts and link URLs).
- **blue role** — ANSI color 12 (bright blue) in dark mode and ANSI color 4
  (blue) in light mode: the status bar, the inline text in the article border
  (date, title, help hint, position indicator), the scrollbar thumb, and popup
  titles and popup borders.
- **text role** — ANSI color 7 (white) in dark mode and ANSI color 0 (black)
  in light mode: the article body text and the list view's article rows (read
  article titles, timestamps, and source identifiers).
- **bright role** — ANSI color 15 (bright white) in dark mode and ANSI color 0
  (black) with bold weight in light mode: unread article titles and bold tag
  counts.

The selected-row background highlight SHALL remain a fixed grey
(`#707070`) in both modes.

#### Scenario: Border and muted text share the grey role

- **WHEN** the reader or list view is rendered
- **THEN** the article border, rails, bullets, scrollbar track, day headers,
  image attribution, and popup URLs render in `#707070`

#### Scenario: Status, inline border text, scrollbar thumb, and popup chrome use the blue role

- **WHEN** the status bar, the article border's inline text, the scrollbar
  thumb, or a popup's title or border is rendered in dark mode
- **THEN** it renders in ANSI color 12 (bright blue), and in ANSI color 4
  (blue) in light mode

#### Scenario: Unread titles use the bright role

- **WHEN** an unread article title is rendered in the list view in dark mode
- **THEN** the title renders in ANSI color 15 (bright white) with bold weight,
  and in light mode it renders in ANSI color 0 (black) with bold weight

#### Scenario: Article text uses the text role

- **WHEN** article body text or a list article row (read title, timestamp, or
  source identifier) is rendered in dark mode
- **THEN** it renders in ANSI color 7 (white), and in ANSI color 0 (black) in
  light mode

#### Scenario: Selection highlight keeps its fixed background

- **WHEN** a row is selected in the list view in either mode
- **THEN** the selection highlight renders with the fixed `#707070` background
### Requirement: Inline image caption width

When an inline image renders in the article body with an attribution, the
system SHALL render the attribution at the same standard caption width as the
lead image attribution — a fixed fraction of the content width that does not
depend on the inline photo's own width — wrapped onto additional centered lines
when the caption is longer and centered beneath the image. Inline image blocks
SHALL reserve the wrapped line count of the caption they compose with in their
height fit, so an inline block (image plus caption) stays within the viewport
height budget even for a long caption shared by a gallery of narrow images.

#### Scenario: Long shared gallery caption wraps compactly

- **WHEN** an article renders two inline photos that share a long identical caption and each photo is narrower than the content width
- **THEN** each attribution wraps to the standard caption width beneath its own photo, and neither image-plus-caption block exceeds the viewport height budget

#### Scenario: Inline caption centered at the standard width

- **WHEN** an inline image with a short caption renders
- **THEN** the caption is centered at the standard caption width beneath the image, not stretched to the photo's width

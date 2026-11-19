## MODIFIED Requirements

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view. The
left-most elements SHALL be the time and long-form date the feeds were last
refreshed, formatted as `HH:MM Weekday Day-ordinal Month Year last refresh`
(for example `14:30 Saturday 29th August 2026 last refresh`), or
`HH:MM:SS Weekday Day-ordinal Month Year last refresh` when the `show_seconds`
display option is enabled, with the
`last refresh` label rendered in the dim role (#707070), or the text
`never refreshed` when no refresh has occurred. The right-most elements SHALL
be a literal `?: help` hint, followed by the bullet and the on-disk size of the
article database as `DB <n>MB`, the scroll position percentage, and the article
count as `<n>/<total>` — where the database size is rendered as a whole number
of megabytes (floored to an integer, with no space between the number and `MB`),
`<n>` is the position of the selected article among the shown articles, and
`<total>` is the number of articles currently shown after any active tag filter
(equal to the whole-database count only when no filter is active) — joined by the
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
  `HH:MM Weekday Day-ordinal Month Year last refresh` and, on the right, the
  database size as `DB <n>MB`, the percentage, and the `<n>/<total>` position
  joined by ` • `

#### Scenario: Status bar shows seconds when enabled

- **WHEN** the `show_seconds` display option is enabled and the list view is displayed
- **THEN** the status bar shows the last-refresh time on the left as `HH:MM:SS Weekday Day-ordinal Month Year last refresh`

#### Scenario: Status bar uses dim bullets and dim refresh label

- **WHEN** the list view is displayed
- **THEN** the bullets between the database size, percentage, and position
  render in the dim role (#707070), the `last refresh` label renders in the dim
  role, and the elements between the bullets render in the chrome role

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the refresh
  time, the database size, the percentage, and the `<n>/<total>` position

#### Scenario: Status bar total counts shown articles under a filter

- **WHEN** a tag filter is active and shows 23 of 340 articles in the database
- **THEN** the status bar position reads `<n>/23`, where `<n>` is the position
  of the selected article among the 23 shown articles, not `<n>/340`

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

### Requirement: Article top border title truncation

The system SHALL render the article's date and title inline with the top border
with the date always fully visible and the title left-aligned after the date.
The date prefix SHALL be `YYYY-MM-DD HH:MM` showing the publication date and
time in the user's local timezone by default, or `YYYY-MM-DD HH:MM:SS` when the
`show_seconds` display option is enabled, followed by the bullet
(`•`, ASCII
`.`) and the title. The top border SHALL also display a right-aligned position
counter `<n>/<total>`, where `<n>` is the 1-based position of the open article
among the shown articles and `<total>` is the number of shown articles after any
active tag filter, separated from the title by dash fill and from the right
corner glyph by a single dash (for example
`┌─ 2026-01-02 15:04 • Hello world ───── 3/23 ─╖`); the counter SHALL render in
the same inline-border role as the date and title. When the open article is not
present in the shown list, the counter SHALL be omitted and the top border SHALL
fall back to dash fill to the corner. When the title does not fit the available
width, the system SHALL truncate the title and end it with an ellipsis glyph —
`…` in Unicode mode and `...` in ASCII fallback mode — and SHALL place a single
horizontal dash adjacent to the right corner glyph so the right end mirrors the
left (`┌─` on the left, `─╖` on the right). When the title fits with room to
spare, the system SHALL fill the leftover space with horizontal dashes as before
and SHALL NOT append an ellipsis.

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

#### Scenario: Top border shows the position counter

- **WHEN** the open article is the 3rd of 23 shown articles
- **THEN** the top border's right side reads `3/23`, separated from the title by dash fill and from the right corner by a single dash

#### Scenario: Position counter omitted when the article is not in the list

- **WHEN** the open article is not present in the shown list
- **THEN** the top border omits the position counter and fills the space to the right corner with dashes

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
bottom border SHALL display a left-aligned help hint and, on the right, a
position indicator led by a literal
`?: help` hint in the form
`?: help • <percent>% • <bottomLine>/<totalLines>`, where `<percent>` is the
scroll percentage, `<bottomLine>` is the line number of the last visible
viewport line clamped to the total, and `<totalLines>` is the article's total
line count; horizontal dashes fill the space between the hint and the indicator
(for example at the very bottom:
`o: open article in browser ───── ?: help • 100% • 120/120`). The left-aligned
hint SHALL normally be `o: open article in browser`; when the article is
scrollable and the viewport is at the top edge, the hint SHALL instead read
`K: next`, and when the viewport is at the bottom edge, `J: previous`, each
shown for two seconds before reverting to `o: open article in browser` (moving
the viewport off the edge also reverts the hint immediately). The bullets used
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

#### Scenario: Bottom border shows the next hint at the top

- **WHEN** a scrollable article is at the top edge (offset zero)
- **THEN** the bottom border's left hint reads `K: next` instead of `o: open article in browser`

#### Scenario: Bottom border shows the previous hint at the bottom

- **WHEN** a scrollable article is at the bottom edge (fully scrolled)
- **THEN** the bottom border's left hint reads `J: previous` instead of `o: open article in browser`

#### Scenario: Navigation hint reverts to the open hint after two seconds

- **WHEN** the `K: next` or `J: previous` hint has been shown for two seconds
- **THEN** the bottom border's left hint reverts to `o: open article in browser`

#### Scenario: Short article keeps the open hint

- **WHEN** the article fits entirely within the viewport (not scrollable)
- **THEN** the bottom border's left hint stays `o: open article in browser` and no `J:`/`K:` hint appears

#### Scenario: ASCII fallback border

- **WHEN** ASCII fallback mode is enabled or the terminal lacks box-drawing support
- **THEN** single-line border glyphs are replaced with ASCII equivalents (`-`, `+`), the left and right vertical borders are both `:` rendered in the grey style, and the scrollbar thumb is `|` rendered in the bright/white style (the unfilled track stays `:` in grey)

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
exceed 70 columns wide. The popup SHALL dismiss — closing and returning to the
underlying view — when the user presses Esc, Enter, Space, `b`, or `h`.

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

#### Scenario: B and H dismiss the help popup

- **WHEN** the help popup is open over the list or article view and the user presses `b` or `h`
- **THEN** the popup closes and the underlying view remains

### Requirement: Quit and utility keys

The system SHALL quit on `q` or Ctrl-C from the list and article views, and SHALL
also quit on `q` or Ctrl-C while a popup is open (from a popup, `q` quits the
application rather than merely closing the popup). The system SHALL provide `m`
to toggle read/unread status, `a` to mark all articles in the current view as
read, `o` to open the article URL in the system browser, `c` to copy the article
URL to the clipboard, `C` to copy the article text to the clipboard, and `?` to
display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in the list or article view
- **THEN** the application exits

#### Scenario: Q quits from an open popup

- **WHEN** the user presses `q` while a popup is open over the list or article view
- **THEN** the application exits, and the popup does not merely close

#### Scenario: Ctrl-C quits from an open popup

- **WHEN** the user presses Ctrl-C while a popup is open over the list or article view
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

### Requirement: Article links popup

The system SHALL provide a links popup in the article view, opened by pressing
`l` or the right arrow key. The popup SHALL list every hyperlink rendered in
the article body (the article's own URL is not listed) in document order,
deduplicated by URL. Each row SHALL show the link text followed by the URL
rendered in the dim style and truncated to the popup width. A link that wraps
an inline image with no other label text (a photo linked to itself, as Substack
figures are) SHALL show the wrapped photo's caption — the same attribution the
article view composes beneath the photo — as its link text, elided to at most
40 columns with the middle replaced by the glyph-mode ellipsis. Navigation SHALL
use `j`/`k` and the arrow keys and SHALL wrap around at both ends. Enter SHALL
open the selected URL in the system browser using the same non-shell
single-argument opener used for other URL opens. While the links popup is open,
`o` SHALL still open the article's own URL in the system browser without closing
the popup. Esc, the left arrow, or `h` SHALL close the popup
without navigating, while `q` SHALL quit the application. When the article
contains no links, the popup SHALL open
with an empty list and Enter SHALL do nothing. While the popup is closed,
Enter SHALL continue to close the article view and return to the list.

#### Scenario: Open links popup from article

- **WHEN** the user presses `l` or the right arrow key in the article view
- **THEN** a popup opens listing the article's links in document order

#### Scenario: Row shows link text and dimmed URL

- **WHEN** the article contains `<a href="https://example.com/post">world</a>` and the links popup is open
- **THEN** the popup shows a row with the text `world` followed by the URL `https://example.com/post` in the dim style, truncated to the popup width

#### Scenario: A self-linked photo borrows its caption as link text

- **WHEN** an inline image linked to a photo URL has no label text (a figure like `<a href=".../a-detail"><img src=".../a"></a>` under a long `figcaption`) and the links popup is open
- **THEN** the row's label is the photo's caption shortened to at most 40 columns — its start and end joined by `…` (`...` in ASCII glyph mode) — followed by the wrapped URL

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

#### Scenario: Esc, left, or H closes the popup without navigating

- **WHEN** the user presses Esc, the left arrow, or `h` while the links popup is open
- **THEN** the popup closes, the article view remains, and no URL is opened

#### Scenario: Q quits from the links popup

- **WHEN** the user presses `q` while the links popup is open
- **THEN** the application exits

#### Scenario: Empty article shows an empty popup

- **WHEN** the article contains no links and the user opens the links popup
- **THEN** the popup opens with an empty list and Enter does nothing

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
the list (via `b`, `esc`, or `h`) leaves the same article selected. The active
tag filter SHALL NOT be persisted: on startup the list SHALL reopen in the
default (unfiltered) view regardless of the filter that was active when the
user quit, and selecting a source or tag from a popup SHALL replace the current
filter rather than stack on top of it.

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

#### Scenario: Tag filter is not restored on start

- **WHEN** the user quits with a source or tag filter active on the list and restarts
- **THEN** the list reopens in the default unfiltered view, not under the filter that was active at quit

#### Scenario: Selecting a tag replaces the current filter

- **WHEN** a filter is active and the user selects a different source or tag from a popup
- **THEN** the new filter replaces the old one (the list never stacks filters)

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

## ADDED Requirements

### Requirement: Article next/previous navigation

The system SHALL let the reader step between articles directly from the article
view. `J` SHALL open the previous article — the older article immediately below
the current one in the newest-first list — and `K` SHALL open the next article —
the newer article immediately above it — mirroring the list view's `j`/`k`
movement direction. Both SHALL wrap around the ends: `J` on the oldest article
SHALL open the newest, and `K` on the newest SHALL open the oldest. Navigation
SHALL follow the currently displayed list, so when a tag or source filter is
active the step stays within that filter's articles, and when no filter is
active it walks the whole feed. When the open article is not present in the
displayed list (for example it was pruned after being opened), `J` and `K` SHALL
be no-ops. Stepping to another article SHALL behave like opening it from the
list: the new article's content SHALL be composed, its read-on-open marking
SHALL apply, and its images SHALL be scheduled for load; returning to the list
afterward SHALL select the newly opened article.

#### Scenario: J opens the older article

- **WHEN** the reader is viewing an article that is not the oldest and the user presses `J`
- **THEN** the article immediately below the current one (older) opens

#### Scenario: K opens the newer article

- **WHEN** the reader is viewing an article that is not the newest and the user presses `K`
- **THEN** the article immediately above the current one (newer) opens

#### Scenario: J wraps from the oldest to the newest

- **WHEN** the reader is viewing the oldest article and the user presses `J`
- **THEN** the newest article opens

#### Scenario: K wraps from the newest to the oldest

- **WHEN** the reader is viewing the newest article and the user presses `K`
- **THEN** the oldest article opens

#### Scenario: Navigation stays within a tag filter

- **WHEN** the list is filtered to a source or tag and the reader is viewing one of its articles
- **THEN** `J` and `K` step only among that filter's articles

#### Scenario: Navigation is a no-op when the article is not in the list

- **WHEN** the open article is not present in the displayed list and the user presses `J` or `K`
- **THEN** nothing happens

#### Scenario: Returning to the list selects the newly opened article

- **WHEN** the reader steps to another article with `J` or `K` and then returns to the list
- **THEN** the list selects the newly opened article

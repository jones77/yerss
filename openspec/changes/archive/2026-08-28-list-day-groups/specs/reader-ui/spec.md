## MODIFIED Requirements

### Requirement: Article list rendering

The system SHALL display articles grouped under a collapsible header for each
day, with the most recent day first and articles within a day sorted reverse
chronologically. Each day header SHALL display the long date in the form
`Weekday Day-ordinal Month, Year` (for example `Saturday 20th August, 2025`),
computed in the user's local timezone, where the day is suffixed with an English
ordinal (`1st`, `2nd`, `3rd`, `4th`, …). The system SHALL NOT render a top title
banner; the most recent day header is the first content at the top of the list.
Each article row SHALL show the article title on the left and its publication
time as `HH:MM:SS` on the right, in the user's local timezone. Unread articles
SHALL be rendered in bold text and read articles in plain (dim) text. A
collapsed day header SHALL hide its article rows; an expanded day header SHALL
show them. Selection SHALL wrap around when moving past the first or last
visible row (headers plus expanded articles).

#### Scenario: Unread article appears bold

- **WHEN** the article list is displayed and an article has read status false
- **THEN** that article's title is rendered in bold text

#### Scenario: Read article appears plain

- **WHEN** the article list is displayed and an article has read status true
- **THEN** that article's title is rendered in plain (dim) text

#### Scenario: Articles grouped under a day header

- **WHEN** the article list is displayed and articles exist on more than one day
- **THEN** each day's articles are listed beneath a header for that day, with the most recent day first and articles reverse chronological within the day

#### Scenario: Day header shows the long local date

- **WHEN** a day header is rendered for a date with day-of-month 20 in August 2025
- **THEN** the header displays the long date with weekday, an ordinal day number, month, and year (for example `Wednesday 20th August, 2025`) in the user's local timezone

#### Scenario: Article row shows local HH:MM:SS

- **WHEN** an article row is rendered
- **THEN** the row shows the article title on the left and the publication time as `HH:MM:SS` on the right in the user's local timezone

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
in the database, followed by the date the feeds were last refreshed. When a tag
filter is active, the status bar SHALL also display the active filter name.

#### Scenario: Status bar shows filtered/total count and refresh date

- **WHEN** the list view is displayed
- **THEN** the status bar shows the article count as `<n>/<total>` (shown over total) and the last-refreshed date

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the `<n>/<total>` count and the date

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation
(PgUp/PgDn, Ctrl-F/Ctrl-B) SHALL scroll by a full viewport page. Half-page
navigation (Ctrl-D/Ctrl-U, Space) SHALL scroll by half a viewport. `g` or
Ctrl-Up SHALL move to the top; `G` or Ctrl-Down SHALL move to the bottom. In the
list view, up/down moves selection across visible rows (day headers and expanded
articles); in the article view, up/down scrolls the article by one line. `Enter`
or `l` opens the selected article from the list; `Esc`, `Enter`, or `h` returns
from the article view to the list, reselecting the previously viewed article. In
the list view, `Esc` and `h` clear an active tag filter or are no-ops when no
filter is active.

When the list cursor is on a day header, fold keys SHALL act on that day:
`h` SHALL collapse it, `l` SHALL expand it, and `Enter`, Space, or Tab SHALL
toggle its expansion; the article-open behavior of `l`/Enter SHALL NOT apply on
a header row. When the cursor is on an article row, `Tab` SHALL toggle the
expansion of that article's day group, and the existing bindings (`h` clears the
filter or is a no-op, `l`/Enter open the article, Space half-pages) SHALL apply
unchanged.

#### Scenario: Open article from list

- **WHEN** the user presses Enter or `l` on a selected article row in the list view
- **THEN** the article reader view opens displaying that article

#### Scenario: Fold keys on a day header

- **WHEN** the list cursor is on a day header and the user presses `h`, `l`, Enter, Space, or Tab
- **THEN** `h` collapses that day, `l` expands it, and Enter/Space/Tab toggle its expansion; no article is opened

#### Scenario: Tab toggles the current day group from an article row

- **WHEN** the list cursor is on an article row and the user presses Tab
- **THEN** the day group containing that article toggles its expansion

#### Scenario: Return to list from article

- **WHEN** the user presses Esc, Enter, or `h` in the article view
- **THEN** the list view is displayed with the previously viewed article selected

#### Scenario: Clear tag filter from list

- **WHEN** a tag filter is active on the list view and the user presses Esc, `h`, or left arrow on an article row
- **THEN** the tag filter is removed and all articles are shown

#### Scenario: Esc on unfiltered list is a no-op

- **WHEN** no tag filter is active on the list view and the user presses Esc or `h` on an article row
- **THEN** nothing happens

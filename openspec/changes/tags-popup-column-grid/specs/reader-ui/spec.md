## MODIFIED Requirements

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
`reallylongnewspaperdomainname.net` → `reallyl…ainname`). When the article has
no link, the system SHALL derive the identifier from the feed URL host instead;
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

## MODIFIED Requirements

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
host by stripping any leading `www.`, removing the top-level domain (the part
after the final dot), and truncating the remainder to at most 12 characters
(for example `newrepublic.com` → `newrepublic`,
`reallylongnewspaperdomainname.net` → `reallylongne`). When the article has no
link, the system SHALL derive the identifier from the feed URL host instead.
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

## MODIFIED Requirements

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
headers SHALL render entirely in the blue role: the rail corner, the dash
connector, and the date label. Each
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

#### Scenario: Day header renders in the blue role

- **WHEN** a day header is rendered in dark mode
- **THEN** the rail corner, the dash connector, and the date label render in ANSI color 12 (bright blue), and in ANSI color 4 (blue) in light mode

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

### Requirement: Standard terminal color roles

The system SHALL derive all UI foreground colors from the standard terminal
(ANSI) palette instead of bespoke RGB values. The following color roles SHALL
apply across the list, reader, and popup views:

- **grey/dim role** — the colour `#707070` in both dark and light modes:
  the article border, rails, bullets, scrollbar track, image
  attribution, and muted popup text (such as plain tag counts and link URLs).
- **blue role** — ANSI color 12 (bright blue) in dark mode and ANSI color 4
  (blue) in light mode: the status bar, day headers, the inline text in the
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

#### Scenario: Status, day headers, inline border text, scrollbar thumb, and popup chrome use the blue role

- **WHEN** the status bar, a day header, the article border's inline text, the scrollbar
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

#### Scenario: Selection highlight is theme-aware

- **WHEN** a row is selected in the list view in dark mode
- **THEN** the selection highlight renders with the `#242424` background, and
  with the `#d3d3d3` background in light mode
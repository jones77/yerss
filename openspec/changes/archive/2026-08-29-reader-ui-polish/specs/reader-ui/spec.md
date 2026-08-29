## MODIFIED Requirements

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

## ADDED Requirements

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

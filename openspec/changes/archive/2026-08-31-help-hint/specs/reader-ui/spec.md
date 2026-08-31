## MODIFIED Requirements

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view. The
left-most elements SHALL be the time and long-form date the feeds were last
refreshed, formatted as `HH:MM Weekday Day-ordinal Month, Year last refresh`
(for example `14:30 Saturday 29th August, 2026 last refresh`), with the
`last refresh` label rendered in the dim role (#707070), or the text
`never refreshed` when no refresh has occurred. The right-most elements SHALL
be a literal `?: help` hint followed by the bullet and
the on-disk size of the article database, formatted as a human-readable
value with binary unit suffixes (B, KB, MB, GB), the scroll position percentage,
and the article count as `<n>/<total>`, where `<n>` is the position of the
selected article among the shown articles and `<total>` is the total number of
articles in the database, joined by the middle-dot bullet (`·`, ASCII `.`) (for
example `?: help · 20.6 MB · 100% · 2/2`). The bullets and the `last refresh`
label SHALL render in the dim role (#707070); the remaining status bar text
SHALL render in the chrome role.
The reported size SHALL include the WAL sidecar file when one is present, so it
reflects the actual on-disk footprint. When a tag filter is active, the status
bar SHALL also display the active filter name.

#### Scenario: Status bar leads the right side with the help hint

- **WHEN** the list view is displayed
- **THEN** the status bar's right side reads `?: help · <size> · <percent>% ·
  <n>/<total>`, with `?: help` as the first right-aligned element

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
ASCII `.`). The content area SHALL have two spaces of horizontal padding on
each side and one space of vertical padding below the
content and none above it, so the first content line (the article URL) sits
directly beneath the top border. When the entire article fits within the viewport
and no scrolling is
possible, the right border SHALL be fully filled with the bright single-line
glyph and the bottom border SHALL display `100%` with `<bottomLine>` equal to
`<totalLines>`.

#### Scenario: Bottom border leads the right side with the help hint

- **WHEN** the article view is rendered
- **THEN** the bottom border's right side reads `?: help · <percent>% ·
  <bottomLine>/<totalLines>`, with `?: help` as the first right-aligned element
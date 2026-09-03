# Delta: reader-ui (restore-permanent-article-margin)

## MODIFIED Requirements

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
horizontal padding on each side and one space of vertical padding above and
below the content, so a permanent one-row margin separates the content
viewport from both the top and the bottom borders. When the entire article fits within the viewport
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
## MODIFIED Requirements

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
the viewport SHALL snap to the top boundary (its first line at the viewport top)
as soon as a downward single-line scroll brings its top into the window from
below, mirroring the upward scroll's entry snap, so a short photo entered from
above is aligned flush rather than scrolled through line-by-line; the next
downward scroll then skips the whole block onto the line after its caption. The
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

#### Scenario: Short inline image entered from above snaps flush to the viewport top

- **WHEN** a downward single-line scroll brings a short inline image's top into the window from below (its block is shorter than the viewport, so it is never partially clipped), the viewport height is 15, and the image's top line is 10
- **THEN** the viewport offset snaps to 10 (the top boundary) so the image aligns flush to the viewport top in a single step, rather than scrolling into view line-by-line

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
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
the viewport top (the top boundary); and a scroll landing in its photo range
then skips it onto its caption, leaving the caption at the viewport top. A
downward scroll that would move the offset into the photo's range from above
SHALL snap to the photo's first line (the top boundary); one starting at or
inside it skips it onto its caption. The fully-visible skip SHALL NOT apply to
inline images, so an inline image is never skipped past without first being
shown. Upward single-line scrolls mirror the transitions: an image entering from
above snaps its top to the viewport top (the top boundary), the next upward
scroll snaps its bottom boundary so the caption's last line is at the viewport
bottom, a scroll landing in its photo range reveals it (snaps to the photo's
first line), and an upward scroll that cuts the block's last line below the fold
SHALL snap it fully below the fold on the next move so it scrolls off the bottom
edge cleanly. When an inline image's block fills the viewport exactly (its image
plus wrapped caption), its bottom and top boundaries coincide, so there is no
distinct bottom stage: the upward scroll from the top-aligned position SHALL
scroll the image fully out of view rather than looping on the same offset. When
an exit or scroll-off position would land inside a preceding image block's photo
range (image blocks spaced closer than the viewport height), the system SHALL
snap to that image's first line instead, revealing it. Scrolling within the
attribution lines SHALL behave as normal line scrolling in both directions. On a
full-image-capable terminal, when the lead photo is fully visible and the user
scrolls upward with the offset past the article top, the system SHALL snap the
offset to the article top in a single step.

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

#### Scenario: Inline image snaps to the viewport top on the next downward scroll

- **WHEN** an inline image is fully visible with its last line at the viewport bottom (at its bottom boundary) and the user scrolls down one line
- **THEN** the viewport offset snaps to the image's first line so the image moves to the top of the viewport

#### Scenario: Inline image skips to its caption when a scroll lands in its photo

- **WHEN** the viewport offset lands inside an inline image's photo range after a downward scroll that started at or inside the image
- **THEN** the viewport offset snaps to the attribution's first line, skipping the image out of view onto its caption

#### Scenario: Upward scroll reveals full inline image

- **WHEN** the viewport offset lands inside an inline image's photo range and the user scrolls up one line
- **THEN** the viewport offset snaps to the photo's first line and the full image and its attribution are visible

#### Scenario: Upward scroll moves a top-aligned inline image to the bottom

- **WHEN** an inline image is fully visible with its top at the viewport top (at its top boundary) and the user scrolls up one line
- **THEN** the viewport offset snaps so the image's bottom boundary holds: the image and its wrapped caption are visible at the bottom of the screen with the caption's last line at the viewport bottom

#### Scenario: Upward scroll snaps a partially-scrolled-off inline image below the fold

- **WHEN** an upward scroll cuts an inline image's last line below the fold, leaving it partially visible at the bottom edge
- **THEN** the next upward scroll snaps the offset so the image's top sits at the fold, scrolling the image fully out of view; a frame where the image pokes only partially into the window renders its halfblock preview rather than a blank strip

#### Scenario: Exit or scroll-off reveals a preceding image when the target lands in its photo

- **WHEN** image blocks are spaced closer than the viewport height and a scroll-off position for one image (its top at the fold) would land inside a preceding image's photo range
- **THEN** the viewport offset snaps to the preceding image's first line so it is revealed rather than left partially clipped

#### Scenario: Page-down lands naturally even mid-photo

- **WHEN** the user pages down and the resulting viewport offset lands inside an image block's photo range
- **THEN** no snapping occurs; the offset stays where the page-down landed, even with a partially clipped photo

#### Scenario: Image load does not leave the offset inside a photo

- **WHEN** an inline image finishes loading and the article re-composes while the viewport offset falls inside the image's photo range or with its top within the window
- **THEN** the viewport offset snaps to the image's first line so the image is shown, never left partially clipped

#### Scenario: Image load at the article bottom keeps the end visible

- **WHEN** an image finishes loading and the article re-composes while the user is on the article's last screen (they moved to the bottom with `G` or scrolled to the end), and the bottom offset falls inside a photo's range whose caption is cut off below the fold
- **THEN** the offset stays at the article bottom; it does not snap up to the photo's first line, so the article's end remains visible

#### Scenario: Downward scroll within the caption is normal

- **WHEN** the viewport offset is on an attribution line and the user scrolls down one line
- **THEN** no snapping occurs and the next attribution or body line scrolls into view

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
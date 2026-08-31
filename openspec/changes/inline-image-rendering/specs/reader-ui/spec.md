## MODIFIED Requirements

### Requirement: Atomic image scroll behavior

When an image block (a lead image or an inline image) is rendered, the system
SHALL ensure the photo's lines are either fully visible within the viewport or
fully scrolled out of view; it SHALL never display a partially clipped photo.
Each image block SHALL apply these rules independently, using its own caption.
When a downward scroll would move the viewport offset into the photo's line
range, the system SHALL snap the offset to the photo's attribution line when one
is rendered — landing on the caption so it can be read — and to the line
immediately after the block when no attribution is rendered, skipping the photo
in a single keystroke. When a downward scroll leaves the photo fully visible —
the photo's first line at or above the viewport top and its last line within the
viewport bottom — the system SHALL snap the offset to the first non-photo line
(the attribution line, or the line after the block when there is none).
Scrolling within the attribution lines SHALL behave as normal line scrolling in
both directions: an upward scroll landing within the attribution SHALL advance
the caption line by line rather than snapping to the photo. When an upward
scroll would move the offset into the photo's line range, the system SHALL snap
the offset to the photo's first line, revealing the full photo and its
attribution. On a full-image-capable terminal, when a photo is fully visible and
the user scrolls upward with the offset past the article top, the system SHALL
snap the offset to the article top in a single step. When a downward scroll
brings an inline image into view from below, the system SHALL snap the image's
top to the viewport top so the full image is visible, and the next downward
scroll SHALL scroll it out onto its caption; symmetrically, an upward scroll
that brings an inline image into view from above SHALL snap the image so its
bottom aligns with the viewport bottom. This snapping SHALL apply to all scroll
operations: line, page, half-page, and goto-bottom.

#### Scenario: Downward scroll skips the photo onto its caption

- **WHEN** an image block occupies content lines 4 through 10 (photo lines 4 through 7 plus attribution lines 8 through 10), the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 8 (the attribution's first line): the photo is fully scrolled out and the caption is visible at the top of the viewport

#### Scenario: Downward scroll within the caption is normal

- **WHEN** the viewport offset is on an attribution line and the user scrolls down one line
- **THEN** no snapping occurs and the next attribution or body line scrolls into view

#### Scenario: Downward scroll without attribution skips the block

- **WHEN** an image block has no attribution, the photo occupies lines 4 through 9, the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 10 (the line after the block) and the photo is fully scrolled out

#### Scenario: Downward scroll skips a fully visible photo

- **WHEN** an image block occupies content lines 4 through 10, the viewport offset is 2 so the full block is on screen, and the user scrolls down one line
- **THEN** the viewport offset snaps to 8 (the first non-photo line) and the photo is fully scrolled out in one step

#### Scenario: Upward scroll through the caption is normal

- **WHEN** an image block occupies content lines 4 through 10, the viewport offset is 11 (just below the block), and the user scrolls up one line
- **THEN** the viewport offset becomes 10 (the last caption line) rather than snapping to the photo; further upward scrolls advance through the caption before revealing the full image

#### Scenario: Upward scroll from a fully visible photo skips to the top

- **WHEN** on a full-image-capable terminal an image block occupies content lines 4 through 10, the viewport offset is 2 so the full block is on screen with the header above, and the user scrolls up one line
- **THEN** the viewport offset snaps to 0 (the article top) in a single step

#### Scenario: Upward scroll reveals full image with attribution

- **WHEN** an image block occupies content lines 4 through 10, the viewport offset is on a photo line (between 5 and 7), and the user scrolls up one line
- **THEN** the viewport offset snaps to 4 (the photo's first line) and the full image and its attribution are visible

#### Scenario: Page-down skips the photo onto the caption

- **WHEN** the photo is partially within the viewport after a page-down operation
- **THEN** the offset snaps to the attribution's first line when an attribution is rendered, otherwise past the block

#### Scenario: Inline image snaps to the viewport top when scrolled into view

- **WHEN** an inline image is below the viewport and a downward scroll brings its top into the viewport bottom
- **THEN** the offset snaps so the image's top is at the viewport top, and the next downward scroll scrolls it out onto its caption

#### Scenario: Inline image snaps to the viewport bottom when scrolled up into view

- **WHEN** an inline image is above the viewport and an upward scroll brings its bottom into the viewport top
- **THEN** the offset snaps so the image's bottom is at the viewport bottom, and the next upward scroll reveals it

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below an image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally

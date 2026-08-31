## Purpose

Renders article lead photos inline through the terminal's native image protocol
on full-image-capable terminals, showing the stored halfblock placeholder first.

## ADDED Requirements

### Requirement: Native image capability detection

The system SHALL detect, from the terminal environment, whether the terminal can
render real inline images. Full-image-capable terminals SHALL include iTerm2
(OSC 1337 inline images) and kitty-family terminals (kitty graphics protocol).
Detection SHALL be environment-based, mirroring the existing ASCII and
light-background detection, and SHALL NOT require user configuration. Terminals
that cannot be identified as full-image-capable SHALL be treated as
halfblock-only.

#### Scenario: iTerm2 detected as full-image-capable

- **WHEN** the terminal environment identifies iTerm2
- **THEN** the system renders lead photos via the OSC 1337 inline image protocol

#### Scenario: Kitty-family terminal detected as full-image-capable

- **WHEN** the terminal environment identifies a kitty graphics-protocol terminal
- **THEN** the system renders lead photos via the kitty graphics protocol

#### Scenario: Unrecognized terminal treated as halfblock-only

- **WHEN** the terminal environment does not identify a full-image-capable terminal
- **THEN** lead images render as halfblock blocks only

### Requirement: Native photo render with block placeholder

On a full-image-capable terminal, opening an article with a lead image SHALL
display a halfblock block immediately — the stored block when present, otherwise
one rendered from a fetch — and SHALL fetch the real photo asynchronously. When
the photo arrives it SHALL be rendered natively (inline image protocol) in place
of the block, sized to fit the content width with aspect ratio preserved and
capped to the viewport, centered within the content width exactly like the
halfblock block (its reserved rows span only the photo's own cell box), and
SHALL be part of the scrolling content with the photo attribution centered
beneath. When the photo fetch fails, the block SHALL remain as the final render.
Resizing the terminal SHALL re-scale the native render from the cached photo
without re-fetching. The fetched photo SHALL be held only in memory for the
session and SHALL NOT be persisted.

#### Scenario: Block shown before photo fetch

- **WHEN** a full-image-capable terminal opens an article with a lead image URL and a stored block
- **THEN** the stored block renders immediately and the photo fetch runs in the background

#### Scenario: Native render replaces the block

- **WHEN** the background photo fetch succeeds on a full-image-capable terminal
- **THEN** the halfblock block is replaced by a native inline render of the photo sized to the content width, with the attribution still beneath

#### Scenario: Height-capped photo is centered

- **WHEN** a native render is narrowed below the content width by the viewport height cap or the attribution wrap
- **THEN** the photo and its reserved rows span only the photo's own cell box and the block is centered within the content width with equal left and right margins, with the attribution centered beneath, matching the halfblock block's placement

#### Scenario: Native fetch failure keeps the block

- **WHEN** the background photo fetch fails on a full-image-capable terminal
- **THEN** the halfblock block remains as the final render and no error is surfaced

#### Scenario: Resize re-scales native render without re-fetch

- **WHEN** the terminal is resized while a native photo render is displayed
- **THEN** the native render is re-scaled to the new content width from the cached photo with no network request

#### Scenario: Photo not persisted

- **WHEN** a full-image-capable terminal fetches a lead photo
- **THEN** the photo bytes are cached only in memory for the session and never written to the database

### Requirement: Native image placement cleanup

The system SHALL delete the terminal's native image placements on every frame
that does not display the open article's native image. On kitty graphics
protocol terminals a placement floats above text and survives the erase
commands of a frame repaint, so frames that leave the article view, open
another article, show the placeholder block, or scroll the image out of the
viewport SHALL emit the protocol's delete action for all visible placements.
Frames that display the native image block SHALL NOT emit the delete (their
re-transmission with the stable placement id replaces the placement), and a
scrolled-out image SHALL be re-transmitted when scrolled back into view.
Terminals whose inline images are cell-bound (OSC 1337) SHALL NOT emit any
delete sequence, since the repaint itself erases those images.

#### Scenario: Backing out of an article deletes the photo

- **WHEN** the reader returns from an article showing a native photo to the list view on a kitty-family terminal
- **THEN** the photo's placement is deleted from the screen and does not persist over the list

#### Scenario: Opening another article does not stack photos

- **WHEN** an article with a native photo is opened after an earlier article's photo was displayed on a kitty-family terminal
- **THEN** the earlier placement has been deleted, so the new photo renders alone without stacking over the old one

#### Scenario: Scrolling the image out of view deletes the placement

- **WHEN** the article is scrolled so the photo's rows leave the visible viewport window on a kitty-family terminal (including landing on the caption, which leaves the caption visible)
- **THEN** the placement is deleted so the photo does not float over the caption or body text, and scrolling the photo back into view re-transmits it

#### Scenario: OSC 1337 images need no delete

- **WHEN** frames are rendered on an OSC 1337 terminal
- **THEN** no kitty delete sequence is emitted, because cell-bound inline images are erased by the frame repaint
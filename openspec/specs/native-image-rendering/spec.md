## Purpose

Renders article lead photos inline through the terminal's native image protocol
on full-image-capable terminals, showing the stored halfblock placeholder first.

## Requirements

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
one rendered from a fetch — and SHALL fetch the real photo asynchronously. The
native render itself — decoding the photo, scaling it to the display box, and
encoding it for the terminal — SHALL also be performed asynchronously and never
on the UI thread, so that opening or resizing an article with a native photo
never freezes the UI on a CPU-bound operation. When the photo arrives it SHALL
be rendered natively (inline image protocol) in place of the block, sized to fit
the content width with aspect ratio preserved and capped to the viewport,
centered within the content width exactly like the halfblock block (its reserved
rows span only the photo's own cell box), and SHALL be part of the scrolling
content with the photo attribution centered beneath. The rendered native output
SHALL be cached in memory keyed by render size — the content width together with
the viewport-height cap — so re-opening an article at the same size SHALL
display the photo immediately from cache without re-downloading or
re-rendering. When the photo fetch fails, the block SHALL remain as the final
render. Resizing the terminal SHALL re-scale the native render from the cached
photo without re-fetching, asynchronously and without blocking the UI. The
native render SHALL decode the source once per render and reuse the decoded
image across the height-fit iterations, and when the cached photo bytes are
unavailable it SHALL re-read the photo from the stored database bytes rather
than failing. The fetched photo SHALL be persisted to the database (in the
unified image table) so a later open re-renders natively from the stored bytes
without a network request; the rendered native output SHALL remain an in-memory
render cache keyed by render size (content width and viewport-height cap).

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

#### Scenario: Content-width resize re-scales the render

- **WHEN** the terminal is resized so the content width changes while a native photo render is displayed
- **THEN** the native render is re-scaled to the new content width from the cached photo with no network request

#### Scenario: Viewport-height resize re-renders the block

- **WHEN** the terminal is resized so only the viewport height changes while the content width stays the same
- **THEN** the native render is re-rendered for the new viewport-height cap from the cached photo, without a network request

#### Scenario: Photo persisted for later re-render

- **WHEN** a full-image-capable terminal fetches a lead photo
- **THEN** the photo bytes are written to the database so a later open re-renders natively without re-fetching

#### Scenario: Opening a high-resolution photo does not block the UI

- **WHEN** a full-image-capable terminal opens an article whose lead photo is high-resolution (a large source image)
- **THEN** the article text renders immediately and the photo's decode, scale, and encode happen in the background rather than on the UI thread

#### Scenario: Re-opened article shows the cached native render instantly

- **WHEN** the user re-opens an article whose photo was already rendered natively at the current render size
- **THEN** the photo appears immediately from the in-memory render cache, with no network request and no re-decode or re-encode

#### Scenario: Unavailable photo re-derives from storage

- **WHEN** an article's photo bytes are not in the in-memory photo cache (for example on a fresh session) and stored bytes exist
- **THEN** the render reads the photo from the stored database bytes instead of failing, and the halfblock block remains the fallback only when no stored bytes exist

#### Scenario: Height-fit reuses one decode

- **WHEN** the height-fit loop renders the same photo at multiple candidate heights
- **THEN** the source is decoded once and the candidates scale and re-encode the shared decoded image, never re-decoding the source bytes per candidate

### Requirement: Native image placement cleanup

The system SHALL delete the terminal's native image placements on every frame
that does not display the open article's native image. On kitty graphics
protocol terminals a placement floats above text and survives the erase
commands of a frame repaint, so frames that leave the article view, open
another article, show the placeholder block, or scroll the image out of the
viewport SHALL emit the protocol's delete action for all visible placements.
When the reader leaves the article view entirely — returning to the list,
opening another article, or an article whose native image never composes — the
system SHALL also delete each native image's cached image data by its stable id
so the terminal's image cache is freed rather than accumulating every image
ever shown for the life of the terminal tab. The delete-by-id sequence SHALL
use the protocol's data-freeing form (the uppercase `d=I` delete action), which
removes the placement and releases the terminal's cached image data, not the
placement-only form (`d=i`) that retains the cached data. On terminals whose
kitty-graphics delete handling is known to be partial or unreliable (Ghostty),
the system SHALL additionally emit a delete-all clear (`d=a`) on the frame that
leaves the article view, so a placement that survives the by-id delete cannot
float over the list. Frames that display the native image block SHALL NOT emit
the delete (their re-transmission with the stable placement id replaces the
placement), and a scrolled-out image SHALL be re-transmitted when scrolled back
into view. Terminals whose inline images are cell-bound (OSC 1337) SHALL NOT
emit any delete sequence, since the repaint itself erases those images.

#### Scenario: Backing out of an article deletes the photo

- **WHEN** the reader returns from an article showing a native photo to the list view on a kitty-family terminal
- **THEN** the photo's placement is deleted from the screen and does not persist over the list

#### Scenario: Leaving the article frees the terminal image cache

- **WHEN** the reader leaves an article whose native photo was displayed on a kitty-family terminal
- **THEN** the terminal receives a delete-by-id for the photo's stable id using the data-freeing delete form (`d=I`), freeing its cached image data as well as its placement

#### Scenario: Leaving the article on Ghostty also clears all placements

- **WHEN** the reader leaves an article whose native photo was displayed on a Ghostty terminal
- **THEN** the frame that leaves the article view additionally emits a delete-all clear (`d=a`) so no placement survives even if the by-id delete is ignored

#### Scenario: Opening another article does not stack photos

- **WHEN** an article with a native photo is opened after an earlier article's photo was displayed on a kitty-family terminal
- **THEN** the earlier placement has been deleted, so the new photo renders alone without stacking over the old one

#### Scenario: Scrolling the image out of view deletes the placement

- **WHEN** the article is scrolled so the photo's rows leave the visible viewport window on a kitty-family terminal (including landing on the caption, which leaves the caption visible)
- **THEN** the placement is deleted so the photo does not float over the caption or body text, and scrolling the photo back into view re-transmits it

#### Scenario: OSC 1337 images need no delete

- **WHEN** frames are rendered on an OSC 1337 terminal
- **THEN** no kitty delete sequence is emitted, because cell-bound inline images are erased by the frame repaint

### Requirement: Stable image identity per render size

On kitty graphics-protocol terminals, the system SHALL identify each native
image render by a stable id derived from both the image URL and the render's
pixel dimensions, so two renders of the same image at different sizes have
distinct ids. The system SHALL record the id each composed native image block
was rendered under. When a render is superseded by a re-render at a different
geometry, the system SHALL delete the prior render's image by its id before
transmitting the new one, so the terminal's image cache holds at most the
current render of each image rather than accumulating one entry per geometry.

#### Scenario: Different render sizes get distinct ids

- **WHEN** the same image URL is rendered natively at two different content
  widths on a kitty-family terminal
- **THEN** the two renders carry different image ids

#### Scenario: Re-render at a new size deletes the prior id

- **WHEN** a native image is re-rendered at a new geometry after a prior render
  at a different size
- **THEN** the terminal receives a delete for the prior render's id before the
  new render's transmit

#### Scenario: Same size reuses the same id

- **WHEN** the same image URL is rendered natively at the same content width
  again
- **THEN** the render reuses the same image id rather than minting a new one

### Requirement: Native render payload is bounded

The system SHALL cap the encoded pixel dimensions of each native image render
at a fixed maximum edge length — a constant independent of the terminal's
reported cell pixel size and well below the current absolute cap — so a
full-width photo's inline-image payload stays bounded on any terminal. The
encoded render SHALL still fill the image's cell box (the terminal scales it to
the box), so the displayed photo is unchanged in size, placement, and aspect
ratio.

#### Scenario: Full-width photo encodes at the cap

- **WHEN** a full-width photo's cell-pixel box exceeds the maximum encoded edge
  length
- **THEN** the render is encoded with the longer edge at the cap, preserving
  aspect ratio, and the displayed photo still fills the same cell box

#### Scenario: Displayed photo unchanged by the encode cap

- **WHEN** a photo is rendered natively with the bounded encode resolution
- **THEN** the photo occupies the same cells, centered the same way, with the
  same aspect ratio as an uncapped encode

### Requirement: Kitty frames re-show by placement reference

On kitty graphics-protocol terminals, a frame that displays a native image
whose payload was already transmitted in this session SHALL re-show it with a
placement reference that names the cached image id, without re-transmitting the
base64 payload. A full transmit SHALL occur only on the first frame that
displays the image, after the terminal's cached image data was freed (the image
scrolled out of view and its data deleted), or when the render is superseded at
a new render size. The placement-reference escape SHALL carry the same stable
image id as the full transmit, so identity tracking and delete-by-id cleanup are
unchanged.

#### Scenario: Photo stays visible across frames without re-transmitting

- **WHEN** a native photo is fully visible and the frame re-renders (for example
  on each keystroke while the photo is in view)
- **THEN** the frames after the first re-show the photo by placement reference
  and do not re-send its base64 payload

#### Scenario: Scroll-out frees data and scroll-back re-transmits

- **WHEN** a native photo scrolls out of view (its data deleted) and is later
  scrolled back into view
- **THEN** the frame that re-displays it sends a full transmit once, and
  subsequent frames re-show it by placement reference

#### Scenario: Re-render at a new size transmits then references

- **WHEN** a native photo is re-rendered at a new render size and displayed
- **THEN** the first frame transmits the new payload and later frames re-show it
  by placement reference, with the prior size's id deleted as today

### Requirement: Native render reuses the cached decode

The native render of an article image SHALL reuse the decoded image already
cached by the halfblock path when one is present, rather than re-decoding the
stored photo bytes a second time. The photo bytes SHALL be decoded at most once
per image per session: the halfblock render's decoded image feeds the native
render, and only when no decoded image is cached does the native path decode
the source itself. The decoded image reused by the native path SHALL respect
the same width cap as the halfblock path, so both paths render from the same
capped bitmap.

#### Scenario: Native render scales the halfblock's decoded image

- **WHEN** an article image's halfblock block was rendered from a decoded image
  that remains in the decoded-image cache and the photo's native render runs
- **THEN** the native render scales and encodes the cached decoded image and
  does not re-decode the stored bytes

#### Scenario: Native render decodes when no cache exists

- **WHEN** the native render runs for an image whose decoded image is not
  cached (for example a fresh session re-reading stored bytes)
- **THEN** the native path decodes the source itself and the render succeeds
  from the stored bytes

#### Scenario: Render output unchanged by decode reuse

- **WHEN** a photo renders natively from the reused decoded image
- **THEN** the rendered output is identical to rendering from a fresh decode of
  the same source

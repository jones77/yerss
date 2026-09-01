## MODIFIED Requirements

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
SHALL be cached in memory keyed by content width, so re-opening an article at
the same content width SHALL display the photo immediately from cache without
re-downloading or re-rendering. When the photo fetch fails, the block SHALL
remain as the final render. Resizing the terminal SHALL re-scale the native
render from the cached photo without re-fetching, asynchronously and without
blocking the UI. The native render SHALL decode the source at the session's
maximum decoded dimension and reuse the decoded image across the height-fit
iterations, and when the cached photo bytes were evicted from the bounded photo
cache it SHALL re-read the photo from the stored database bytes rather than
failing. The fetched photo SHALL be persisted to the database (in the unified
image table) so a later open re-renders natively from the stored bytes without a
network request; the rendered native output SHALL remain an in-memory render
cache keyed by content width.

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

#### Scenario: Viewport-height resize reuses the render

- **WHEN** the terminal is resized so only the viewport height changes while the content width stays the same
- **THEN** the existing native render is reused rather than re-decoded or re-encoded

#### Scenario: Photo persisted for later re-render

- **WHEN** a full-image-capable terminal fetches a lead photo
- **THEN** the photo bytes are written to the database so a later open re-renders natively without re-fetching

#### Scenario: Opening a high-resolution photo does not block the UI

- **WHEN** a full-image-capable terminal opens an article whose lead photo is high-resolution (a large source image)
- **THEN** the article text renders immediately and the photo's decode, scale, and encode happen in the background rather than on the UI thread

#### Scenario: Re-opened article shows the cached native render instantly

- **WHEN** the user re-opens an article whose photo was already rendered natively at the current content width
- **THEN** the photo appears immediately from the in-memory render cache, with no network request and no re-decode or re-encode

#### Scenario: Photo cache eviction re-derives from storage

- **WHEN** the raw-photo cache evicted an article photo before its native render reads it
- **THEN** the render re-reads the photo from the stored database bytes instead of failing, and the halfblock block remains the fallback only when no stored bytes exist

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
ever shown for the life of the terminal tab. Frames that display the native
image block SHALL NOT emit the delete (their re-transmission with the stable
placement id replaces the placement), and a scrolled-out image SHALL be
re-transmitted when scrolled back into view. Terminals whose inline images are
cell-bound (OSC 1337) SHALL NOT emit any delete sequence, since the repaint
itself erases those images.

#### Scenario: Backing out of an article deletes the photo

- **WHEN** the reader returns from an article showing a native photo to the list view on a kitty-family terminal
- **THEN** the photo's placement is deleted from the screen and does not persist over the list

#### Scenario: Leaving the article frees the terminal image cache

- **WHEN** the reader leaves an article whose native photo was displayed on a kitty-family terminal
- **THEN** the terminal receives a delete-by-id for the photo's stable id, freeing its cached image data

#### Scenario: Opening another article does not stack photos

- **WHEN** an article with a native photo is opened after an earlier article's photo was displayed on a kitty-family terminal
- **THEN** the earlier placement has been deleted, so the new photo renders alone without stacking over the old one

#### Scenario: Scrolling the image out of view deletes the placement

- **WHEN** the article is scrolled so the photo's rows leave the visible viewport window on a kitty-family terminal (including landing on the caption, which leaves the caption visible)
- **THEN** the placement is deleted so the photo does not float over the caption or body text, and scrolling the photo back into view re-transmits it

#### Scenario: OSC 1337 images need no delete

- **WHEN** frames are rendered on an OSC 1337 terminal
- **THEN** no kitty delete sequence is emitted, because cell-bound inline images are erased by the frame repaint

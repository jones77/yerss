## Purpose

Defines how article images — the publisher-attached lead image and inline
`<img>` images in the body — are identified, credited, and drawn on screen
(halfblock art or native terminal images), replacing the previously fragmented
image-rendering, inline-image, lead-dedup, and decode-cap specs.

## ADDED Requirements

### Requirement: Inline image discovery

The system SHALL discover inline `<img>` elements in an article's HTML content
and extract their source URLs in document order, so each can be rendered in
place. Discovery SHALL be performed when the article is opened, not at refresh
time. The system SHALL NOT treat inline images as the lead image; the lead image
continues to come from the publisher-attached enclosure or media:thumbnail
metadata.

#### Scenario: Inline images discovered in document order

- **WHEN** an article's content contains two `<img>` elements with sources `a.jpg` and `b.jpg` in that order
- **THEN** the system discovers `a.jpg` followed by `b.jpg` as inline images

#### Scenario: No inline images

- **WHEN** an article's content contains no `<img>` elements
- **THEN** the system discovers no inline images and renders the article as before

### Requirement: Canonical source image identity

The system SHALL identify an image URL's canonical source by stripping CDN
transform/rewrite prefixes and recovering the underlying source. Two image URLs
SHALL be treated as the same source photo when their canonical sources are
equal, even when their full fetch URLs differ. When an image URL contains a
URL-encoded `http` or `https` source segment inside a CDN fetch path, the system
SHALL decode that trailing segment and treat it as the canonical source. An
image URL with no recognizable transform wrapper SHALL use its full URL as its
own canonical source.

#### Scenario: Substack fetch variants resolve to one source

- **WHEN** an article's enclosure URL is
  `https://substackcdn.com/image/fetch/$s_!X!,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2F...%2Fimages%2Fac2c7f0f-_1600x900.jpeg`
  and a body image URL is
  `https://substackcdn.com/image/fetch/$s_!X!,w_2400,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2F...%2Fimages%2Fac2c7f0f-_1600x900.jpeg`
- **THEN** the two URLs resolve to the same canonical source
  (`https://.../images/ac2c7f0f-_1600x900.jpeg`), so the system treats them as
  the same photo

#### Scenario: Plain URL is its own source

- **WHEN** an image URL has no CDN transform wrapper (for example
  `https://cdn.example.com/photos/a.jpg`)
- **THEN** the URL's canonical source is the URL itself

### Requirement: Duplicate lead image suppressed

When an article has an enclosure/lead image and the lead image's canonical
source matches the canonical source of any inline image in the article body, the
system SHALL NOT render the lead image as a top-of-article block. The system
SHALL render the first body image whose canonical source matches the lead in
place at its natural position in the article flow, and SHALL skip any later
inline image whose canonical source matches an already-rendered image. When the
lead image is suppressed, the system SHALL NOT fetch or persist the lead image
separately at the lead position; the body copy is fetched and persisted at its
own position.

#### Scenario: Enclosure matches first body image

- **WHEN** an article's enclosure and its first body image share a canonical
  source
- **THEN** no lead block renders at the top, the first body image renders in
  place, and the photo is fetched and stored once

#### Scenario: Enclosure matches a later body image

- **WHEN** an article's enclosure shares a canonical source with a body image
  that is not the first image in the body
- **THEN** no lead block renders at the top and the matching body image renders
  at its natural position alongside the other body images

#### Scenario: No body image matches the lead

- **WHEN** an article's enclosure shares a canonical source with no body image
- **THEN** the lead image renders as a top-of-article block as before

#### Scenario: Repeated body image still renders once

- **WHEN** the article body contains the same source photo multiple times
- **THEN** the first occurrence renders and later occurrences are skipped,
  whether or not the lead image was also suppressed

### Requirement: Image attribution

The system SHALL render a centered attribution directly beneath each rendered
image (lead or inline), with no blank line between the image and the attribution.
The attribution text SHALL be resolved in order: first the `figcaption` of the
figure that directly contains the image and carries its own figcaption — used
only when the image is the figure's last `<img>` in document order (a figure's
caption belongs to its last image) — then the image's alt text, then the text of
any markdown link that wraps the image, otherwise falling back to `photo:
<source>` derived by the source-identifier rules. The system SHALL NOT apply a
document-global credit fallback: an image SHALL NOT borrow another image's
caption — neither the first figure with a caption nor the first element whose
class indicates a credit — for the lead image or any inline image, so an image
that has no caption of its own renders without one. When an image sits inside a
captioned figure whose caption belongs to a later image in that figure (a
photo-grid), the image SHALL render with no caption at all: it SHALL NOT borrow
the figure's caption and SHALL NOT fall back to `photo: <source>`, because the
fallback applies only to images with no caption source of their own. An image's
figcaption SHALL be its centered attribution, and the article body SHALL NOT
also render that caption text as a separate left-aligned paragraph. The
attribution SHALL render at a standard caption width — a fixed fraction of the
content width (such as 70%) that does not depend on the photo's own width —
wrapped onto additional centered lines when the credit is longer and centered
beneath the image, and a blank line SHALL follow the attribution before the
article content resumes. The combined image-plus-attribution block height SHALL
be capped so the full block plus at least one line of body text fits within the
viewport height, and the height fit SHALL reserve the wrapped line count of the
caption the block actually composes with, so a long caption cannot overflow the
reserved space and inflate the block past the viewport.

#### Scenario: Inline image attribution

- **WHEN** an inline image renders
- **THEN** a centered attribution appears directly beneath it, using the directly-containing figure's caption, the alt text, the wrapping link's text, or the `photo: <source>` fallback in that order

#### Scenario: Gallery photo shows its own figure's caption

- **WHEN** a photo gallery contains photos each inside its own figure with its own figcaption (nested figures within a gallery wrapper, as in a WordPress `wp-block-gallery`), and the article renders
- **THEN** each photo renders its own figcaption centered directly beneath it, and the figcaption text does not also appear as a left-aligned body paragraph elsewhere in the article

#### Scenario: Inline image without its own caption has none

- **WHEN** an inline image has no alt text and is not inside a figure with a caption or credit, while another image elsewhere in the article is
- **THEN** the inline image renders without an attribution rather than displaying the other image's caption

#### Scenario: Photo-grid first image renders without a caption

- **WHEN** an inline image is the first of two `<img>` elements inside a `<figure>` whose `figcaption` labels the second (`<figure><img src="a.jpg"><img src="b.jpg"><figcaption>…</figcaption></figure>`)
- **THEN** the first image (a.jpg) renders with no attribution at all — it does not display the figure's caption and does not fall back to `photo: <source>`

#### Scenario: Photo-grid last image shows the figure's caption

- **WHEN** an inline image is the last of multiple `<img>` elements inside a `<figure>` whose `figcaption` labels it
- **THEN** the last image renders with the figure's caption centered beneath it

#### Scenario: Linked promo image uses its link text as the caption

- **WHEN** an article's markup wraps an inline image and banner text in a hyperlink (`<a href="/promo"><img src="logo.jpg"><p>Read Our Complete Coverage</p>…</a>`)
- **THEN** the image renders with the banner text ("Read Our Complete Coverage…") centered beneath it as its caption, and no leftover `[`, link text, or `](…)` fragments render around the block

#### Scenario: Attribution taken from the article HTML

- **WHEN** the article's HTML contains a `figcaption` credit associated with the image's figure
- **THEN** the attribution line shows that credit text, centered beneath the photo

#### Scenario: Lead image in a shared captioned figure renders without a caption

- **WHEN** the lead image is not the last `<img>` of a figure whose `figcaption` labels a later image in the same figure
- **THEN** no attribution renders beneath the lead image, with no `photo: <source>` fallback

#### Scenario: Attribution falls back to the source organization

- **WHEN** the article's HTML contains no credit for the image and the article link host is `theintercept.com`
- **THEN** the attribution line shows `photo: theintercept`

#### Scenario: Attribution rendered at a standard caption width

- **WHEN** an image narrower than the content width renders with a long credit
- **THEN** the attribution wraps to the standard caption width (a fixed fraction of the content width), centered beneath the image, rather than wrapping to the photo's own width

#### Scenario: Attribution wraps at the standard caption width

- **WHEN** the attribution text is longer than the standard caption width (for example a long credit beneath a narrow photo)
- **THEN** the attribution wraps onto additional centered lines within the standard caption width — a fixed fraction of the content width that does not depend on the photo's own width — and the photo is re-rendered smaller so the whole block plus at least one line of body text still fits the viewport

#### Scenario: Long caption does not inflate the block past the viewport

- **WHEN** an image's credit is long enough that wrapping it at the standard caption width takes several lines
- **THEN** the image is scaled down so the image plus the full wrapped attribution and one line of body text fit within the viewport height

#### Scenario: Long shared gallery caption wraps compactly

- **WHEN** an article renders two inline photos that share a long identical caption and each photo is narrower than the content width
- **THEN** each attribution wraps to the standard caption width beneath its own photo, and neither image-plus-caption block exceeds the viewport height budget

#### Scenario: Inline caption centered at the standard width

- **WHEN** an inline image with a short caption renders
- **THEN** the caption is centered at the standard caption width beneath the image, not stretched to the photo's width

#### Scenario: Blank line after the attribution

- **WHEN** an image with attribution is rendered
- **THEN** exactly one blank line separates the attribution from the first body line

### Requirement: Halfblock image rendering

When image rendering is enabled and an article has an image URL, the system
SHALL render the image inline — the lead image below the reader header and above
the body, and each inline image interleaved with the surrounding text — as a
block of content that scrolls with the article. On a terminal without native
image support the image SHALL be rendered using a Unicode halfblocks
representation (half-block characters with ANSI color) sized to fit the content
width while preserving aspect ratio, and SHALL be horizontally centered within
the content width. The halfblock block SHALL be served from the stored text
block when one exists at the current content width, re-rendered from the
in-memory decoded image when one is cached, or re-fetched and re-rendered when
neither is available. When image rendering is disabled (config `off` or
`--ascii` mode), inline content images SHALL render as `[alt]` placeholders (the
image's alt text, or `[image]` when none is available) and no lead image block
or attribution SHALL render.

#### Scenario: Lead image rendered below the header and above the body

- **WHEN** image rendering is enabled and an article with an enclosure image URL is opened
- **THEN** the image is rendered as a halfblocks block positioned below the URL/title/author header lines and above the body text, sized to the content width with aspect ratio preserved

#### Scenario: Inline image rendered in place

- **WHEN** an article has body text before and after an inline image, and the image loads
- **THEN** the image block appears between the surrounding text paragraphs, not at the top of the article

#### Scenario: Image is horizontally centered

- **WHEN** a rendered image occupies fewer columns than the content width
- **THEN** the image block is centered with equal left and right padding within the content area

#### Scenario: Inline image rendered as alt text

- **WHEN** image rendering is disabled and an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

#### Scenario: Inline image alt text with markdown characters renders literally

- **WHEN** image rendering is disabled and an article contains an `<img>` element with alt text "a *b* [c]"
- **THEN** the reader displays `[a *b* [c]]` literally, with the `*` and `[`/`]` shown as plain characters and no styled or link formatting applied

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

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
centered within the content width exactly like the halfblock block, and SHALL be
part of the scrolling content with the photo attribution centered beneath. The
rendered native output SHALL be cached in memory keyed by render size — the
content width together with the viewport-height cap — so re-opening an article
at the same size SHALL display the photo immediately from cache without
re-downloading or re-rendering. When the photo fetch fails, the block SHALL
remain as the final render. Resizing the terminal SHALL re-scale the native
render from the cached photo without re-fetching, asynchronously and without
blocking the UI. The native render SHALL decode the source once per render and
reuse the decoded image across the height-fit iterations, and when the cached
photo bytes are unavailable it SHALL re-read the photo from the stored database
bytes rather than failing.

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

#### Scenario: Viewport-height resize re-renders the block

- **WHEN** the terminal is resized so only the viewport height changes while the content width stays the same
- **THEN** the native render is re-rendered for the new viewport-height cap from the cached photo, without a network request

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
leaves the article view. Frames that display the native image block SHALL NOT
emit the delete, and a scrolled-out image SHALL be re-transmitted when scrolled
back into view. Terminals whose inline images are cell-bound (OSC 1337) SHALL
NOT emit any delete sequence.

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

- **WHEN** the article is scrolled so the photo's rows leave the visible viewport window on a kitty-family terminal
- **THEN** the placement is deleted so the photo does not float over the caption or body text, and scrolling the photo back into view re-transmits it

#### Scenario: OSC 1337 images need no delete

- **WHEN** frames are rendered on an OSC 1337 terminal
- **THEN** no kitty delete sequence is emitted, because cell-bound inline images are erased by the frame repaint

### Requirement: Partially visible native photo shows a preview

On a full-image-capable terminal, when an article's native photo is only partly
inside the visible viewport window (its top is inside the window and its bottom
below the fold, or its bottom inside and its top above the fold), the system
SHALL fill the photo's visible rows with a rendered preview of that portion of
the photo rather than leaving the rows blank. A photo that is taller than the
viewport and therefore can never be shown in full SHALL still show its visible
rows as a preview. The preview SHALL match the halfblock path's behavior of
showing the visible slice of a poking-in image instead of a blank strip.

#### Scenario: Tall photo top-aligned shows its visible rows

- **WHEN** a native photo taller than the viewport is scrolled into view so its top is aligned with the top of the viewport and its bottom is below the fold
- **THEN** the visible rows render a preview of the top portion of the photo, and no blank strip appears

#### Scenario: Poking-in photo shows a preview at the fold

- **WHEN** a native photo is only partially scrolled into view so part of it pokes in from the top or bottom edge
- **THEN** the visible rows show a preview of that portion of the photo rather than blank space

#### Scenario: Fully visible photo is unchanged

- **WHEN** a native photo fits entirely within the viewport
- **THEN** the photo renders natively as before, with no preview substitution

### Requirement: Stable image identity per render size

On kitty graphics-protocol terminals, the system SHALL identify each native
image render by a stable id derived from both the image URL and the render's
pixel dimensions, so two renders of the same image at different sizes have
distinct ids. The system SHALL record the id each composed native image block
was rendered under. When a render is superseded by a re-render at a different
geometry, the system SHALL delete the prior render's image by its id before
transmitting the new one.

#### Scenario: Different render sizes get distinct ids

- **WHEN** the same image URL is rendered natively at two different content widths on a kitty-family terminal
- **THEN** the two renders carry different image ids

#### Scenario: Re-render at a new size deletes the prior id

- **WHEN** a native image is re-rendered at a new geometry after a prior render at a different size
- **THEN** the terminal receives a delete for the prior render's id before the new render's transmit

#### Scenario: Same size reuses the same id

- **WHEN** the same image URL is rendered natively at the same content width again
- **THEN** the render reuses the same image id rather than minting a new one

### Requirement: Native render payload is bounded

The system SHALL cap the encoded pixel dimensions of each native image render
at a fixed maximum edge length — a constant independent of the terminal's
reported cell pixel size — so a full-width photo's inline-image payload stays
bounded on any terminal. The encoded render SHALL still fill the image's cell
box, so the displayed photo is unchanged in size, placement, and aspect ratio.

#### Scenario: Full-width photo encodes at the cap

- **WHEN** a full-width photo's cell-pixel box exceeds the maximum encoded edge length
- **THEN** the render is encoded with the longer edge at the cap, preserving aspect ratio, and the displayed photo still fills the same cell box

#### Scenario: Displayed photo unchanged by the encode cap

- **WHEN** a photo is rendered natively with the bounded encode resolution
- **THEN** the photo occupies the same cells, centered the same way, with the same aspect ratio as an uncapped encode

### Requirement: Kitty frames re-show by placement reference

On kitty graphics-protocol terminals, a frame that displays a native image
whose payload was already transmitted in this session SHALL re-show it with a
placement reference that names the cached image id, without re-transmitting the
base64 payload. A full transmit SHALL occur only on the first frame that
displays the image, after the terminal's cached image data was freed, or when
the render is superseded at a new render size. The placement-reference escape
SHALL carry the same stable image id as the full transmit.

#### Scenario: Photo stays visible across frames without re-transmitting

- **WHEN** a native photo is fully visible and the frame re-renders (for example on each keystroke while the photo is in view)
- **THEN** the frames after the first re-show the photo by placement reference and do not re-send its base64 payload

#### Scenario: Scroll-out frees data and scroll-back re-transmits

- **WHEN** a native photo scrolls out of view (its data deleted) and is later scrolled back into view
- **THEN** the frame that re-displays it sends a full transmit once, and subsequent frames re-show it by placement reference

#### Scenario: Re-render at a new size transmits then references

- **WHEN** a native photo is re-rendered at a new render size and displayed
- **THEN** the first frame transmits the new payload and later frames re-show it by placement reference, with the prior size's id deleted as today

### Requirement: Native render reuses the cached decode

The native render of an article image SHALL reuse the decoded image already
cached by the halfblock path when one is present, rather than re-decoding the
stored photo bytes a second time. The photo bytes SHALL be decoded at most once
per image per session: the halfblock render's decoded image feeds the native
render, and only when no decoded image is cached does the native path decode
the source itself. The decoded image reused by the native path SHALL respect
the same width cap as the halfblock path.

#### Scenario: Native render scales the halfblock's decoded image

- **WHEN** an article image's halfblock block was rendered from a decoded image that remains in the decoded-image cache and the photo's native render runs
- **THEN** the native render scales and encodes the cached decoded image and does not re-decode the stored bytes

#### Scenario: Native render decodes when no cache exists

- **WHEN** the native render runs for an image whose decoded image is not cached (for example a fresh session re-reading stored bytes)
- **THEN** the native path decodes the source itself and the render succeeds from the stored bytes

#### Scenario: Render output unchanged by decode reuse

- **WHEN** a photo renders natively from the reused decoded image
- **THEN** the rendered output is identical to rendering from a fresh decode of the same source

### Requirement: Decoded image width cap

The system SHALL decode every article image — the lead image and each inline
image — at at most 2048 pixels wide for in-memory caching and rendering. A
source image wider than 2048 pixels SHALL be downscaled to 2048 pixels wide,
preserving aspect ratio, before it is cached in the decoded-image cache and
before it is rendered by either the halfblock or the native render path. An
image at or below 2048 pixels wide SHALL be decoded and cached at its natural
size. The cap SHALL apply to the decoded in-memory bitmap only: the stored photo
bytes SHALL remain the original compressed source, unchanged.

#### Scenario: Oversized source decoded at the cap

- **WHEN** an article image's source is wider than 2048 pixels
- **THEN** the image is decoded and held in the decoded-image cache at 2048 pixels wide, not at its full source width

#### Scenario: Image at or below the cap decoded at natural size

- **WHEN** an article image's source width is 2048 pixels or less
- **THEN** the image is decoded and cached at its natural size

#### Scenario: Cap applies to the halfblock render

- **WHEN** an article image wider than 2048 pixels renders as halfblock art
- **THEN** the halfblock block is rendered from the capped decoded image, not the full-resolution source

#### Scenario: Cap applies to the native render

- **WHEN** an article image wider than 2048 pixels renders natively on a full-image-capable terminal
- **THEN** the native render is scaled to the display box from the capped decoded image

#### Scenario: Stored bytes remain the compressed original

- **WHEN** an image wider than 2048 pixels is fetched and its decoded bitmap is capped
- **THEN** the photo bytes written to the database are the original compressed source, unchanged by the cap

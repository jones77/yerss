## Purpose

Defines the asynchronous lifecycle of article images: how image data is stored
in the unified image table, when images are fetched and decoded (viewport-driven
and bounded), how the article re-composes when image loads land, and how the
viewport snaps around image blocks during scrolling. Replaces the previously
fragmented image-storage, load-scheduling, and recomposition specs.

## Requirements

### Requirement: Unified image storage

The system SHALL persist an article's images in a unified per-article image
table (`article_images`) that stores both the rendered text block — the Unicode
halfblocks art lines (with ANSI color) joined by newlines — and the full photo
bytes, associated with the article and ordered by position (the lead image is
position 0; each inline image follows at positions after the lead). When an
article is opened, a stored block SHALL render without any network request to
the image host, and a stored photo SHALL re-render natively or as halfblocks at
the current width without a re-fetch. The system SHALL persist the full photo
bytes to the database (not only the rendered block). When the stored block's
rendering width differs from the current content width, the system SHALL
re-render a block at the current width from the stored photo without
re-fetching; when the widths match, the stored block SHALL be displayed
directly. The lead image's URL SHALL remain denormalized on the article row so
the list view and refresh can detect a lead image without reading the image
table.

#### Scenario: Stored block renders without network

- **WHEN** an article with a stored image block is opened
- **THEN** the block renders immediately from the database and no request is made to the image host

#### Scenario: Full photo bytes are persisted

- **WHEN** a lead image is fetched from the network
- **THEN** both the rendered text block and the full photo bytes are written to the article's image table

#### Scenario: Block width matches content width

- **WHEN** the stored block's rendering width equals the current content width
- **THEN** the stored block is displayed as-is

#### Scenario: Block width mismatches content width

- **WHEN** the stored block's rendering width differs from the current content width
- **THEN** the block is re-rendered at the current width from the stored photo without a network request, and the new block is persisted

#### Scenario: Native render from stored photo without re-fetch

- **WHEN** a full-image-capable terminal opens an article whose photo was previously stored
- **THEN** the photo is re-rendered natively from the stored bytes without a network request

#### Scenario: Inline images persisted after the lead

- **WHEN** an inline image is fetched
- **THEN** its rendered block and full photo bytes are stored in the article's image table at a position after the lead image

#### Scenario: Inline images render from storage

- **WHEN** the application is restarted and an article with previously stored inline images is opened
- **THEN** the inline images render from the stored bytes without a network request

### Requirement: Legacy photo purge

When an article still has legacy data in the old `articles.image_block` and
`image_data` columns (from before the unified image table), the system SHALL
migrate it into the `article_images` table as position 0 on first startup, and
SHALL thereafter read only from the image table. The old columns SHALL be
dropped when the SQLite version supports column removal, or left empty and
unused otherwise.

#### Scenario: Legacy image migrated to the unified table

- **WHEN** a database upgraded from before the unified table still has a stored block or raw bytes in the old columns
- **THEN** the block and bytes are copied into the image table as position 0 and the old columns are no longer read

### Requirement: Photo bytes written once on acquisition

The system SHALL write an article image's full photo bytes to the database only
when the photo is first acquired — from the network, or from a legacy-column
migration — and SHALL NOT re-write those bytes when the image is later
re-rendered. When a stored block's rendering width differs from the current
content width, the system SHALL update only the rendered block and its width,
leaving the stored photo bytes untouched. A re-render that changes the block
SHALL therefore not grow the write-ahead log by the size of the photo.

#### Scenario: Width change re-renders without rewriting the photo

- **WHEN** an article's stored image is re-rendered at a new content width from its stored photo
- **THEN** the block and width columns are updated and the stored photo bytes are not rewritten

#### Scenario: First acquisition writes photo and block together

- **WHEN** a photo is fetched from the network for an article
- **THEN** the block and the full photo bytes are written together to the article's image row

#### Scenario: Block-only update preserves the photo

- **WHEN** a block-only update is applied to an image row that already has stored photo bytes
- **THEN** the photo bytes remain unchanged and still render natively or as halfblocks on a later open

### Requirement: Suppressed lead is not stored separately

When an article's lead image is suppressed because it duplicates an inline body
image, the system SHALL NOT persist the lead image as a distinct article image
at position 0. The body image that shares its canonical source SHALL be fetched
and persisted at its own position only, so the photo's bytes are stored once
rather than once as the lead and again as an inline image.

#### Scenario: Suppressed lead photo stored once

- **WHEN** an article whose lead image duplicates its first body image is opened and the body image is fetched
- **THEN** the photo bytes are stored at the body image's position and no separate position-0 row duplicates them

#### Scenario: Stored rows read by canonical source

- **WHEN** a stored article image is loaded for rendering
- **THEN** the image is matched by URL (and canonical source) rather than by its stored position, so a previously-stored lead image that is now suppressed is not misread as the body copy

### Requirement: Viewport-driven load frontier

The system SHALL load only the images an open article renders that are within or
near the visible viewport (the load frontier), and SHALL load additional images
as the reader scrolls toward them. Images outside the frontier SHALL NOT be
fetched, decoded, or rendered. An image whose block or photo is already
available — from the in-memory caches or the stored image table — SHALL compose
immediately when its position is reached, without a fetch, regardless of the
frontier.

#### Scenario: Opening a photo-heavy article loads only nearby images

- **WHEN** an article with images both near and far below the reading position is opened
- **THEN** the text renders immediately and only the images within or near the viewport are loaded; the distant images are not fetched, decoded, or rendered

#### Scenario: Scrolling toward an image loads it

- **WHEN** the reader scrolls the article so a previously unloaded image approaches the viewport
- **THEN** the image's load starts and its block composes in place when it arrives

#### Scenario: Stored block off-screen composes without fetching

- **WHEN** an image outside the frontier has a stored block at the current width
- **THEN** the block composes when scrolled into view with no network request

### Requirement: Bounded load concurrency

The system SHALL bound the number of concurrent image fetch, decode, and render
operations so a photo-heavy article does not start one operation per image at
once. The bound SHALL be a small constant independent of the article's image
count, and SHALL NOT change the rendered output.

#### Scenario: Many-image article starts at most the bound

- **WHEN** an article with more images than the concurrency bound is opened
- **THEN** no more than the bound of fetch/decode/render operations run at once, and remaining loads start as prior ones complete

#### Scenario: Bounded loads render identically

- **WHEN** image loads run under the concurrency bound
- **THEN** every image renders the same block, caption, and native output as an unbounded load would

### Requirement: Lazy native render

The native render of a photo SHALL be performed only when the photo's block is
in or near the visible viewport. A photo outside the viewport SHALL keep its
halfblock placeholder and SHALL be rendered natively when it approaches the
viewport. A photo whose native render is already cached at the current render
size SHALL display natively immediately whenever it comes into view, without a
new render.

#### Scenario: Off-screen photo keeps its placeholder

- **WHEN** a photo's block is composed but its rows are outside the viewport
- **THEN** the block remains a halfblock placeholder and no native render is started for it

#### Scenario: Scrolling a photo into view triggers the native render

- **WHEN** the reader scrolls an unrendered photo into or near the viewport
- **THEN** the photo's native render starts and replaces the placeholder when it completes

#### Scenario: Cached native render shows immediately on scroll

- **WHEN** the reader scrolls a photo whose native render is already cached at the current render size into view
- **THEN** the photo displays natively immediately, with no new render

### Requirement: Async image load lifecycle

The system SHALL fetch article images asynchronously so that opening an article
never blocks the UI on a network request. Opening an article with an image URL
SHALL first attempt to load the block from the cache hierarchy (in-memory
decoded image, then the stored database block when its width matches the content
width); only on a miss SHALL it issue a network fetch. On a successful block
load, the article re-renders with the block inserted. On a network success, the
image is decoded, the rendered block and the full photo bytes are persisted to
the database, and the decoded image is cached in memory. On a full-image-capable
terminal the fetched photo bytes SHALL additionally be cached in memory and a
native photo render SHALL replace the block. On failure (network error, timeout,
or non-image content type), the system SHALL produce a failure message and
render the article without an image block. The decoded image SHALL be cached in
memory by URL for the session so that resizing the viewport does not re-fetch or
re-decode from the database. When the terminal is resized while an article with
a loaded image is open, the system SHALL re-render the image block to the new
content width using the cached decoded image (or the cached photo on a native
terminal) without re-fetching. When an image loads after the user has scrolled
past the insertion point, the system SHALL preserve the user's reading position
by adjusting the viewport offset by the number of inserted image-block lines.

#### Scenario: Article opens with image loading asynchronously

- **WHEN** the user opens an article that has an image URL and image rendering is enabled
- **THEN** the article text renders immediately and the image fetch runs in the background

#### Scenario: Image loads and appears inline

- **WHEN** the background image fetch succeeds while the article is open and the user is at the top of the article
- **THEN** the image block is inserted at the top of the content and becomes visible

#### Scenario: Image fetch fails gracefully

- **WHEN** the background image fetch fails or times out
- **THEN** the article renders without an image block and no error is surfaced to the user

#### Scenario: Resize re-renders image without re-fetch

- **WHEN** the terminal is resized while an article with a loaded image is open
- **THEN** the image block is re-rendered to the new content width using the cached image and no network request is made

#### Scenario: Scroll position preserved on late image load

- **WHEN** an image loads after the user has scrolled into the body text past the image insertion point
- **THEN** the viewport offset is increased by the number of image-block lines so the user's reading position is unchanged

#### Scenario: Re-opened article uses cached image

- **WHEN** the user closes and re-opens an article whose image was previously fetched
- **THEN** the image block renders immediately from the in-memory cache without a network request

#### Scenario: Restarted article uses stored block

- **WHEN** the application is restarted and the user opens an article whose block was previously rendered and stored in the database at the current content width
- **THEN** the image block renders from the stored block without making a network request to the image host

#### Scenario: Restarted article uses stored photo

- **WHEN** the application is restarted and the user opens an article whose photo was previously stored at a different content width
- **THEN** the image is re-rendered at the current width from the stored photo bytes without making a network request to the image host

### Requirement: Article derivation cached per article and geometry

The system SHALL cache an article's derived display content — the converted
markdown body with inline-image sentinels, the discovered inline-image list, the
harvested hyperlinks, and the rendered markdown segments — keyed by article id
together with the content geometry (content width and viewport height) it was
derived for. Re-composing an article at the same geometry SHALL reuse the cached
derivation instead of re-running HTML-to-markdown conversion, markdown
rendering, or link harvesting on the UI thread. A derivation cache entry SHALL
be invalidated or re-derived when the content geometry changes. Re-composition
SHALL re-render only the changed image block and re-join it into the cached
body, preserving the article's text, header, attribution, and reading position.

#### Scenario: Image load recomposes without re-deriving the body

- **WHEN** an open article's image block finishes loading and the article re-composes at the same content geometry
- **THEN** the recompose reuses the cached markdown conversion, rendered text segments, and harvested links, and only the newly available image block is rendered and inserted

#### Scenario: Resize re-derives at the new geometry

- **WHEN** the terminal is resized while an article with a cached derivation is open
- **THEN** the article re-derives at the new content geometry and the cached derivation is replaced, exactly as a fresh composition would

#### Scenario: Derived content identical after recompose

- **WHEN** an image load lands and re-composes an article whose other content is unchanged
- **THEN** the article text, header, attributions, and link set render identically to the pre-load composition

### Requirement: Image-load messages recompose once with the cached derivation

Each image-load message (block, photo, native render, or failure arrival) for
the open article SHALL recompose the article at most once, after the message,
rather than as part of the per-message cache work. The UI thread SHALL NOT run a
full article derivation as part of processing any single image message; the
per-message work SHALL be bounded to updating the in-memory image caches and
marking the article for recomposition, and the recompose SHALL reuse the cached
derivation so HTML-to-markdown conversion, markdown rendering, and link
harvesting never re-run while images load. The reading position SHALL be
preserved across the recompose: the scroll offset SHALL be bumped only by the
number of lines inserted above the viewport's first visible line.

#### Scenario: Several images land and each recomposes once

- **WHEN** multiple image-load messages for the open article are processed
- **THEN** each message recomposes the article at most once, after the message, reusing the cached derivation, and the reading position is preserved

#### Scenario: Single image message recomposes once

- **WHEN** a single image-load message for the open article is processed
- **THEN** the article re-composes once for that message, with the reading position preserved

#### Scenario: Block landing above the fold bumps the offset

- **WHEN** an image-load message inserts an image block above the viewport's first visible line while the user is reading further down
- **THEN** the scroll offset is increased by the number of inserted lines so the user's reading position is unchanged

#### Scenario: Block landing below the fold does not shift the reader

- **WHEN** an image-load message materializes an image block entirely below the viewport's first visible line
- **THEN** the scroll offset is unchanged and the user's reading position does not move

### Requirement: List view loads without article bodies

The list view SHALL be populated from a query that retrieves only the columns
the list renders — article id, title, read status, publication time, link, and
feed URL — and SHALL NOT select the article's full HTML content column. Loading
the list after returning from an article view SHALL therefore not copy every
article body from the database.

#### Scenario: List load after backing out does not fetch article bodies

- **WHEN** the user returns from an article view to the list view
- **THEN** the list is populated without selecting the stored HTML content of the listed articles

#### Scenario: List rendering unchanged by the slimmer query

- **WHEN** the list view is displayed
- **THEN** every article row shows the same time, title, read state, and source identifier as before the query slimming

### Requirement: Recompose serves the cached block at the current width

When the open article re-composes after an image-load message, an image whose
rendered block is already cached at the current content width SHALL be served
from the rendered-block cache without re-rendering from the decoded image on the
UI thread. The UI thread SHALL NOT re-run a halfblock render for an image whose
width-matched block is already available; re-rendering from the decoded image
SHALL occur only when no width-matched block is cached (a resize to a new
content width, or a first load). Serving the cached block SHALL produce output
identical to re-rendering.

#### Scenario: Recompose serves a cached inline block

- **WHEN** an article with an inline image whose block is already cached at the current content width re-composes (for example another image's load lands)
- **THEN** the recompose inserts the cached block and does not re-run the halfblock render for that image on the UI thread

#### Scenario: Resize re-renders from the decoded image

- **WHEN** the content width changes so the cached block's width no longer matches
- **THEN** the recompose re-renders the block at the new width from the decoded image

#### Scenario: First load renders once and then serves

- **WHEN** an image's first load renders its block and later messages for other images re-compose the article at the same width
- **THEN** the later recomposes serve the rendered block from the cache instead of re-rendering it

### Requirement: Load-frontier recomposes stay bounded

Advancing the load frontier as the reader scrolls SHALL recompose the article at
most once per image-load message, reusing the cached derivation exactly as the
existing recompose contract requires, and SHALL NOT add per-message recompose
work beyond that contract. The reading position SHALL be preserved when the
frontier advance's loads land and re-compose the article.

#### Scenario: Scroll-advanced loads recompose once per message

- **WHEN** the reader scrolls toward a group of unloaded images and their loads land
- **THEN** each image-load message recomposes the article at most once, reusing the cached derivation

#### Scenario: Reading position preserved across frontier recomposes

- **WHEN** a scroll-advanced image load lands and re-composes the article
- **THEN** the reader's position in the article is preserved, matching the existing recompose position rules

### Requirement: Recompose never snaps forward while the reader is in the header

When the open article re-composes after an image-load message and the viewport's
first visible line lies within the article's header region (the URL line, the
title, the author line, and the blank separators), the scroll offset SHALL be
bumped only by the lines inserted above it and SHALL NOT snap forward to an
image block whose top materialized inside the window. A photo below the header
is the article's normal open state, not a scrolled-past image, so a load landing
there SHALL NOT scroll the header off. The snap-forward rule for an inline image
whose top materializes mid-window SHALL apply only once the reader is past the
header.

#### Scenario: Header-region image load does not scroll the header off

- **WHEN** the reader is at the top of an article with a photo below the header (a suppressed lead whose URL also opens the body, or any first inline image) and that photo's load lands and re-composes the article
- **THEN** the viewport stays at the top with the header (URL line, title, and author) still visible, and does not jump to the photo's top

#### Scenario: Reader past the header still snaps to a materialized inline image

- **WHEN** the reader is scrolled into the body text past the header and an inline image's top materializes inside the window below the offset
- **THEN** the recompose snaps the viewport forward to the image's top, exactly as before

### Requirement: A top-of-article restart restore lands at the very top

When a saved reader selection records an article offset of zero (the user quit
at the very top of the article), restoring the selection SHALL place the
viewport at offset 0 regardless of the saved content anchor. The top is
unambiguous across reflows, and the anchor mechanism would land one line lower,
which a photo load landing just afterwards would snap forward past and scroll
the header off. Restoring to 0 keeps the header — including the expanded mode's
URL line — visible through the subsequent image loads.

#### Scenario: Restart at the top of an expanded article stays at the top

- **WHEN** the user quits at the very top of an expanded-mode article whose photo sits below the header, and restarts so the mode is restored, the article restores at the default size, the terminal reports its real size, and then the photo load lands
- **THEN** the viewport remains at offset 0 through the resize and the photo load, and the expanded URL header line is the first visible line

#### Scenario: Restore below the top still uses the content anchor

- **WHEN** the saved selection's article offset is non-zero (the user quit scrolled into the article)
- **THEN** the restore continues to relocate the saved content anchor as before, with the photo and text anchor paths unchanged

### Requirement: Atomic image scroll behavior

When an image block (a lead image or an inline image) is rendered, the system
SHALL ensure the photo's lines are either fully visible within the viewport or
fully scrolled out of view; it SHALL never display a partially clipped photo.
Each image block SHALL apply these rules independently, using its own caption.
The snapping applies to single-line moves (the line keys and the mouse wheel);
page, half-page, and goto-bottom moves land wherever they land, even mid-photo.

Each image block SHALL define two snap boundaries. The top boundary is the
offset at which the image's first line sits at the viewport's first row; the
bottom boundary is the offset at which the block's last line — the last line of
the wrapped caption — sits at the viewport's last row. A block whose bottom
boundary equals its top boundary (it fills the viewport exactly) has a single
snap position.

For the lead image, a downward single-line scroll that would move the viewport
offset into the photo's line range SHALL first rise the photo flush to the
viewport top (the top boundary) when the move entered the photo from above, and
the following downward scroll SHALL snap the offset to the first line of the
next image or paragraph, skipping the whole photo-and-caption block and the
space below it out of view as one unit. The lead image's boundaries are
pre-consumed — it was shown on open — so it has no entry stages, but it is never
skipped past unseen in a single keystroke. When the article's lead image URL is
suppressed because it also appears as an inline image, the first composed block
follows the pre-consumed lead transitions only when it is composed at the top of
the content, directly below the header, and shown on open; when the first
content element is text (a promo banner reusing the featured photo), the first
composed block is a regular inline image that was not shown on open and SHALL
follow the inline-image boundary transitions.

For an inline image, a downward single-line scroll follows the boundary
transitions: a scroll that leaves the image partially visible in the window
SHALL snap the block's bottom boundary; the next downward scroll SHALL snap the
image's first line to the viewport top (the top boundary); and a scroll landing
in its photo range skips the entire block onto the first line of the next image
or paragraph. The rise from a bottom boundary SHALL take precedence over the
photo-range skip. An inline image whose block is shorter than the viewport SHALL
snap to the bottom boundary as soon as a downward single-line scroll brings its
top into the window from below; the next downward scroll SHALL rise it to the
top boundary. The fully-visible skip SHALL NOT apply to inline images. Upward
single-line scrolls mirror the transitions: an image entering from above snaps
its top to the viewport top, the next upward scroll snaps its bottom boundary,
and an upward scroll that cuts the block's last line below the fold SHALL snap
it fully below the fold on the next move. When a block fills the viewport
exactly, there is no distinct bottom stage: the upward scroll from the
top-aligned position SHALL scroll the image fully out of view rather than
looping. When an exit or scroll-off position would land inside a preceding image
block's photo range, the system SHALL snap to the immediately preceding image's
first line instead. On a full-image-capable terminal, when the lead photo is
fully visible and the user scrolls upward with the offset past the article top,
the system SHALL snap the offset to the article top in a single step.

When an image load re-composes the article, the system SHALL NOT leave the
viewport offset inside an image block's photo range or with an image's top
materialized mid-window; it SHALL snap the offset to that image's first line.
When the offset is already on the article's last screen, re-composition SHALL
keep the offset at the bottom so the article's end stays visible. On a
native-image terminal, a photo's placement SHALL be drawn only when its rows are
fully contained in the viewport window; a partially visible photo's placement is
deleted and its transmit suppressed so the image never paints over the article
border.

#### Scenario: Downward scroll rises the lead photo flush to the viewport top

- **WHEN** the lead image block occupies content lines 4 through 10 (photo lines 4 through 7 plus attribution lines 8 through 10), the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 4 (the photo's first line at the viewport top), and the next downward scroll snaps to 12 (the first line of the next paragraph, past the blank line below the block)

#### Scenario: Inline image snaps to the viewport bottom when its top enters from below

- **WHEN** an inline image block occupies content lines 20 through 24 (photo lines 20 through 22 plus attribution lines 23 through 24), the viewport height is 15, and a downward scroll brings the image's top into view from below
- **THEN** the viewport offset snaps to 10 (the block's bottom boundary, its last line at the viewport bottom) so the image and its full caption are visible at the bottom of the screen

#### Scenario: Wrapped caption with its last line below the fold snaps to the bottom boundary

- **WHEN** an inline image has a caption that wraps to three lines (photo lines 20 through 22, caption lines 23 through 25), the viewport height is 15, and a downward scroll leaves the image's top inside the window with the caption's last line (25) below the fold while its first line (23) is already visible
- **THEN** the viewport offset snaps to 11 (the block's bottom boundary) so the caption's last line (25) sits at the viewport bottom and the full wrapped caption is visible

#### Scenario: Short inline image entered from below snaps to the viewport bottom

- **WHEN** a downward single-line scroll brings a short inline image's top into the window from below (its block is shorter than the viewport), the viewport height is 15, and the image's top line is 10
- **THEN** the viewport offset snaps to the block's bottom boundary so the image and its full caption are visible at the bottom of the screen, and the next downward scroll snaps the image's first line to the viewport top (the top boundary)

#### Scenario: Inline image snaps to the viewport top on the next downward scroll

- **WHEN** an inline image is fully visible with its last line at the viewport bottom (at its bottom boundary) and the user scrolls down one line
- **THEN** the viewport offset snaps to the image's first line so the image moves to the top of the viewport

#### Scenario: Inline image skips past its caption when a scroll lands in its photo

- **WHEN** the viewport offset lands inside an inline image's photo range after a downward scroll that started at or inside the image
- **THEN** the viewport offset snaps to the first line of the next image or paragraph, past the block's blank separator, skipping the whole image-and-caption block and the space below it out of view as one unit

#### Scenario: Upward scroll reveals the whole block at once, never a caption-only frame

- **WHEN** an upward single-line scroll brings an image block's last line into the window from above at the window's first row
- **THEN** the viewport offset snaps to the image's first line so the whole image and its full caption are visible with the image at the top of the screen

#### Scenario: Upward scroll moves a top-aligned inline image to the bottom

- **WHEN** an inline image is fully visible with its top at the viewport top (at its top boundary) and the user scrolls up one line
- **THEN** the viewport offset snaps so the image's bottom boundary holds: the image and its wrapped caption are visible at the bottom of the screen with the caption's last line at the viewport bottom

#### Scenario: Upward scroll snaps a partially-scrolled-off inline image below the fold

- **WHEN** an upward scroll cuts an inline image's last line below the fold, leaving it partially visible at the bottom edge
- **THEN** the next upward scroll snaps the offset so the image's top sits at the fold, scrolling the image fully out of view

#### Scenario: Exit or scroll-off reveals the immediately preceding image

- **WHEN** image blocks are spaced closer than the viewport height and a scroll-off position for one image (its top at the fold) would land inside a preceding image's photo range
- **THEN** the viewport offset snaps to the immediately preceding image's first line so it is revealed rather than left partially clipped

#### Scenario: Page-down lands naturally even mid-photo

- **WHEN** the user pages down and the resulting viewport offset lands inside an image block's photo range
- **THEN** no snapping occurs; the offset stays where the page-down landed, even with a partially clipped photo

#### Scenario: Image load does not leave the offset inside a photo

- **WHEN** an inline image finishes loading and the article re-composes while the viewport offset falls inside the image's photo range or with its top within the window
- **THEN** the viewport offset snaps to the image's first line so the image is shown, never left partially clipped

#### Scenario: Image load at the article bottom keeps the end visible

- **WHEN** an image finishes loading and the article re-composes while the user is on the article's last screen, and the bottom offset falls inside a photo's range whose caption is cut off below the fold
- **THEN** the offset stays at the article bottom; it does not snap up to the photo's first line

#### Scenario: Native photo never paints over the border

- **WHEN** on a native-image terminal an image block's rows are only partially within the viewport window (its top above the fold or its bottom below it)
- **THEN** the image's placement is deleted and its transmit suppressed, so no image rows draw over the article border, and the visible rows render the image's halfblock preview rather than a blank strip

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below an image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally

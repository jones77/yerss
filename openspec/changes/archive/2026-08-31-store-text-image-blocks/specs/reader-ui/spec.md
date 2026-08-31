## MODIFIED Requirements

### Requirement: Article lead image rendering

When image rendering is enabled and an article has a publisher-attached image
URL (captured from `<enclosure>` or `media:thumbnail` during feed parsing), the
system SHALL render the image inline below the reader header (the URL, title,
and author lines) and above the rendered article content, as a block of content
that scrolls with the article text. On a terminal without native image support
the image SHALL be rendered using a Unicode halfblocks representation
(half-block characters with ANSI color) sized to fit the content width while
preserving aspect ratio, and SHALL be horizontally centered within the content
width. The halfblock block SHALL be served from the stored text block when one
exists at the current content width, re-rendered from the in-memory decoded
image when one is cached, or re-fetched and re-rendered when neither is
available. On a full-image-capable terminal the halfblock block SHALL be shown
first as an immediate placeholder and SHALL be replaced by a native inline
photo render once the real photo loads. Directly beneath the image or photo,
with no blank line between them, the system SHALL render a photo attribution
constrained to the photo's own width (wrapped onto additional centered lines
when the credit is longer) and centered beneath it; a blank line SHALL follow
the attribution before the article content resumes. The attribution text SHALL
be the photo's credit extracted from the article's stored HTML when present — a
`figcaption` associated with the figure containing the lead image, otherwise an
element whose class indicates a credit line — and SHALL fall back to
`photo: <source>`, where `<source>` is derived by the same source-identifier
rules used for the list view's organization column. The combined
image-plus-attribution block height SHALL be capped so that the full block plus
at least one line of body text fits within the viewport height; when the block
is taller than this cap — for example because a wrapped attribution needs more
lines — it SHALL be scaled down to fit. When image rendering is disabled
(config `off` or `--ascii` mode), no image block and no attribution SHALL be
rendered and the article SHALL render exactly as it does without an image URL.

#### Scenario: Lead image rendered below the header and above the body

- **WHEN** image rendering is enabled and an article with an enclosure image URL is opened
- **THEN** the image is rendered as a halfblocks block positioned below the URL/title/author header lines and above the body text, sized to the content width with aspect ratio preserved

#### Scenario: Image is horizontally centered

- **WHEN** a rendered lead image occupies fewer columns than the content width
- **THEN** the image block is centered with equal left and right padding within the content area

#### Scenario: Attribution centered directly beneath the image

- **WHEN** a lead image is rendered with an attribution
- **THEN** the attribution line is centered within the photo's own width and there is no blank line between the image and the attribution

#### Scenario: Attribution wraps under the photo

- **WHEN** the attribution text is longer than the rendered photo's width (for example after the photo is scaled down by the height cap)
- **THEN** the attribution wraps onto additional lines within the photo's width, each centered beneath the photo, and the photo is re-rendered smaller so the whole block plus at least one line of body text still fits the viewport

#### Scenario: Attribution taken from the article HTML

- **WHEN** the article's HTML contains a `figcaption` credit associated with the lead image's figure
- **THEN** the attribution line shows that credit text, centered beneath the photo

#### Scenario: Attribution falls back to the source organization

- **WHEN** the article's HTML contains no credit for the lead image and the article link host is `theintercept.com`
- **THEN** the attribution line shows `photo: intercept`

#### Scenario: Blank line after the attribution

- **WHEN** a lead image with attribution is rendered
- **THEN** exactly one blank line separates the attribution from the first body line

#### Scenario: Lead image scrolls with content

- **WHEN** a rendered lead image is visible and the user scrolls down
- **THEN** the image block scrolls upward with the body text as a single scrolling content region

#### Scenario: Lead image capped to viewport

- **WHEN** an article's enclosure image, when aspect-fit to the content width, would make the combined image-plus-attribution block taller than the available viewport height minus one line
- **THEN** the image is scaled down so that the full block and at least one line of body text are visible without scrolling

#### Scenario: Stored block renders on open

- **WHEN** image rendering is enabled and an article with a stored image block at the current content width is opened
- **THEN** the stored block renders immediately as the lead image without a network request

#### Scenario: Native terminal swaps block for the photo

- **WHEN** a full-image-capable terminal opens an article and the fetched photo loads
- **THEN** the halfblock placeholder block is replaced by a native inline photo render with the attribution still beneath

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

### Requirement: Async image load lifecycle

The system SHALL fetch article lead images asynchronously so that opening an
article never blocks the UI on a network request. Opening an article with an
image URL SHALL first attempt to load the block from the cache hierarchy
(in-memory decoded image, then the stored database block when its width matches
the content width); only on a miss SHALL it issue a network fetch. On a
successful block load, the article re-renders with the block inserted. On a
network success, the image is decoded, the rendered block is persisted to the
database, and the decoded image is cached in memory; the raw photo bytes SHALL
NOT be persisted. On a full-image-capable terminal the fetched photo bytes SHALL
additionally be cached in memory and a native photo render SHALL replace the
block. On failure (network error, timeout, non-image content type, or exceeded
byte limit), the system SHALL produce a failure message and render the article
without an image block. The decoded image SHALL be cached in memory by URL for
the session so that resizing the viewport does not re-fetch or re-decode from
the database. When the terminal is resized while an article with a loaded image
is open, the system SHALL re-render the image block to the new content width
using the cached decoded image (or the cached photo on a native terminal)
without re-fetching. When an image loads after the user has scrolled past the
insertion point, the system SHALL preserve the user's reading position by
adjusting the viewport offset by the number of inserted image-block lines.

#### Scenario: Article opens with image loading asynchronously

- **WHEN** the user opens an article that has an enclosure image URL and image rendering is enabled
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

### Requirement: Atomic image scroll behavior

When a lead image block is rendered inline, the system SHALL ensure the photo's
lines are either fully visible within the viewport or fully scrolled out of
view; it SHALL never display a partially clipped photo. When a downward scroll
would move the viewport offset into the photo's line range, the system SHALL
snap the offset to the photo's attribution line when one is rendered — landing
on the caption so it can be read — and to the line immediately after the block
when no attribution is rendered, skipping the photo in a single keystroke.
Scrolling within the attribution lines SHALL behave as normal line scrolling.
When an upward scroll would move the offset into the block's range, the system
SHALL snap the offset to the block's first line, revealing the full photo and
its attribution. This snapping SHALL apply to all scroll operations: line,
page, half-page, and goto-bottom.

#### Scenario: Downward scroll skips the photo onto its caption

- **WHEN** the image block occupies content lines 4 through 10 (photo lines 4 through 7 plus attribution lines 8 through 10), the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 8 (the attribution's first line): the photo is fully scrolled out and the caption is visible at the top of the viewport

#### Scenario: Downward scroll within the caption is normal

- **WHEN** the viewport offset is on an attribution line and the user scrolls down one line
- **THEN** no snapping occurs and the next attribution or body line scrolls into view

#### Scenario: Downward scroll without attribution skips the block

- **WHEN** the image block has no attribution, the photo occupies lines 4 through 9, the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 10 (the line after the block) and the photo is fully scrolled out

#### Scenario: Upward scroll reveals full image with attribution

- **WHEN** the image block occupies content lines 4 through 10, the viewport offset is on a line within or below the block, and the user scrolls up one line
- **THEN** the viewport offset snaps to 4 (the block's first line) and the full image and its attribution are visible

#### Scenario: Page-down skips the photo onto the caption

- **WHEN** the photo is partially within the viewport after a page-down operation
- **THEN** the offset snaps to the attribution's first line when an attribution is rendered, otherwise past the block

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below the image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally
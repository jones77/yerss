## MODIFIED Requirements

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
block. On failure (network error, timeout, or non-image content type), the
system SHALL produce a failure message and render the article without an image
block. The decoded image SHALL be cached in memory by URL for the session so
that resizing the viewport does not re-fetch or re-decode from the database.
When the terminal is resized while an article with a loaded image is open, the
system SHALL re-render the image block to the new content width using the cached
decoded image (or the cached photo on a native terminal) without re-fetching.
When an image loads after the user has scrolled past the insertion point, the
system SHALL preserve the user's reading position by adjusting the viewport
offset by the number of inserted image-block lines.

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

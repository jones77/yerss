## Purpose

Schedules article image fetches, decodes, and native renders on demand so the
work is proportional to the visible viewport rather than the article's length.

## Requirements

### Requirement: Viewport-driven load frontier

The system SHALL load only the images an open article renders that are within or
near the visible viewport (the load frontier), and SHALL load additional images
as the reader scrolls toward them. Images outside the frontier SHALL NOT be
fetched, decoded, or rendered. An image whose block or photo is already
available — from the in-memory caches or the stored image table — SHALL compose
immediately when its position is reached, without a fetch, regardless of the
frontier.

#### Scenario: Opening a photo-heavy article loads only nearby images

- **WHEN** an article with images both near and far below the reading position
  is opened
- **THEN** the text renders immediately and only the images within or near the
  viewport are loaded; the distant images are not fetched, decoded, or rendered

#### Scenario: Scrolling toward an image loads it

- **WHEN** the reader scrolls the article so a previously unloaded image
  approaches the viewport
- **THEN** the image's load starts and its block composes in place when it
  arrives

#### Scenario: Stored block off-screen composes without fetching

- **WHEN** an image outside the frontier has a stored block at the current width
- **THEN** the block composes when scrolled into view with no network request

### Requirement: Lazy native render

The native render of a photo SHALL be performed only when the photo's block is
in or near the visible viewport. A photo outside the viewport SHALL keep its
halfblock placeholder and SHALL be rendered natively when it approaches the
viewport. A photo whose native render is already cached at the current render
size SHALL display natively immediately whenever it comes into view, without a
new render.

#### Scenario: Off-screen photo keeps its placeholder

- **WHEN** a photo's block is composed but its rows are outside the viewport
- **THEN** the block remains a halfblock placeholder and no native render is
  started for it

#### Scenario: Scrolling a photo into view triggers the native render

- **WHEN** the reader scrolls an unrendered photo into or near the viewport
- **THEN** the photo's native render starts and replaces the placeholder when it
  completes

#### Scenario: Cached native render shows immediately on scroll

- **WHEN** the reader scrolls a photo whose native render is already cached at
  the current render size into view
- **THEN** the photo displays natively immediately, with no new render

### Requirement: Bounded load concurrency

The system SHALL bound the number of concurrent image fetch, decode, and render
operations so a photo-heavy article does not start one operation per image at
once. The bound SHALL be a small constant independent of the article's image
count, and SHALL NOT change the rendered output.

#### Scenario: Many-image article starts at most the bound

- **WHEN** an article with more images than the concurrency bound is opened
- **THEN** no more than the bound of fetch/decode/render operations run at once,
  and remaining loads start as prior ones complete

#### Scenario: Bounded loads render identically

- **WHEN** image loads run under the concurrency bound
- **THEN** every image renders the same block, caption, and native output as an
  unbounded load would
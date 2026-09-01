## Purpose

Bounds the reader's in-memory image caches so a long session of opening many
articles — each with photos — keeps memory flat instead of growing without
limit, and refuses to hold oversized source images full-size in memory.

## Requirements

### Requirement: Bounded in-memory image caches

The system SHALL bound every session image cache — decoded images, rendered
halfblock blocks, raw fetched photo bytes, and rendered native blocks — by a
total byte budget, and SHALL evict least-recently-used entries when a new entry
would exceed the budget. Once evicted, an entry SHALL be re-derived from the
stored database image or re-fetched rather than failing; eviction SHALL be
invisible to the user apart from a possible re-render or re-fetch cost. The
budget SHALL be a fixed constant so memory usage stays flat regardless of how
many distinct articles and images are opened in a session.

#### Scenario: Many opened images stay within budget

- **WHEN** a session opens many distinct articles with distinct images so the
  cumulative cached size would exceed the budget
- **THEN** the oldest least-recently-used entries are evicted and total cached
  bytes stay at or below the budget

#### Scenario: Evicted image is re-rendered without a network request

- **WHEN** an image whose decoded bytes were evicted is opened again and its
  photo bytes are still stored in the database
- **THEN** the image is re-rendered from the stored photo without a network
  request

#### Scenario: Native render evicted and re-rendered

- **WHEN** a native render is evicted from the native render cache and the same
  article is re-opened at the same content width
- **THEN** the native render is re-derived from the cached photo bytes rather
  than failing or fetching

#### Scenario: Photo evicted before its native render

- **WHEN** an article opens many photos so the raw-photo cache evicts one
  before its native render reads it
- **THEN** the native render re-reads the photo from the stored database bytes
  instead of failing, so eviction is invisible apart from the re-read cost

#### Scenario: Halfblock placeholder survives decoded-image eviction

- **WHEN** the decoded-image cache evicts an image that still has a stored
  halfblock block rendered for the current content width
- **THEN** the stored block is served as the placeholder at that width rather
  than rendering a blank, so eviction of the decoded buffer does not hide the
  photo

### Requirement: Decoded source resolution cap

The system SHALL NOT hold a decoded image in memory at more than a fixed
maximum source dimension; a source image larger than the cap SHALL be decoded
at the cap's resolution for caching and rendering rather than stored full-size.
This bound SHALL apply to both the native render path and the halfblock render
path, so an oversized or hostile source image cannot allocate a
disproportionately large decoded buffer for the whole session. The native
render path SHALL additionally decode the source once per render and reuse the
decoded image across the height-fit iterations rather than re-decoding the
source bytes for every candidate height.

#### Scenario: Oversized source image decoded at the cap

- **WHEN** an article's image has a source dimension larger than the maximum
  decoded dimension
- **THEN** the image is decoded at the cap's resolution and cached at that
  resolution, not at its full source resolution

#### Scenario: Small image decoded at its natural size

- **WHEN** an article's image source dimension is at or below the maximum
  decoded dimension
- **THEN** the image is decoded and cached at its natural resolution
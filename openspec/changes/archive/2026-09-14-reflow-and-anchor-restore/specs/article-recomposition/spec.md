## MODIFIED Requirements

### Requirement: Image-load messages recompose once with the cached derivation

Each image-load message (block, photo, native render, or failure arrival) for
the open article SHALL recompose the article at most once, after the message,
rather than as part of the per-message cache work. The UI thread SHALL NOT run
a full article derivation as part of processing any single image message; the
per-message work SHALL be bounded to updating the in-memory image caches and
marking the article for recomposition, and the recompose SHALL reuse the cached
derivation so HTML-to-markdown conversion, markdown rendering, and link
harvesting never re-run while images load. The reading position SHALL be
preserved across the recompose: the scroll offset SHALL be bumped only by the
number of lines inserted above the viewport's first visible line, so an image
block that materializes entirely below the viewport top SHALL NOT shift the
reader's position.

#### Scenario: Several images land and each recomposes once

- **WHEN** multiple image-load messages for the open article are processed
- **THEN** each message recomposes the article at most once, after the message,
  reusing the cached derivation, and the reading position is preserved

#### Scenario: Single image message recomposes once

- **WHEN** a single image-load message for the open article is processed
- **THEN** the article re-composes once for that message, with the reading
  position preserved

#### Scenario: Block landing above the fold bumps the offset

- **WHEN** an image-load message inserts an image block above the viewport's
  first visible line while the user is reading further down
- **THEN** the scroll offset is increased by the number of inserted lines so
  the user's reading position is unchanged

#### Scenario: Block landing below the fold does not shift the reader

- **WHEN** an image-load message materializes an image block entirely below the
  viewport's first visible line
- **THEN** the scroll offset is unchanged and the user's reading position does
  not move
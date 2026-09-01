## ADDED Requirements

### Requirement: Photo bytes written once on acquisition

The system SHALL write an article image's full photo bytes to the database only
when the photo is first acquired — from the network, or from a legacy-column
migration — and SHALL NOT re-write those bytes when the image is later
re-rendered. When a stored block's rendering width differs from the current
content width, the system SHALL update only the rendered block and its width,
leaving the stored photo bytes untouched. A re-render that changes the block
SHALL therefore not grow the write-ahead log by the size of the photo.

#### Scenario: Width change re-renders without rewriting the photo

- **WHEN** an article's stored image is re-rendered at a new content width from
  its stored photo
- **THEN** the block and width columns are updated and the stored photo bytes
  are not rewritten

#### Scenario: First acquisition writes photo and block together

- **WHEN** a photo is fetched from the network for an article
- **THEN** the block and the full photo bytes are written together to the
  article's image row

#### Scenario: Block-only update preserves the photo

- **WHEN** a block-only update is applied to an image row that already has
  stored photo bytes
- **THEN** the photo bytes remain unchanged and still render natively or as
  halfblocks on a later open

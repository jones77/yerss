## Purpose

Persists article lead images as compact rendered text blocks rather than raw
photo bytes, and purges legacy stored photos, so the database stays small and
reopening an article renders offline.

## Requirements

### Requirement: Text block image persistence

The system SHALL persist an article's lead image in a unified per-article image
table (`article_images`) that stores both the rendered text block — the Unicode
halfblocks art lines (with ANSI color) joined by newlines — and the full photo
bytes, associated with the article and ordered by position (the lead image is
position 0). When an article is opened, a stored block SHALL render without any
network request to the image host, and a stored photo SHALL re-render natively
or as halfblocks at the current width without a re-fetch. The system SHALL
persist the full photo bytes to the database (not only the rendered block). When
the stored block's rendering width differs from the current content width, the
system SHALL re-render a block at the current width from the stored photo
without re-fetching; when the widths match, the stored block SHALL be displayed
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

### Requirement: Suppressed lead is not stored separately

When an article's lead image is suppressed because it duplicates an inline body
image, the system SHALL NOT persist the lead image as a distinct article image
at position 0. The body image that shares its canonical source SHALL be fetched
and persisted at its own position only, so the photo's bytes are stored once
rather than once as the lead and again as an inline image.

#### Scenario: Suppressed lead photo stored once

- **WHEN** an article whose lead image duplicates its first body image is opened
  and the body image is fetched
- **THEN** the photo bytes are stored at the body image's position and no
  separate position-0 row duplicates them

#### Scenario: Stored rows read by canonical source

- **WHEN** a stored article image is loaded for rendering
- **THEN** the image is matched by URL (and canonical source) rather than by its
  stored position, so a previously-stored lead image that is now suppressed is
  not misread as the body copy

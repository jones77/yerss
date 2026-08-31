## Purpose

Persists article lead images as compact rendered text blocks rather than raw
photo bytes, and purges legacy stored photos, so the database stays small and
reopening an article renders offline.

## ADDED Requirements

### Requirement: Text block image persistence

The system SHALL persist an article's lead image as a rendered text block — the
Unicode halfblocks art lines (with ANSI color) joined by newlines — rather than
the raw compressed photo bytes. The stored block SHALL be associated with the
article. When an article is opened, a stored block SHALL render without any
network request to the image host. The system SHALL NOT persist the raw photo
bytes to the database. When the stored block's rendering width differs from the
current content width, the system SHALL re-fetch the photo and re-render a
block at the current width before displaying it; when the widths match, the
stored block SHALL be displayed directly.

#### Scenario: Stored block renders without network

- **WHEN** an article with a stored image block is opened
- **THEN** the block renders immediately from the database and no request is made to the image host

#### Scenario: Raw photo bytes are not persisted

- **WHEN** a lead image is fetched from the network
- **THEN** the rendered text block is persisted and the raw photo bytes are not written to the database

#### Scenario: Block width matches content width

- **WHEN** the stored block's rendering width equals the current content width
- **THEN** the stored block is displayed as-is

#### Scenario: Block width mismatches content width

- **WHEN** the stored block's rendering width differs from the current content width
- **THEN** the photo is re-fetched and a new block is rendered and persisted at the current width before display

### Requirement: Legacy photo purge

When an article still has legacy raw photo bytes stored in the database (from
before text-block persistence), the system SHALL convert them into a rendered
text block at the current content width, persist the block, and purge the raw
bytes so the database does not retain the photo.

#### Scenario: Legacy photo converted and purged

- **WHEN** an article opened after an upgrade still has raw photo bytes in the database
- **THEN** the photo is decoded, rendered to a block, the block is persisted, and the raw bytes are removed from the database

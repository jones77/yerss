## MODIFIED Requirements

### Requirement: Article persistence across restarts

The system SHALL persist every downloaded article in the SQLite database. Articles SHALL survive application restarts. Read status SHALL be persisted and restored on the next launch. Each article SHALL persist a nullable lead image URL captured from the article's `<enclosure>` or `media:thumbnail` metadata when present; articles without such metadata SHALL store a null image URL. When an article's lead image is fetched for display, the system SHALL persist the raw image bytes in the database so that subsequent opens and application restarts render the image without re-fetching from the network.

#### Scenario: Articles survive restart

- **WHEN** the application is restarted after fetching articles
- **THEN** all previously fetched articles are loaded from the database

#### Scenario: Read status persisted

- **WHEN** an article is marked as read and the application is restarted
- **THEN** the article retains its read status on the next launch

#### Scenario: Lead image URL persisted

- **WHEN** an article with an enclosure image URL is fetched and the application is restarted
- **THEN** the article's image URL is restored from the database on the next launch

#### Scenario: Lead image bytes persisted across restart

- **WHEN** an article's lead image has been fetched and stored, and the application is restarted and the article is re-opened
- **THEN** the image renders from the stored bytes without making a network request to the image host

## ADDED Requirements

### Requirement: Lead image URL extraction from feeds

During feed parsing, the system SHALL extract a lead image URL for each article from the first image-typed `<enclosure>` or the `media:thumbnail` extension when present. When multiple enclosures exist, the system SHALL select the first one whose MIME type is an image type (`image/*`). When no image enclosure or media thumbnail is present, the article SHALL have no lead image URL. The system SHALL NOT extract image URLs from inline `<img>` elements within the article HTML content; only publisher-attached enclosure/thumbnail metadata is captured.

#### Scenario: Enclosure image URL captured

- **WHEN** a feed item has an enclosure with type `image/jpeg` and URL `https://example.com/lead.jpg`
- **THEN** the article's lead image URL is set to `https://example.com/lead.jpg`

#### Scenario: Media thumbnail URL captured

- **WHEN** a feed item has no image enclosure but has a `media:thumbnail` URL `https://example.com/thumb.png`
- **THEN** the article's lead image URL is set to `https://example.com/thumb.png`

#### Scenario: Non-image enclosure ignored

- **WHEN** a feed item has an enclosure with type `audio/mpeg` and no image enclosure or media thumbnail
- **THEN** the article has no lead image URL

#### Scenario: Inline content image not captured

- **WHEN** a feed item has no enclosure or media thumbnail but its HTML content contains `<img src="https://cdn.example.com/track.png">`
- **THEN** the article has no lead image URL

### Requirement: Additive schema migration for image columns

The system SHALL add two nullable columns to the `articles` table: `image_url TEXT` (the publisher-attached image URL) and `image_data BLOB` (the fetched image bytes). The migration SHALL be idempotent: it SHALL check the existing column set before adding each column so that re-running the migration on an already-migrated database is a no-op. The migration SHALL not drop or alter existing columns or data.

#### Scenario: Columns added on first migration

- **WHEN** the database is opened and the `articles` table does not have an `image_url` or `image_data` column
- **THEN** both columns are added as nullable (`TEXT` and `BLOB` respectively) and existing rows retain their data with null image URLs and image data

#### Scenario: Migration idempotent on already-migrated database

- **WHEN** the database is opened and the `articles` table already has both `image_url` and `image_data` columns
- **THEN** the migration makes no changes to the schema

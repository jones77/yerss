## MODIFIED Requirements

### Requirement: Article persistence across restarts

The system SHALL persist every downloaded article in the SQLite database. Articles SHALL survive application restarts. Read status SHALL be persisted and restored on the next launch. Each article SHALL persist a nullable lead image URL captured from the article's `<enclosure>` or `media:thumbnail` metadata when present; articles without such metadata SHALL store a null image URL. Persistence of the fetched image's rendered block and photo bytes is specified by the `image-loading` capability, not here.

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

## REMOVED Requirements

### Requirement: Additive schema migration for image columns

**Reason**: The `articles.image_data` BLOB column (and the migration that added it) was superseded by the unified `article_images` table; the image-column migration requirement is stale and contradicted the current schema.

**Migration**: Image storage and the legacy-column purge are specified by the `image-loading` capability. See `openspec/specs/image-loading/spec.md`.

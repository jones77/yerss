## MODIFIED Requirements

### Requirement: Automatic tagging from feed categories

The system SHALL extract categories from each article using the feed's
category metadata (as provided by gofeed). Each unique category name SHALL be
stored as a tag. Tag names SHALL be unique case-insensitively but
case-preserving: the first spelling ever stored becomes the canonical name
forever, so `energy`, `Energy`, and `ENERGY` resolve to one tag whose spelling
is the first one seen. Tag matching and filtering SHALL be case-insensitive.
Articles SHALL be associated with their categories in the
database, enabling tag-based filtering and browsing. In addition, every
article SHALL be associated with a source-domain tag derived from its link
host — or its feed URL host when the article has no link — using the
registrable organization label (for example `https://www.nytimes.com` →
`nytimes`, `https://a.lot.of.subdomains.nytimes.com` → `nytimes`). Source tags
SHALL be marked as source tags so the tag popup can group news organizations
ahead of categories, and they SHALL filter the list like category tags. When
the source label duplicates a category name (case-insensitively), it SHALL be
associated once.

#### Scenario: Article with categories tagged

- **WHEN** a feed is fetched and an article has categories "tech" and "news"
- **THEN** the article is associated with tags "tech" and "news" in the database

#### Scenario: Article with no categories

- **WHEN** a feed is fetched and an article has no categories
- **THEN** the article is still stored and displayed, and is associated with its source-domain tag derived from its feed or link host

#### Scenario: Article tagged with its source domain

- **WHEN** a feed is fetched and an article links to `https://www.nytimes.com/story/1` from a feed at `https://feeds.example.com/rss`
- **THEN** the article is additionally associated with the tag `nytimes`, so the tag popup lists it and filtering by it shows that article

#### Scenario: Source-domain tag deduplicates against a category

- **WHEN** an article's source-domain label equals one of its category names (case-insensitively)
- **THEN** the article is associated with that name once, not twice

#### Scenario: Tag names are case-insensitive and case-preserving

- **WHEN** articles are tagged `energy`, then `ENERGY`, then `Energy`
- **THEN** the database stores a single `energy` tag (the first spelling seen) with all three articles associated, and filtering by `ENERGY` or `Energy` finds them

#### Scenario: Case-insensitive tag matching

- **WHEN** the user filters by a tag name whose spelling differs in case from the stored tag
- **THEN** the matching articles are returned, because tag matching is case-insensitive
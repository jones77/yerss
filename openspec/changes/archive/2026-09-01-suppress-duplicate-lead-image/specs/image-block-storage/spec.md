## ADDED Requirements

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

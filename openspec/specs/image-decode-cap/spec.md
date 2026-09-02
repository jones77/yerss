## Purpose

Bounds the in-memory size of decoded article images by capping their decoded
width at 2048 pixels, without altering the stored photo bytes.

## Requirements

### Requirement: Decoded image width cap

The system SHALL decode every article image — the lead image and each inline
image — at at most 2048 pixels wide for in-memory caching and rendering. A
source image wider than 2048 pixels SHALL be downscaled to 2048 pixels wide,
preserving aspect ratio, before it is cached in the decoded-image cache and
before it is rendered by either the halfblock or the native render path. An
image at or below 2048 pixels wide SHALL be decoded and cached at its natural
size. The cap SHALL apply to the decoded in-memory bitmap only: the cached and
stored photo bytes SHALL remain the original compressed source, unchanged.

#### Scenario: Oversized source decoded at the cap

- **WHEN** an article image's source is wider than 2048 pixels
- **THEN** the image is decoded and held in the decoded-image cache at 2048 pixels wide, not at its full source width

#### Scenario: Image at or below the cap decoded at natural size

- **WHEN** an article image's source width is 2048 pixels or less
- **THEN** the image is decoded and cached at its natural size

#### Scenario: Cap applies to the halfblock render

- **WHEN** an article image wider than 2048 pixels renders as halfblock art
- **THEN** the halfblock block is rendered from the capped decoded image, not the full-resolution source

#### Scenario: Cap applies to the native render

- **WHEN** an article image wider than 2048 pixels renders natively on a full-image-capable terminal
- **THEN** the native render is scaled to the display box from the capped decoded image

#### Scenario: Stored bytes remain the compressed original

- **WHEN** an image wider than 2048 pixels is fetched and its decoded bitmap is capped
- **THEN** the photo bytes written to the database are the original compressed source, unchanged by the cap
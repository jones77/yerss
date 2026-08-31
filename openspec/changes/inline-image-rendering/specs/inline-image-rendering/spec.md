## Purpose

Renders inline `<img>` images from article content in place, interleaved with the
surrounding text, with attribution and the same native/halfblock rendering and
persistence as the lead image.

## ADDED Requirements

### Requirement: Inline image discovery

The system SHALL discover inline `<img>` elements in an article's HTML content
and extract their source URLs in document order, so each can be rendered in
place. Discovery SHALL be performed when the article is opened, not at refresh
time. The system SHALL NOT treat inline images as the lead image; the lead image
continues to come from the publisher-attached enclosure or media:thumbnail
metadata.

#### Scenario: Inline images discovered in document order

- **WHEN** an article's content contains two `<img>` elements with sources `a.jpg` and `b.jpg` in that order
- **THEN** the system discovers `a.jpg` followed by `b.jpg` as inline images

#### Scenario: No inline images

- **WHEN** an article's content contains no `<img>` elements
- **THEN** the system discovers no inline images and renders the article as before

### Requirement: Inline image rendering with attribution

The system SHALL render each discovered inline image in place within the article
body — interleaved with the surrounding text — using the same rendering path as
the lead image: a native inline photo on a full-image-capable terminal (with a
halfblock placeholder first), and halfblock art otherwise, sized to the content
width and centered, capped so the block fits the viewport. Each inline image
SHALL have its own attribution centered directly beneath it, taken from the
image's alt text when present and otherwise falling back to `photo: <source>`
derived by the source-identifier rules.

#### Scenario: Inline image rendered in place

- **WHEN** an article has body text before and after an inline image, and the image loads
- **THEN** the image block appears between the surrounding text paragraphs, not at the top of the article

#### Scenario: Inline image attribution

- **WHEN** an inline image renders
- **THEN** a centered attribution appears directly beneath it, using the alt text or the `photo: <source>` fallback

#### Scenario: Inline image loads asynchronously

- **WHEN** an article with an inline image is opened
- **THEN** the article text renders immediately and the inline image loads in the background

### Requirement: Inline image persistence

The system SHALL persist each inline image's rendered block and full photo bytes
in the article's image table (at positions following the lead image), so a later
open re-renders inline images offline without a network request.

#### Scenario: Inline images persisted

- **WHEN** an inline image is fetched
- **THEN** its rendered block and full photo bytes are stored in the article's image table at a position after the lead image

#### Scenario: Inline images render from storage

- **WHEN** the application is restarted and an article with previously stored inline images is opened
- **THEN** the inline images render from the stored bytes without a network request

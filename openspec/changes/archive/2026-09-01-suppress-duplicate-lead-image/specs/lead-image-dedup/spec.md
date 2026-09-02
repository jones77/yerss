## Purpose

Recognizes when an article's enclosure/lead image is the same underlying source
photo as an inline body image, so the redundant top lead block is suppressed and
the body copy renders once in place.

## ADDED Requirements

### Requirement: Canonical source image identity

The system SHALL identify an image URL's canonical source by stripping CDN
transform/rewrite prefixes and recovering the underlying source. Two image URLs
SHALL be treated as the same source photo when their canonical sources are
equal, even when their full fetch URLs differ. When an image URL contains a
URL-encoded `http` or `https` source segment inside a CDN fetch path, the system
SHALL decode that trailing segment and treat it as the canonical source. An
image URL with no recognizable transform wrapper SHALL use its full URL as its
own canonical source.

#### Scenario: Substack fetch variants resolve to one source

- **WHEN** an article's enclosure URL is
  `https://substackcdn.com/image/fetch/$s_!X!,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2F...%2Fimages%2Fac2c7f0f-_1600x900.jpeg`
  and a body image URL is
  `https://substackcdn.com/image/fetch/$s_!X!,w_2400,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2F...%2Fimages%2Fac2c7f0f-_1600x900.jpeg`
- **THEN** the two URLs resolve to the same canonical source
  (`https://.../images/ac2c7f0f-_1600x900.jpeg`), so the system treats them as
  the same photo

#### Scenario: Plain URL is its own source

- **WHEN** an image URL has no CDN transform wrapper (for example
  `https://cdn.example.com/photos/a.jpg`)
- **THEN** the URL's canonical source is the URL itself

### Requirement: Duplicate lead image suppressed

When an article has an enclosure/lead image and the lead image's canonical
source matches the canonical source of any inline image in the article body, the
system SHALL NOT render the lead image as a top-of-article block. The system
SHALL render the first body image whose canonical source matches the lead in
place at its natural position in the article flow, and SHALL skip any later
inline image whose canonical source matches an already-rendered image. When the
lead image is suppressed, the system SHALL NOT fetch or persist the lead image
separately at the lead position; the body copy is fetched and persisted at its
own position.

#### Scenario: Enclosure matches first body image

- **WHEN** an article's enclosure and its first body image share a canonical
  source
- **THEN** no lead block renders at the top, the first body image renders in
  place, and the photo is fetched and stored once

#### Scenario: Enclosure matches a later body image

- **WHEN** an article's enclosure shares a canonical source with a body image
  that is not the first image in the body
- **THEN** no lead block renders at the top and the matching body image renders
  at its natural position alongside the other body images

#### Scenario: No body image matches the lead

- **WHEN** an article's enclosure shares a canonical source with no body image
- **THEN** the lead image renders as a top-of-article block as before

#### Scenario: Repeated body image still renders once

- **WHEN** the article body contains the same source photo multiple times
- **THEN** the first occurrence renders and later occurrences are skipped,
  whether or not the lead image was also suppressed

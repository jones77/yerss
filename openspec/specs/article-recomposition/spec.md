## Purpose

Defines how the article view re-composes when image loads land: derived display
content is cached per article and geometry, so the expensive HTML-to-markdown
conversion and markdown rendering never re-run on the UI thread, and image-load
messages coalesce into one recompose per update batch.

## Requirements

### Requirement: Article derivation cached per article and geometry

The system SHALL cache an article's derived display content — the converted
markdown body with inline-image sentinels, the discovered inline-image list,
the harvested hyperlinks, and the rendered markdown segments — keyed by article
id together with the content geometry (content width and viewport height) it
was derived for. Re-composing an article at the same geometry SHALL reuse the
cached derivation instead of re-running HTML-to-markdown conversion, markdown
rendering, or link harvesting on the UI thread. A derivation cache entry SHALL
be invalidated or re-derived when the content geometry changes, so a resize
re-renders at the new geometry exactly as today. Re-composition SHALL re-render
only the changed image block and re-join it into the cached body, preserving
the article's text, header, attribution, and reading position.

#### Scenario: Image load recomposes without re-deriving the body

- **WHEN** an open article's image block finishes loading and the article
  re-composes at the same content geometry
- **THEN** the recompose reuses the cached markdown conversion, rendered text
  segments, and harvested links, and only the newly available image block is
  rendered and inserted

#### Scenario: Resize re-derives at the new geometry

- **WHEN** the terminal is resized while an article with a cached derivation is
  open
- **THEN** the article re-derives at the new content geometry and the cached
  derivation is replaced, exactly as a fresh composition would

#### Scenario: Derived content identical after recompose

- **WHEN** an image load lands and re-composes an article whose other content
  is unchanged
- **THEN** the article text, header, attributions, and link set render
  identically to the pre-load composition

### Requirement: Image-load messages recompose once with the cached derivation

Each image-load message (block, photo, native render, or failure arrival) for
the open article SHALL recompose the article at most once, after the message,
rather than as part of the per-message cache work. The UI thread SHALL NOT run
a full article derivation as part of processing any single image message; the
per-message work SHALL be bounded to updating the in-memory image caches and
marking the article for recomposition, and the recompose SHALL reuse the cached
derivation so HTML-to-markdown conversion, markdown rendering, and link
harvesting never re-run while images load. The reading position SHALL be
preserved across the recompose.

#### Scenario: Several images land and each recomposes once

- **WHEN** multiple image-load messages for the open article are processed
- **THEN** each message recomposes the article at most once, after the message,
  reusing the cached derivation, and the reading position is preserved

#### Scenario: Single image message recomposes once

- **WHEN** a single image-load message for the open article is processed
- **THEN** the article re-composes once for that message, with the reading
  position preserved

### Requirement: List view loads without article bodies

The list view SHALL be populated from a query that retrieves only the columns
the list renders — article id, title, read status, publication time, link, and
feed URL — and SHALL NOT select the article's full HTML content column. Loading
the list after returning from an article view SHALL therefore not copy every
article body from the database.

#### Scenario: List load after backing out does not fetch article bodies

- **WHEN** the user returns from an article view to the list view
- **THEN** the list is populated without selecting the stored HTML content of
  the listed articles

#### Scenario: List rendering unchanged by the slimmer query

- **WHEN** the list view is displayed
- **THEN** every article row shows the same time, title, read state, and source
  identifier as before the query slimming

### Requirement: Recompose serves the cached block at the current width

When the open article re-composes after an image-load message, an image whose
rendered block is already cached at the current content width SHALL be served
from the rendered-block cache without re-rendering from the decoded image on
the UI thread. The UI thread SHALL NOT re-run a halfblock render for an image
whose width-matched block is already available; re-rendering from the decoded
image SHALL occur only when no width-matched block is cached (a resize to a new
content width, or a first load). Serving the cached block SHALL produce output
identical to re-rendering, so the composed article is unchanged.

#### Scenario: Recompose serves a cached inline block

- **WHEN** an article with an inline image whose block is already cached at the
  current content width re-composes (for example another image's load lands)
- **THEN** the recompose inserts the cached block and does not re-run the
  halfblock render for that image on the UI thread

#### Scenario: Resize re-renders from the decoded image

- **WHEN** the content width changes so the cached block's width no longer
  matches
- **THEN** the recompose re-renders the block at the new width from the decoded
  image

#### Scenario: First load renders once and then serves

- **WHEN** an image's first load renders its block and later messages for other
  images re-compose the article at the same width
- **THEN** the later recomposes serve the rendered block from the cache instead
  of re-rendering it

### Requirement: Load-frontier recomposes stay bounded

Advancing the load frontier as the reader scrolls SHALL recompose the article at
most once per image-load message, reusing the cached derivation exactly as the
existing recompose contract requires, and SHALL NOT add per-message recompose
work beyond that contract. The reading position SHALL be preserved when the
frontier advance's loads land and re-compose the article.

#### Scenario: Scroll-advanced loads recompose once per message

- **WHEN** the reader scrolls toward a group of unloaded images and their loads
  land
- **THEN** each image-load message recomposes the article at most once, reusing
  the cached derivation

#### Scenario: Reading position preserved across frontier recomposes

- **WHEN** a scroll-advanced image load lands and re-composes the article
- **THEN** the reader's position in the article is preserved, matching the
  existing recompose position rules
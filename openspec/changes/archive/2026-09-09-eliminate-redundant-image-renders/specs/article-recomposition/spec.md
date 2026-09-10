## ADDED Requirements

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
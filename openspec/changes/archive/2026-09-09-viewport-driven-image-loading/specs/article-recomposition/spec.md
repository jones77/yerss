## ADDED Requirements

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
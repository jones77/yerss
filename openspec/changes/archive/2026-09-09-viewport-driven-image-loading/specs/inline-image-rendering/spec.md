## ADDED Requirements

### Requirement: Inline image loading is viewport-driven

An article's inline images SHALL load on demand as the reader's viewport
approaches them rather than all at once on open: the article text SHALL render
immediately, images near the reading position SHALL load in the background, and
images farther down SHALL load when the reader scrolls toward them. An inline
image whose block or photo is already available from the caches or storage SHALL
compose immediately when its position is reached.

#### Scenario: Nearby inline images load, distant ones do not

- **WHEN** an article with many inline images is opened
- **THEN** the article text renders immediately, the inline images near the
  reading position load in the background, and the distant inline images do not
  load until the reader scrolls toward them

#### Scenario: Cached inline image composes on arrival without a fetch

- **WHEN** the reader scrolls to an inline image whose block or photo was
  already stored or cached
- **THEN** the image composes in place without a network request
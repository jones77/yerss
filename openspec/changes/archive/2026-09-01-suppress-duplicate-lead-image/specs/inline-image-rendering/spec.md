## ADDED Requirements

### Requirement: Inline images that duplicate the lead render in place

When an inline image's canonical source matches the article's lead image
canonical source, the system SHALL render that inline image in place at its
natural position in the body and SHALL NOT render a separate lead block at the
top of the article. This supersedes the behavior of suppressing the inline copy
in favor of the lead: the body copy is the one shown. Attribution for the
rendered inline image SHALL resolve by the inline attribution rules (alt text,
figure caption, link text, then `photo: <source>`), including for the case where
the image URL is also the article's lead URL.

#### Scenario: Inline duplicate of the lead renders in the body

- **WHEN** an article's body contains an inline image whose canonical source
  matches the lead image and the article is opened
- **THEN** the image renders at its body position and no duplicate lead block
  renders at the top, with the inline attribution shown beneath it

#### Scenario: Inline duplicate resolves inline attribution

- **WHEN** an inline image whose URL equals the lead URL has a figure caption or
  alt text
- **THEN** the rendered image shows the inline attribution (alt, figure caption,
  or link text in that order) rather than the lead-only attribution fallback

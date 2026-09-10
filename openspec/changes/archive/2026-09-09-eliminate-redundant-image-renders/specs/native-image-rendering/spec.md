## ADDED Requirements

### Requirement: Native render reuses the cached decode

The native render of an article image SHALL reuse the decoded image already
cached by the halfblock path when one is present, rather than re-decoding the
stored photo bytes a second time. The photo bytes SHALL be decoded at most once
per image per session: the halfblock render's decoded image feeds the native
render, and only when no decoded image is cached does the native path decode
the source itself. The decoded image reused by the native path SHALL respect
the same width cap as the halfblock path, so both paths render from the same
capped bitmap.

#### Scenario: Native render scales the halfblock's decoded image

- **WHEN** an article image's halfblock block was rendered from a decoded image
  that remains in the decoded-image cache and the photo's native render runs
- **THEN** the native render scales and encodes the cached decoded image and
  does not re-decode the stored bytes

#### Scenario: Native render decodes when no cache exists

- **WHEN** the native render runs for an image whose decoded image is not
  cached (for example a fresh session re-reading stored bytes)
- **THEN** the native path decodes the source itself and the render succeeds
  from the stored bytes

#### Scenario: Render output unchanged by decode reuse

- **WHEN** a photo renders natively from the reused decoded image
- **THEN** the rendered output is identical to rendering from a fresh decode of
  the same source
## ADDED Requirements

### Requirement: Recompose never snaps forward while the reader is in the header

When the open article re-composes after an image-load message and the viewport's
first visible line lies within the article's header region (the URL line, the
title, the author line, and the blank separators), the scroll offset SHALL be
bumped only by the lines inserted above it and SHALL NOT snap forward to an
image block whose top materialized inside the window. A photo below the header
is the article's normal open state, not a scrolled-past image, so a load landing
there SHALL NOT scroll the header off. The snap-forward rule for an inline
image whose top materializes mid-window SHALL apply only once the reader is past
the header.

#### Scenario: Header-region image load does not scroll the header off

- **WHEN** the reader is at the top of an article with a photo below the header
  (a suppressed lead whose URL also opens the body, or any first inline image)
  and that photo's load lands and re-composes the article
- **THEN** the viewport stays at the top with the header (URL line, title, and
  author) still visible, and does not jump to the photo's top

#### Scenario: Reader past the header still snaps to a materialized inline image

- **WHEN** the reader is scrolled into the body text past the header and an
  inline image's top materializes inside the window below the offset
- **THEN** the recompose snaps the viewport forward to the image's top, exactly
  as before

### Requirement: A top-of-article restart restore lands at the very top

When a saved reader selection records an article offset of zero (the user quit
at the very top of the article), restoring the selection SHALL place the
viewport at offset 0 regardless of the saved content anchor. The top is
unambiguous across reflows, and the anchor mechanism would land one line lower
(the title row minus the expanded header's blank gap), which a photo load
landing just afterwards would snap forward past and scroll the header off.
Restoring to 0 keeps the header — including the expanded mode's URL line —
visible through the subsequent image loads.

#### Scenario: Restart at the top of an expanded article stays at the top

- **WHEN** the user quits at the very top of an expanded-mode article whose
  photo sits below the header, and restarts so the mode is restored, the article
  restores at the default size, the terminal reports its real size, and then the
  photo load lands
- **THEN** the viewport remains at offset 0 through the resize and the photo
  load, and the expanded URL header line is the first visible line

#### Scenario: Restore below the top still uses the content anchor

- **WHEN** the saved selection's article offset is non-zero (the user quit
  scrolled into the article)
- **THEN** the restore continues to relocate the saved content anchor as before,
  with the photo and text anchor paths unchanged
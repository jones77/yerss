## ADDED Requirements

### Requirement: No-op article scrolls do not re-render

The system SHALL memoize the article reader's rendered frame and SHALL NOT
re-render it when a scroll action produces no change in scroll position. A
no-op scroll is one that leaves the viewport offset unchanged — for example
scrolling the mouse wheel up while already at the top of the article, or
scrolling down while already at the bottom — so the article's viewport
composition, native-image re-show and clipped-transmit transforms, and border
rendering are not re-run for that input. The frame the system produces after a
no-op scroll SHALL be byte-identical to the frame it produced before the
no-op. A scroll that changes the scroll position, or any other change to the
state the article view renders (a resize, an image-load recomposition, a
selection change, a glyph-mode change, a link-mode change, or a native-image
transmit transition), SHALL re-render the frame and update the cached frame.

#### Scenario: Wheel up at the top does not re-render

- **WHEN** the article view is scrolled to the top (offset zero) and the user
  scrolls the mouse wheel up
- **THEN** the scroll is a no-op, the article frame is not re-rendered, and the
  rendered output is unchanged

#### Scenario: Wheel down at the bottom does not re-render

- **WHEN** the article view is scrolled to the bottom and the user scrolls the
  mouse wheel down
- **THEN** the scroll is a no-op, the article frame is not re-rendered, and the
  rendered output is unchanged

#### Scenario: No-op frame is byte-identical

- **WHEN** the article view renders once and then receives a no-op scroll
- **THEN** the frame rendered after the no-op is byte-identical to the frame
  rendered before it

#### Scenario: A real scroll re-renders

- **WHEN** the article view is scrolled away from a boundary and the user
  scrolls the mouse wheel in a direction that changes the scroll position
- **THEN** the article frame is re-rendered and the cached frame is updated to
  reflect the new scroll offset

#### Scenario: Image recomposition re-renders

- **WHEN** an image load lands and re-composes the open article while the
  reader is at a boundary scroll position
- **THEN** the article frame is re-rendered to include the newly available
  image block rather than served from the stale cache

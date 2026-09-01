## ADDED Requirements

### Requirement: Stable image identity per render size

On kitty graphics-protocol terminals, the system SHALL identify each native
image render by a stable id derived from both the image URL and the render's
pixel dimensions, so two renders of the same image at different sizes have
distinct ids. The system SHALL record the id each composed native image block
was rendered under. When a render is superseded by a re-render at a different
geometry, the system SHALL delete the prior render's image by its id before
transmitting the new one, so the terminal's image cache holds at most the
current render of each image rather than accumulating one entry per geometry.

#### Scenario: Different render sizes get distinct ids

- **WHEN** the same image URL is rendered natively at two different content
  widths on a kitty-family terminal
- **THEN** the two renders carry different image ids

#### Scenario: Re-render at a new size deletes the prior id

- **WHEN** a native image is re-rendered at a new geometry after a prior render
  at a different size
- **THEN** the terminal receives a delete for the prior render's id before the
  new render's transmit

#### Scenario: Same size reuses the same id

- **WHEN** the same image URL is rendered natively at the same content width
  again
- **THEN** the render reuses the same image id rather than minting a new one

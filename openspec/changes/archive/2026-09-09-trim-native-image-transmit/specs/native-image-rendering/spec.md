## ADDED Requirements

### Requirement: Native render payload is bounded

The system SHALL cap the encoded pixel dimensions of each native image render
at a fixed maximum edge length — a constant independent of the terminal's
reported cell pixel size and well below the current absolute cap — so a
full-width photo's inline-image payload stays bounded on any terminal. The
encoded render SHALL still fill the image's cell box (the terminal scales it to
the box), so the displayed photo is unchanged in size, placement, and aspect
ratio.

#### Scenario: Full-width photo encodes at the cap

- **WHEN** a full-width photo's cell-pixel box exceeds the maximum encoded edge
  length
- **THEN** the render is encoded with the longer edge at the cap, preserving
  aspect ratio, and the displayed photo still fills the same cell box

#### Scenario: Displayed photo unchanged by the encode cap

- **WHEN** a photo is rendered natively with the bounded encode resolution
- **THEN** the photo occupies the same cells, centered the same way, with the
  same aspect ratio as an uncapped encode

### Requirement: Kitty frames re-show by placement reference

On kitty graphics-protocol terminals, a frame that displays a native image
whose payload was already transmitted in this session SHALL re-show it with a
placement reference that names the cached image id, without re-transmitting the
base64 payload. A full transmit SHALL occur only on the first frame that
displays the image, after the terminal's cached image data was freed (the image
scrolled out of view and its data deleted), or when the render is superseded at
a new render size. The placement-reference escape SHALL carry the same stable
image id as the full transmit, so identity tracking and delete-by-id cleanup are
unchanged.

#### Scenario: Photo stays visible across frames without re-transmitting

- **WHEN** a native photo is fully visible and the frame re-renders (for example
  on each keystroke while the photo is in view)
- **THEN** the frames after the first re-show the photo by placement reference
  and do not re-send its base64 payload

#### Scenario: Scroll-out frees data and scroll-back re-transmits

- **WHEN** a native photo scrolls out of view (its data deleted) and is later
  scrolled back into view
- **THEN** the frame that re-displays it sends a full transmit once, and
  subsequent frames re-show it by placement reference

#### Scenario: Re-render at a new size transmits then references

- **WHEN** a native photo is re-rendered at a new render size and displayed
- **THEN** the first frame transmits the new payload and later frames re-show it
  by placement reference, with the prior size's id deleted as today
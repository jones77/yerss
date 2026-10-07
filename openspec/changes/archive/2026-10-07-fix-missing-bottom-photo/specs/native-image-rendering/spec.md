## ADDED Requirements

### Requirement: Partially visible native photo shows a preview, not blank rows

On a full-image-capable terminal, when an article's native photo is only partly
inside the visible viewport window (its top is inside the window and its bottom
below the fold, or its bottom inside and its top above the fold), the system
SHALL fill the photo's visible rows with a rendered preview of that portion of
the photo rather than leaving the rows blank. A photo that is taller than the
viewport and therefore can never be shown in full SHALL still show its visible
rows as a preview, so scrolling to it never displays empty space where the photo
should be. The preview SHALL match the halfblock path's behavior of showing the
visible slice of a poking-in image instead of a blank strip.

#### Scenario: Tall photo top-aligned shows its visible rows

- **WHEN** a native photo taller than the viewport is scrolled into view so its top is aligned with the top of the viewport and its bottom is below the fold
- **THEN** the visible rows render a preview of the top portion of the photo, and no blank strip appears

#### Scenario: Poking-in photo shows a preview at the fold

- **WHEN** a native photo is only partially scrolled into view so part of it pokes in from the top or bottom edge
- **THEN** the visible rows show a preview of that portion of the photo rather than blank space

#### Scenario: Fully visible photo is unchanged

- **WHEN** a native photo fits entirely within the viewport
- **THEN** the photo renders natively as before, with no preview substitution

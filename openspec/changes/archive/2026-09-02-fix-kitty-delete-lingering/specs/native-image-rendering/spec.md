## MODIFIED Requirements

### Requirement: Native image placement cleanup

The system SHALL delete the terminal's native image placements on every frame
that does not display the open article's native image. On kitty graphics
protocol terminals a placement floats above text and survives the erase
commands of a frame repaint, so frames that leave the article view, open
another article, show the placeholder block, or scroll the image out of the
viewport SHALL emit the protocol's delete action for all visible placements.
When the reader leaves the article view entirely — returning to the list,
opening another article, or an article whose native image never composes — the
system SHALL also delete each native image's cached image data by its stable id
so the terminal's image cache is freed rather than accumulating every image
ever shown for the life of the terminal tab. The delete-by-id sequence SHALL
use the protocol's data-freeing form (the uppercase `d=I` delete action), which
removes the placement and releases the terminal's cached image data, not the
placement-only form (`d=i`) that retains the cached data. On terminals whose
kitty-graphics delete handling is known to be partial or unreliable (Ghostty),
the system SHALL additionally emit a delete-all clear (`d=a`) on the frame that
leaves the article view, so a placement that survives the by-id delete cannot
float over the list. Frames that display the native image block SHALL NOT emit
the delete (their re-transmission with the stable placement id replaces the
placement), and a scrolled-out image SHALL be re-transmitted when scrolled back
into view. Terminals whose inline images are cell-bound (OSC 1337) SHALL NOT
emit any delete sequence, since the repaint itself erases those images.

#### Scenario: Backing out of an article deletes the photo

- **WHEN** the reader returns from an article showing a native photo to the list view on a kitty-family terminal
- **THEN** the photo's placement is deleted from the screen and does not persist over the list

#### Scenario: Leaving the article frees the terminal image cache

- **WHEN** the reader leaves an article whose native photo was displayed on a kitty-family terminal
- **THEN** the terminal receives a delete-by-id for the photo's stable id using the data-freeing delete form (`d=I`), freeing its cached image data as well as its placement

#### Scenario: Leaving the article on Ghostty also clears all placements

- **WHEN** the reader leaves an article whose native photo was displayed on a Ghostty terminal
- **THEN** the frame that leaves the article view additionally emits a delete-all clear (`d=a`) so no placement survives even if the by-id delete is ignored

#### Scenario: Opening another article does not stack photos

- **WHEN** an article with a native photo is opened after an earlier article's photo was displayed on a kitty-family terminal
- **THEN** the earlier placement has been deleted, so the new photo renders alone without stacking over the old one

#### Scenario: Scrolling the image out of view deletes the placement

- **WHEN** the article is scrolled so the photo's rows leave the visible viewport window on a kitty-family terminal (including landing on the caption, which leaves the caption visible)
- **THEN** the placement is deleted so the photo does not float over the caption or body text, and scrolling the photo back into view re-transmits it

#### Scenario: OSC 1337 images need no delete

- **WHEN** frames are rendered on an OSC 1337 terminal
- **THEN** no kitty delete sequence is emitted, because cell-bound inline images are erased by the frame repaint
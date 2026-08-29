## MODIFIED Requirements

### Requirement: Mouse navigation

The system SHALL enable mouse cell-motion reporting and handle mouse events in
both the list and article views. In the list view, a single left-click on an
article row SHALL set the cursor to that row and open the article. A left-click
on a day-header row SHALL toggle that day group's collapse/expand state. A click
on the status bar (the bottom line) SHALL be ignored. The mouse wheel SHALL move
the list cursor: wheel down moves the cursor down by one row, wheel up moves it
up by one row. In the article view, the mouse wheel SHALL scroll the article
viewport: wheel down scrolls down by one line, wheel up scrolls up by one line.
In the article view, an unmodified left-button press on the content area SHALL
begin a text selection (see "In-view text selection" below); a left-button press
carrying a Ctrl or Cmd modifier SHALL NOT be interpreted by the application, so
the terminal can handle clickable OSC 8 hyperlinks natively.

#### Scenario: Click opens article from list

- **WHEN** the user single-clicks on an article row in the list view
- **THEN** the cursor moves to that row and the article reader view opens displaying that article

#### Scenario: Click on day header toggles collapse

- **WHEN** the user single-clicks on a day-header row in the list view
- **THEN** that day group toggles between collapsed and expanded

#### Scenario: Click on status bar is ignored

- **WHEN** the user clicks on the bottom line of the list view (the status bar)
- **THEN** nothing happens

#### Scenario: Wheel scrolls list cursor

- **WHEN** the user scrolls the mouse wheel down in the list view
- **THEN** the list cursor moves down by one row

- **WHEN** the user scrolls the mouse wheel up in the list view
- **THEN** the list cursor moves up by one row

#### Scenario: Wheel scrolls article viewport

- **WHEN** the user scrolls the mouse wheel down in the article view
- **THEN** the article viewport scrolls down by one line

- **WHEN** the user scrolls the mouse wheel up in the article view
- **THEN** the article viewport scrolls up by one line

#### Scenario: Modifier-click in article is not interpreted

- **WHEN** the user Cmd/Ctrl-clicks inside the article content area
- **THEN** no text selection begins and the application does not handle the click

### Requirement: Quit and utility keys

The system SHALL quit on `q` or Ctrl-C from any view. The system SHALL provide `m` to toggle read/unread status, `a` to mark all articles in the current view as read, `o` to open the article URL in the system browser, `c` to copy the article URL to the clipboard, `C` to copy the article text to the clipboard, and `?` to display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in any view
- **THEN** the application exits

#### Scenario: Toggle read status

- **WHEN** the user presses `m` on an article in the list view or article view
- **THEN** the article's read status is toggled between read and unread

#### Scenario: Mark all as read

- **WHEN** the user presses `a` in the list view
- **THEN** all articles currently displayed are marked as read

#### Scenario: Copy article text to clipboard

- **WHEN** the user presses `C` in the article view
- **THEN** the full article text is copied to the system clipboard

## ADDED Requirements

### Requirement: In-view text selection

The system SHALL let the user select article text with the mouse in the article
view and automatically copy the selected text to the system clipboard on
release. An unmodified left-button press on the article content area SHALL
anchor the selection at that cell. While the button is held, dragging SHALL
extend the selection to the current cell. Releasing the button SHALL finalize
the selection, copy the selected text to the system clipboard, and leave the
selection highlighted in the rendered view. The selection SHALL span the
rendered viewport cells, so it tracks the article's current scroll position and
wrapping. The system SHALL render the selected cells with an inverted
(highlighted) style distinct from unselected text. Copied text SHALL be the
plain text characters within the selected region, preserving the line breaks of
the wrapped rendering, with all ANSI styling and OSC 8 hyperlink escape
sequences stripped. A press with an empty or zero-width selection SHALL copy
nothing. Starting a new selection SHALL replace the previous selection. Leaving
the article view SHALL clear the selection.

#### Scenario: Press anchors selection

- **WHEN** the user presses the left button on a cell in the article content area
- **THEN** a selection is anchored at that cell with no highlight yet

#### Scenario: Drag extends selection

- **WHEN** the user holds the left button and drags across cells in the article content area
- **THEN** the selection spans from the anchor cell to the current cell and those cells are rendered inverted

#### Scenario: Release copies selection

- **WHEN** the user releases the left button after dragging a selection
- **THEN** the selected text is copied to the system clipboard and remains highlighted

#### Scenario: Selected text is plain with escapes stripped

- **WHEN** the selected region includes text rendered with ANSI styling or OSC 8 hyperlink sequences
- **THEN** the copied text contains the visible characters with no ANSI or OSC 8 escape bytes

#### Scenario: Selection tracks scroll

- **WHEN** the user drags a selection over a scrollable article
- **THEN** the highlighted cells correspond to the visible viewport content at the current scroll offset

#### Scenario: New selection replaces previous

- **WHEN** the user starts a second drag selection after a first selection exists
- **THEN** the second selection replaces the first and the first's highlight is cleared as the drag proceeds

#### Scenario: Zero-width selection copies nothing

- **WHEN** the user presses and releases the left button without moving
- **THEN** nothing is copied to the clipboard

#### Scenario: Leaving the article clears the selection

- **WHEN** the user returns from the article view to the list view with a selection active
- **THEN** the selection and its highlight are cleared
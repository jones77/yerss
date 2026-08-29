## Purpose

Provides the terminal user interface for browsing, reading, and filtering RSS articles — comprising the article list view, the article reader view with a scroll-progress border, and the tag popup for category-based filtering.

## ADDED Requirements

### Requirement: Article list rendering

The system SHALL display a list of articles where unread articles are rendered in bold text and read articles in plain text. The list SHALL wrap around when moving selection past the top or bottom entry.

#### Scenario: Unread article appears bold

- **WHEN** the article list is displayed and an article has read status false
- **THEN** that article's title is rendered in bold text

#### Scenario: Read article appears plain

- **WHEN** the article list is displayed and an article has read status true
- **THEN** that article's title is rendered in plain (non-bold) text

#### Scenario: Selection wraps from bottom to top

- **WHEN** the user presses down arrow or `j` while the last article is selected
- **THEN** selection moves to the first article in the list

#### Scenario: Selection wraps from top to bottom

- **WHEN** the user presses up arrow or `k` while the first article is selected
- **THEN** selection moves to the last article in the list

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view showing the total number of articles and the date the feeds were last refreshed. When a tag filter is active, the status bar SHALL also display the active filter name.

#### Scenario: Status bar shows count and refresh date

- **WHEN** the list view is displayed
- **THEN** the status bar shows the article count and last-refreshed date

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the count and date

### Requirement: Article reader border rendering

The system SHALL render the article view inside a thin-line border. The article's date and title SHALL appear inline with the top border. A double-line border on the right side SHALL visually represent scroll progress: the filled portion (top N%) uses an accent style and the unfilled portion uses a dim style, where N is the scroll percentage. A percent-scrolled indicator SHALL appear inline with the bottom border. The content area SHALL have two spaces of horizontal padding on each side and one space of vertical padding at the top and bottom.

#### Scenario: Short article fully visible

- **WHEN** an article is opened that fits entirely within the viewport
- **THEN** the top border displays the date and title, the right border is fully filled with the accent style, and the bottom border displays "100% scrolled"

#### Scenario: Long article partially scrolled

- **WHEN** an article is opened and the user has scrolled to 42% of its content
- **THEN** the top 42% of the right double-line border is rendered in accent style, the remaining 58% in dim style, and the bottom border displays "42% scrolled"

#### Scenario: ASCII fallback border

- **WHEN** ASCII fallback mode is enabled or the terminal lacks box-drawing support
- **THEN** single-line border glyphs are replaced with ASCII equivalents (`-`, `|`, `+`) and the double-line right border uses `#` for filled and `:` for unfilled segments

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation (PgUp/PgDn, Ctrl-F/Ctrl-B) SHALL scroll by a full viewport page. Half-page navigation (Ctrl-D/Ctrl-U, Space) SHALL scroll by half a viewport. `g` or Ctrl-Up SHALL move to the top; `G` or Ctrl-Down SHALL move to the bottom. In the list view, up/down moves article selection; in the article view, up/down scrolls the article by one line. `Enter` or `l` opens the selected article from the list; `Esc`, `Enter`, or `h` returns from the article view to the list, reselecting the previously viewed article. In the list view, `Esc` and `h` clear an active tag filter or are no-ops when no filter is active.

#### Scenario: Open article from list

- **WHEN** the user presses Enter or `l` on a selected article in the list view
- **THEN** the article reader view opens displaying that article

#### Scenario: Return to list from article

- **WHEN** the user presses Esc, Enter, or `h` in the article view
- **THEN** the list view is displayed with the previously viewed article selected

#### Scenario: Clear tag filter from list

- **WHEN** a tag filter is active on the list view and the user presses Esc, `h`, or left arrow
- **THEN** the tag filter is removed and all articles are shown

#### Scenario: Esc on unfiltered list is a no-op

- **WHEN** no tag filter is active on the list view and the user presses Esc or `h`
- **THEN** nothing happens

### Requirement: Read-status on open and scroll

The system SHALL mark an article as read when it is opened and the entire article content fits within the viewport. If the article does not fit entirely within the viewport, the system SHALL NOT mark it read upon opening; instead, any single downward scroll action (down arrow, `j`, PgDn, Ctrl-F, Ctrl-D, Space) SHALL mark the article as read instantly.

#### Scenario: Short article marked read on open

- **WHEN** the user opens an article whose full content fits in the viewport
- **THEN** the article's read status is set to true immediately

#### Scenario: Long article not marked read on open

- **WHEN** the user opens an article whose content exceeds the viewport height
- **THEN** the article's read status remains false

#### Scenario: Long article marked read on first downward scroll

- **WHEN** a partially viewed article is displayed and the user presses down arrow, `j`, PgDn, Ctrl-F, Ctrl-D, or Space
- **THEN** the article's read status is set to true immediately

### Requirement: Tag popup on list view

The system SHALL open a tag popup occupying 80% of the screen when the user presses `T` on the list view. The popup SHALL list all tags sorted by popularity (total article count descending). Each tag SHALL display its unread/total counts in parentheses. When read count is zero (all unread) or unread count is zero (all read), the counts SHALL collapse to a single total number. The unread count SHALL always be rendered in bold. Navigation within the popup SHALL use arrow keys and hjkl and SHALL wrap around.

#### Scenario: Open tag popup from list

- **WHEN** the user presses `T` on the list view
- **THEN** a popup covering 80% of the screen appears listing all tags sorted by popularity

#### Scenario: All-unread tag collapses to total

- **WHEN** a tag has 0 read articles and 60 unread articles
- **THEN** the tag displays as `tagname (60)` with 60 in bold

#### Scenario: All-read tag collapses to total

- **WHEN** a tag has 60 read articles and 0 unread articles
- **THEN** the tag displays as `tagname (60)` with 60 in plain text

#### Scenario: Mixed tag shows unread/total

- **WHEN** a tag has 45 read articles and 15 unread articles out of 60 total
- **THEN** the tag displays as `tagname (15/60)` with 15 in bold and 60 in plain text

#### Scenario: Selecting a tag filters the list

- **WHEN** the user selects a tag in the popup and confirms
- **THEN** the list view closes the popup and displays only articles with that tag

### Requirement: Tag popup on article view

The system SHALL open a tag popup occupying 40% of the screen when the user presses `T` on the article view. The popup SHALL list only the categories assigned to the current article, sorted by popularity. Count display and navigation SHALL follow the same rules as the list-view tag popup. Selecting a tag SHALL filter the list view and navigate back to it.

#### Scenario: Open tag popup from article

- **WHEN** the user presses `T` on the article view
- **THEN** a popup covering 40% of the screen appears listing only the current article's tags

#### Scenario: Selecting a tag from article popup filters list

- **WHEN** the user selects a tag in the article-view popup
- **THEN** the system returns to the list view filtered by that tag

### Requirement: Quit and utility keys

The system SHALL quit on `q` or Ctrl-C from any view. The system SHALL provide `m` to toggle read/unread status, `a` to mark all articles in the current view as read, `o` to open the article URL in the system browser, `c` to copy the article URL to the clipboard, and `?` to display a help popup showing the current keymap.

#### Scenario: Quit from any view

- **WHEN** the user presses `q` or Ctrl-C in any view
- **THEN** the application exits

#### Scenario: Toggle read status

- **WHEN** the user presses `m` on an article in the list view or article view
- **THEN** the article's read status is toggled between read and unread

#### Scenario: Mark all as read

- **WHEN** the user presses `a` in the list view
- **THEN** all articles currently displayed are marked as read

### Requirement: Article content rendering

The system SHALL render article HTML content as plain text using an HTML-to-text conversion that preserves paragraph breaks and link URLs. Images SHALL be rendered as `[alt]` using the image's alt text, or `[image]` if no alt text is available.

#### Scenario: HTML content converted to plain text

- **WHEN** an article with HTML content is displayed in the reader view
- **THEN** the content is rendered as plain text with paragraph breaks and link URLs preserved

#### Scenario: Image rendered as alt text

- **WHEN** an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

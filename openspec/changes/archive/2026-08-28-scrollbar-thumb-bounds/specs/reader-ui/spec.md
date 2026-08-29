## MODIFIED Requirements

### Requirement: Article reader border rendering

The system SHALL render the article view inside a thin-line border. The article's
date and title SHALL appear inline with the top border. A double-line border on
the right side SHALL act as a scrollbar: a contiguous thumb segment rendered in
an accent style represents the currently visible portion of the article and is
positioned along the track to reflect the current scroll offset, while the rest
of the track is rendered in a dim style. The thumb SHALL be present from the
first frame when a scrollable article is opened, sitting at the top of the track
when the scroll offset is zero. The thumb height SHALL be proportional to the
fraction of the article that is visible, clamped to a minimum of one row and a
maximum of one row less than the track height (the right border's interior
height between the top and bottom borders). The thumb SHALL never fill the entire
track on a scrollable article, so it always reads as a movable indicator. A
percent-scrolled indicator SHALL appear inline with the bottom border. The
content area SHALL have two spaces of horizontal padding on each side and one
space of vertical padding at the top and bottom. When the entire article fits
within the viewport and no scrolling is possible, the right border SHALL be
fully filled with the accent style and the bottom border SHALL display "100%
scrolled".

#### Scenario: Short article fully visible

- **WHEN** an article is opened that fits entirely within the viewport
- **THEN** the top border displays the date and title, the right border is fully filled with the accent style, and the bottom border displays "100% scrolled"

#### Scenario: Scrollbar visible on open

- **WHEN** a scrollable article (content taller than the viewport) is opened and the scroll offset is zero
- **THEN** the right border renders a thumb at least one row tall in the accent style at the top of the track, with the remainder of the track in the dim style, and the bottom border displays "0% scrolled"

#### Scenario: Thumb moves with scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to 42% of the scrollable range
- **THEN** the thumb's top row is at 42% of the thumb's travel range along the track (rounded), rendered in the accent style, with the rest of the track in the dim style, and the bottom border displays "42% scrolled"

#### Scenario: Thumb height stays within bounds

- **WHEN** a scrollable article is rendered with any combination of total content height and viewport height
- **THEN** the thumb height is at least one row and at most one row less than the right border's track height

#### Scenario: Thumb reaches the bottom at full scroll

- **WHEN** a scrollable article is displayed and the user has scrolled to the very bottom (100% of the scrollable range)
- **THEN** the thumb sits at the bottom of the track in the accent style, with the rest of the track in the dim style, and the bottom border displays "100% scrolled"

#### Scenario: ASCII fallback border

- **WHEN** ASCII fallback mode is enabled or the terminal lacks box-drawing support
- **THEN** single-line border glyphs are replaced with ASCII equivalents (`-`, `|`, `+`) and the double-line right border uses `#` for the thumb (filled) segments and `:` for the unfilled track segments

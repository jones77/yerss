## MODIFIED Requirements

### Requirement: Link mode toggle

The system SHALL provide an `x` key binding in the article view that toggles
between compact and expanded link display. Pressing `x` SHALL immediately
re-render the open article's links in the other mode without re-scrolling the
reader. The chosen mode SHALL persist for the rest of the session across article
switches, starting in compact mode.

The toggle SHALL preserve the reading position: the content line that was the
first visible line in the viewport before the toggle SHALL remain the first
visible line after the toggle, even though compact and expanded modes occupy
different numbers of lines for the same link. The system SHALL anchor on the
first visible content line's text when re-finding its position after the
re-render. When the preserved line can no longer be positioned at the top of the
viewport because the re-rendered article is shorter than the toggle found it
(e.g. its line now sits below the new total line count), the viewport SHALL stay
scrolled to the bottom, and a toggle at a boundary against the article's bottom
edge SHALL keep the article's end pinned to the bottom of the viewport.

#### Scenario: Toggle reveals and hides URLs

- **WHEN** the user presses `x` in the article reader while in compact mode
- **THEN** the article re-renders in expanded mode showing the URL text, and the reader stays at the same scroll position

#### Scenario: Toggle back to compact

- **WHEN** the user presses `x` again while in expanded mode
- **THEN** the article re-renders in compact mode hiding the URL text

#### Scenario: Mode persists across articles

- **WHEN** the user toggles to expanded mode, closes the article, and opens another article
- **THEN** the new article renders in expanded mode

#### Scenario: Default state is compact

- **WHEN** the application starts and the user opens an article without pressing `x`
- **THEN** the article renders in compact mode

#### Scenario: Toggle keeps the first visible line anchored

- **WHEN** the user has scrolled so the first visible content line is a link line (or any line) partway down the article and presses `x`
- **THEN** the same content line is the first visible line after the toggle, instead of the reader jumping to whatever offset the total line-count change implies

#### Scenario: Toggle keeps the bottom pinned when the article shrinks

- **WHEN** the reader is at the bottom of the article and the user presses `x` in a way that shrinks the article's total line count
- **THEN** the article's end remains pinned to the bottom of the viewport and the viewport does not show blank space above the article's end
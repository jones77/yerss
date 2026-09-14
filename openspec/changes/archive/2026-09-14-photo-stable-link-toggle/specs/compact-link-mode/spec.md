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
re-render. When the first visible content line of the viewport lies inside an
image block — a photo row or one of that photo's attribution lines — the system
SHALL instead anchor on that image block: after the re-render it SHALL
re-position the viewport so the same row of the same photo lies at the start of
the viewport, so the photo does not leave the screen. When the preserved line
can no longer be positioned at the top of the viewport because the re-rendered
article is shorter than the toggle found it (e.g. its line now sits below the
new total line count), the viewport SHALL stay scrolled to the bottom, and a
toggle at a boundary against the article's bottom edge SHALL keep the article's
end pinned to the bottom of the viewport.

Toggling SHALL NOT change the size of rendered photos: a photo block's image
height SHALL be the same in compact and expanded mode for the same article,
content width, and viewport, because the height fit is anchored to a fixed
per-article header budget rather than the mode-dependent rendered header. The
toggle SHALL NOT re-trigger an image load, decode, or re-render of a photo
whose block is already composed at the current width.

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

#### Scenario: Toggle keeps a visible photo on screen

- **WHEN** the user has scrolled so the first visible content line of the viewport lies inside an image block — a photo row or the photo's attribution — and presses `x`
- **THEN** the viewport is re-positioned so the same row of the same photo lies at the start of the viewport after the toggle, and the photo remains visible

#### Scenario: Toggle does not re-size a photo

- **WHEN** the user has scrolled so a photo is visible and presses `x`, and the photo's block re-fits against the header budget (rather than being served width-matched from cache)
- **THEN** the photo renders at the same image height in the other mode, with only the link text — not the photo — changing size
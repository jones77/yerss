## MODIFIED Requirements

### Requirement: Link mode toggle

The system SHALL provide an `x` key binding in the article view that toggles
between compact and expanded link display. Pressing `x` SHALL immediately
re-render the open article's links in the other mode without re-scrolling the
reader. The chosen mode SHALL persist across article switches within a session
and SHALL be saved to the application's SQLite `state` table when the app quits,
so a restart reopens in the same mode instead of falling back to compact. A
first start with no saved mode SHALL begin in compact mode.

The toggle SHALL preserve the reading position: the content line that was the
first visible line in the viewport before the toggle SHALL remain the first
visible line after the toggle, even though compact and expanded modes occupy
different numbers of lines for the same link. The system SHALL anchor on the
first visible content line's text when re-finding its position after the
re-render. When the first visible content line is blank (a paragraph separator)
or a line whose visible text is a bare URL — neither of which survives a mode
change — the system SHALL instead anchor on the first non-blank, non-URL text
line at or below that line, using that line's leading word(s), and SHALL keep the
skipped blank separator in place at the top of the viewport. When text reflows so
the anchored line changes position after its first line, the new first visible
line SHALL contain the first word from the other mode, so the reader can
re-orient. When the first visible content line of the viewport lies inside an
image block — a photo row or one of that photo's attribution lines — the system
SHALL instead anchor on that image block: after the re-render it SHALL
re-position the viewport so the same row of the same photo lies at the start of
the viewport, so the photo does not leave the screen. When the preserved line can
no longer be positioned at the top of the viewport because the re-rendered
article is shorter than the toggle found it (e.g. its line now sits below the new
total line count), the viewport SHALL stay scrolled to the bottom, and a toggle
at a boundary against the article's bottom edge SHALL keep the article's end
pinned to the bottom of the viewport. When the anchor is blank or a bare URL and
the re-anchored position lands at or near the article's end, the viewport SHALL
stay glued to the bottom.

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

#### Scenario: Mode persists across restarts

- **WHEN** the user toggles to expanded mode, quits the app, and starts it again
- **THEN** the first article opened renders in expanded mode, without the user pressing `x`

#### Scenario: Default state is compact on first start

- **WHEN** the application starts with no saved link-expand mode in its `state` table and the user opens an article without pressing `x`
- **THEN** the article renders in compact mode

#### Scenario: Toggle keeps the first visible line anchored

- **WHEN** the user has scrolled so the first visible content line is a link line (or any line) partway down the article and presses `x`
- **THEN** the same content line is the first visible line after the toggle, instead of the reader jumping to whatever offset the total line-count change implies

#### Scenario: Toggle keeps the bottom pinned when the article shrinks

- **WHEN** the reader is at the bottom of the article and the user presses `x` in a way that shrinks the article's total line count
- **THEN** the article's end remains pinned to the bottom of the viewport and the viewport does not show blank space above the article's end

#### Scenario: Toggle keeps a visible photo on screen

- **WHEN** the user has scrolled so the first visible content line of the viewport lies inside an image block — a photo row or the photo's attribution — and presses `x`
- **THEN** the viewport is re-positioned so the same row of the same photo lies at the start of the viewport after the toggle, and the photo remains visible

#### Scenario: Toggle does not re-size a photo

- **WHEN** the user has scrolled so a photo is visible and presses `x`, and the photo's block re-fits against the header budget (rather than being served width-matched from cache)
- **THEN** the photo renders at the same image height in the other mode, with only the link text — not the photo — changing size

#### Scenario: Toggle keeps a blank separator line anchored

- **WHEN** the user has scrolled so the first visible content line is a blank line separating paragraphs and presses `x`
- **THEN** the same blank line stays the first visible line after the toggle, and the paragraph it precedes does not jump away

#### Scenario: Toggle keeps bottom pinned when the anchor is a URL line

- **WHEN** the reader is in expanded mode with a bare URL line as the first visible line at or near the article's end and the user presses `x`
- **THEN** the viewport stays glued to the bottom of the article instead of re-anchoring on a confusing mid-article line

#### Scenario: Toggle re-anchors on the first word when text reflows

- **WHEN** the first visible content line's text reflows across modes so the line's later words move, and the user presses `x`
- **THEN** the new first visible line contains the first word of the pre-toggle first visible line

## ADDED Requirements

### Requirement: Expanded header shows the article URL above the title

In expanded link mode the reader's header SHALL render the article's URL as its
first line, then a single blank line, then the title line that compact mode shows
first. In compact mode the URL line SHALL be omitted entirely so the title is the
first header line. The URL line SHALL be clickable (an OSC 8 hyperlink) and SHALL
show the article's own link target, not a per-body-link URL. The one-blank-line
separation SHALL be a single vertical space between the URL and the title.

#### Scenario: Expanded header shows URL then title

- **WHEN** the reader is in expanded mode and an article with a link is open
- **THEN** the header renders the article's URL as its top line, followed by one blank line, then the title line

#### Scenario: Compact header omits the URL line

- **WHEN** the reader is in compact mode and an article with a link is open
- **THEN** the header's first line is the article title, with no URL line above it

#### Scenario: Restored expanded mode renders the URL-above-title header

- **WHEN** the app restarts in expanded mode (the mode was persisted in the `state` table) and the user opens an article
- **THEN** the article's header renders with the URL line above the title, with one blank line between them

### Requirement: Link-expand mode is persisted as runtime state

The system SHALL save the current compact/expanded choice to the SQLite `state`
table when the app quits, in a key independent of the config file, and SHALL load
it at startup before the first article composes so the saved mode drives the
rendered header. The saved value SHALL be a runtime state: toggling `x` SHALL
never write to the config file. When no saved value exists the app SHALL start in
compact mode.

#### Scenario: Quit saves the mode

- **WHEN** the user toggles to expanded mode and quits the app
- **THEN** the `state` table records expanded mode

#### Scenario: Restart restores the saved mode

- **WHEN** the app starts and the `state` table records expanded mode
- **THEN** the reader operates in expanded mode for the whole session, including the first article opened, without pressing `x`

#### Scenario: Toggling does not touch the config file

- **WHEN** the user presses `x` to change mode and quits
- **THEN** the mode is recorded only in the `state` table and no config file is written or modified

#### Scenario: No saved mode means compact

- **WHEN** the app starts with an empty `state` table entry for the mode
- **THEN** the reader operates in compact mode
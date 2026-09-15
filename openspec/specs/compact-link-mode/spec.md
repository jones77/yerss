## Purpose

Gives the article reader compact and expanded link display modes, so readers can
see just the clickable link name by default and reveal the full URLs on demand.

## Requirements

### Requirement: Compact link display by default

The article reader SHALL render body hyperlinks in compact mode by default: the
styled link name only, without the visible URL text. The link name SHALL render
in the status-bar blue role and SHALL be wrapped in OSC 8 escape sequences
carrying the link's URL, so the link remains clickable in both a supporting
terminal (Cmd/Ctrl-click) and through the application's left-click hit-testing.
The visible URL text SHALL NOT be shown in compact mode.

#### Scenario: Body link shows only the name

- **WHEN** an article body contains `<a href="https://example.com/post">world</a>` and the reader is in compact mode
- **THEN** the rendered text shows only `world` in the status-bar blue role, without `https://example.com/post`, and the name is wrapped in OSC 8 sequences carrying the URL

#### Scenario: Compact link is still clickable

- **WHEN** the user Cmd/Ctrl-clicks or left-clicks the link name in compact mode
- **THEN** the URL `https://example.com/post` opens in the system browser

#### Scenario: Compact links work in ASCII mode

- **WHEN** ASCII fallback mode is enabled and the reader is in compact mode
- **THEN** link names render without the URL text and remain OSC 8 clickable

### Requirement: Expanded link display mode

The system SHALL provide an expanded link display mode in the article reader in
which each body hyperlink renders the styled link name followed by the full URL
(`text url`), preserving the visible URL for readers who want to inspect link
targets. The URL text SHALL be wrapped in OSC 8 sequences carrying the same URL
as the name. Expanded mode SHALL NOT alter clickability: both the name and the
URL open the link.

#### Scenario: Expanded link shows name and URL

- **WHEN** the reader is in expanded mode and an article body contains `<a href="https://example.com/post">world</a>`
- **THEN** the rendered text shows `world https://example.com/post` and both are clickable

#### Scenario: URL shown is the link target

- **WHEN** the reader is in expanded mode and a link renders with its URL text
- **THEN** the displayed URL text matches the link's actual target URL

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
### Requirement: Wrapped link URLs are fully hidden in compact mode

The system SHALL hide the entire visible URL of a body hyperlink in compact mode
even when the URL is longer than the content width and wraps across multiple
rendered lines. When a link's URL wraps, the compact rendering SHALL show only
the link's name and SHALL NOT render any part of the URL text — neither the first
line nor any wrapped continuation line. A link name that itself wraps across
lines SHALL render in full. The link name SHALL remain wrapped in OSC 8 escape
sequences carrying the full URL, so the link stays clickable. Hiding a wrapped
URL SHALL NOT leave a short, unpadded line in the surrounding paragraph: the
paragraph SHALL re-wrap to the content width so every line fills the full width
and no visible gap appears where the URL text was removed.

#### Scenario: Long URL wrapping across lines is fully hidden

- **WHEN** an article body contains a link whose URL is longer than the content width so the URL wraps across multiple rendered lines, and the reader is in compact mode
- **THEN** the rendered text shows only the link name, and no part of the URL text — including the wrapped continuation lines — is visible

#### Scenario: Wrapped link name is preserved

- **WHEN** a compact-mode link's name is long enough to wrap across multiple rendered lines
- **THEN** the full link name renders and the URL remains hidden

#### Scenario: Adjacent links to the same URL each hide their href

- **WHEN** an article body contains two links to the same URL in compact mode
- **THEN** both link names render and neither URL text is visible

#### Scenario: Paragraph re-wraps after a wrapped URL is hidden

- **WHEN** a compact-mode link's URL is long enough to wrap across lines, so hiding the URL would otherwise leave the surrounding paragraph line shorter than the content width
- **THEN** the paragraph re-wraps to the full content width and no line is left visibly short with a trailing blank gap

### Requirement: Stripped link re-wrap preserves single word spacing

The system SHALL re-wrap a compact-mode paragraph that contained a hidden URL so
that word separation is a single space: the leading separator space left where a
URL was removed SHALL NOT combine with the re-wrap join space to produce a double
space. A rendered line SHALL NOT contain a doubled space where a link name (or its
punctuation) meets the following text.

#### Scenario: No double space after a stripped link name

- **WHEN** a compact-mode paragraph renders a link name immediately followed by more text whose URL was hidden
- **THEN** the rendered text shows exactly one space between the link name and the following text, with no doubled space

#### Scenario: No double space after stripped link punctuation

- **WHEN** a compact-mode link whose name ends in punctuation is hidden and the paragraph re-wraps
- **THEN** no extra space appears between the punctuation and the following text

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

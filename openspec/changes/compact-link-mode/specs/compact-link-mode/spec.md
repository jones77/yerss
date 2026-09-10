## Purpose

Gives the article reader compact and expanded link display modes, so readers can
see just the clickable link name by default and reveal the full URLs on demand.

## ADDED Requirements

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
reader. The chosen mode SHALL persist for the rest of the session across article
switches, starting in compact mode.

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
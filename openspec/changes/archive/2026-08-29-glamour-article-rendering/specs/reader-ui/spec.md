## MODIFIED Requirements

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown and then render that
markdown to styled terminal text using a theme-aware markdown renderer that
wraps its output to the content width. The renderer SHALL follow the configured
display theme: in `dark` mode it SHALL use the dark theme, in `light` mode the
light theme, and in `auto` mode the theme SHALL be chosen from the terminal's
preferred background. Inline `<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available. `<strong>` and
`<b>` SHALL render with the theme's bold styling and `<em>`/`<i>` with its
italic styling, not as literal markdown markers. Hyperlinks SHALL render as the
styled link text followed by the URL (`text url`) and SHALL be wrapped in OSC 8
hyperlink escape sequences carrying the URL. Structured HTML
SHALL render as styled markdown: ordered and unordered lists (including nested
lists) with correct bullet markers and indentation, blockquotes, code blocks,
and tables. When a link's rendered text does not fit on the current line, the
renderer SHALL wrap it onto a new line. When a URL is too long to fit on a
single line by itself, the renderer SHALL wrap it across lines; it SHALL NOT
truncate the displayed URL. Non-link text SHALL be unaffected.

#### Scenario: HTML content rendered as styled markdown with paragraph breaks

- **WHEN** an article with HTML content is displayed in the reader view
- **THEN** the content is rendered as styled terminal markdown text with paragraph breaks preserved

#### Scenario: Bold text rendered with theme styling

- **WHEN** an article contains `<p><strong>bold text</strong> and <b>more</b> plain</p>`
- **THEN** "bold text" and "more" are rendered with the theme's bold styling and no literal markdown asterisks are displayed

#### Scenario: Bold inside a link stays styled

- **WHEN** an article contains `<a href="https://example.com/x"><strong>bold link</strong></a>`
- **THEN** the link text "bold link" keeps its bold styling inside the rendered link, with no literal markdown asterisks

#### Scenario: Markdown theme follows the display theme

- **WHEN** the configured display theme is `dark` (or `light`, or `auto` resolving to the terminal background)
- **THEN** the article content is rendered with glamour's matching dark (or light) theme

#### Scenario: Nested list renders with correct indentation

- **WHEN** an article contains a `<ul>` with a nested `<ul>` inside one of its `<li>` elements
- **THEN** top-level items render with a bullet marker at the leftmost column and nested items render indented under their parent with their own bullet markers

#### Scenario: List item containing a paragraph renders on one line

- **WHEN** an article contains `<ul><li><p>first point</p></li></ul>` (as emitted by Substack and similar publishers)
- **THEN** the bullet marker and the item text appear on the same line, with no orphaned bullet marker on its own line

#### Scenario: Blockquote rendered as a styled blockquote

- **WHEN** an article contains a `<blockquote>` element
- **THEN** the quoted text is rendered with the theme's blockquote styling and a quote marker

#### Scenario: Image rendered as alt text

- **WHEN** an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

#### Scenario: Link rendered as styled text followed by the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the reader displays "world" styled as a link followed by `https://example.com/post`, with the link wrapped in OSC 8 hyperlink escapes

#### Scenario: Long URL wraps onto a new line

- **WHEN** a link's rendered text does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL wraps across lines

- **WHEN** a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is wrapped across multiple lines and is not truncated or elided

### Requirement: Clickable links in article content

The system SHALL render hyperlinks in article HTML content as clickable OSC 8
terminal hyperlinks. The rendered link SHALL display the styled link text
followed by the URL (`text url`), with the link text wrapped in
OSC 8 escape sequences carrying the link's URL, so that Cmd/Ctrl-clicking it in
a supporting terminal opens the URL in the system browser. The URL SHALL also be
displayed as visible fallback text following the link text, preserving a readable
fallback for terminals that do not support OSC 8. OSC 8 hyperlinks SHALL
function in ASCII fallback mode, since OSC 8 is an escape sequence and not a
Unicode glyph.

#### Scenario: Inline link rendered as clickable styled text with the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the link text "world" is styled and wrapped in OSC 8 escape sequences carrying the URL `https://example.com/post`, and the rendered text shows `world https://example.com/post`

#### Scenario: Cmd/Ctrl-click on link opens browser

- **WHEN** the user Cmd/Ctrl-clicks on the link text of an OSC 8 hyperlink in the article reader view in a supporting terminal
- **THEN** the terminal opens the link URL in the system browser

#### Scenario: Link URL preserved as fallback text

- **WHEN** an article with a link is displayed in a terminal that does not support OSC 8
- **THEN** the link text and the URL (`text url`) are displayed as plain text, identical to the behavior without OSC 8

#### Scenario: OSC 8 links work in ASCII mode

- **WHEN** ASCII fallback mode is enabled and an article with links is displayed
- **THEN** the link text is wrapped in OSC 8 escape sequences and the markdown links are clickable in supporting terminals

### Requirement: Clickable article URL in reader header

The system SHALL render the article's own URL (displayed in the reader view
header) as a clickable OSC 8 terminal hyperlink, so that Cmd/Ctrl-clicking it
opens the URL in the system browser. This mirrors the existing `o` key behavior
for opening the article URL. The header SHALL be rendered through the same
markdown renderer as the article body so its styling matches the configured
display theme.

#### Scenario: Article header URL is clickable

- **WHEN** an article with a link URL is displayed in the reader view and the user Cmd/Ctrl-clicks the URL in the header
- **THEN** the terminal opens the article URL in the system browser
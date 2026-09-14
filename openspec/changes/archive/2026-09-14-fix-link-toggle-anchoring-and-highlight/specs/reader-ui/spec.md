## MODIFIED Requirements

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown and then render that
markdown to styled terminal text using a theme-aware markdown renderer that
wraps its output to the content width. The renderer SHALL follow the configured
display theme: in `dark` mode it SHALL use the dark theme, in `light` mode the
light theme, and in `auto` mode the theme SHALL be chosen from the terminal's
preferred background. The renderer SHALL render base (non-styled) body text in
ANSI color 7 (white) in dark mode and ANSI color 0 (black) in light mode,
overriding the theme's default base foreground, while styled elements (bold,
italics, headings, links, blockquotes) keep the theme's own styling. Inline
`<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available; the alt text
SHALL be rendered literally, so markdown special characters in the alt text
(such as `*`, `_`, `[`, or `]`) SHALL NOT be interpreted as styling or link
syntax. A publisher-attached lead image (from `<enclosure>` or
`media:thumbnail`) is rendered separately as an inline halfblocks image block
above the rendered article content and is not affected by this placeholder rule.
`<strong>` and `<b>` SHALL render with the theme's bold styling and
`<em>`/`<i>` with its italic styling, not as literal markdown markers.
Hyperlink rendering SHALL follow the active link display mode: in expanded mode a
hyperlink SHALL render as the styled link text followed by the URL
(`text url`), while in compact mode (the default) it SHALL render as the styled
link name only with the URL hidden. In both modes the link SHALL be wrapped in
OSC 8 hyperlink escape sequences carrying the URL. Structured HTML SHALL render
as styled markdown: ordered and unordered
lists (including nested lists) with correct bullet markers and indentation,
blockquotes, code blocks, and tables. In expanded mode, when a link's rendered
text does not fit on the current line, the renderer SHALL wrap it onto a new
line, and when a URL is too long to fit on a single line by itself, the renderer
SHALL wrap it across lines; it SHALL NOT truncate the displayed URL. Non-link
text SHALL be unaffected. When HTML-to-markdown conversion fails, the system
SHALL render the article body as plain text with HTML markup removed rather than
displaying the raw HTML source. When markdown rendering fails, the system SHALL
display the article text without markdown formatting markers (no literal `**` or
`[text](url)` syntax) rather than the raw markdown source.

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

#### Scenario: Body text uses the standard white/black base color

- **WHEN** an article is displayed in the reader view in dark (or light) mode
- **THEN** the base body text renders in ANSI color 7 (white) in dark mode and ANSI color 0 (black) in light mode, while bold text, links, and headings keep their theme styling

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

#### Scenario: Image alt text with markdown characters renders literally

- **WHEN** an article contains an `<img>` element with alt text "a *b* [c]"
- **THEN** the reader displays `[a *b* [c]]` literally, with the `*` and `[`/`]` shown as plain characters and no styled or link formatting applied

#### Scenario: Expanded mode renders link as styled text followed by the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view in expanded mode
- **THEN** the reader displays "world" styled as a link followed by `https://example.com/post`, with the link wrapped in OSC 8 hyperlink escapes

#### Scenario: Compact mode renders only the link name

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view in compact mode
- **THEN** the reader displays "world" styled as a link without the URL text, with the link wrapped in OSC 8 hyperlink escapes carrying the URL

#### Scenario: Long URL wraps onto a new line in expanded mode

- **WHEN** in expanded mode a link's rendered text does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL wraps across lines in expanded mode

- **WHEN** in expanded mode a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is wrapped across multiple lines and is not truncated or elided

#### Scenario: Conversion failure renders plain text, not raw HTML

- **WHEN** HTML-to-markdown conversion of the article content fails
- **THEN** the reader displays the article's plain text with HTML markup removed, and no `<p>`/`<a>`/`<div>` tag text is shown

#### Scenario: Markdown render failure shows text without markdown markers

- **WHEN** the markdown renderer fails to render the article content
- **THEN** the reader displays the article text without literal markdown formatting markers such as `**` or `[text](url)`

The system SHALL remove U+00AD soft hyphens from the article content before
rendering: the app counts them as zero-width but the terminal renders each as a
visible cell, so a line padded to the content width would overflow the frame by
one column per soft hyphen and wrap the right border (scrollbar) onto the next
line's left.

### Requirement: In-view text selection

The system SHALL let the user select article text with the mouse in the article
view and automatically copy the selected text to the system clipboard on
release. An unmodified left-button press on the article content area SHALL
anchor the selection at that cell. While the button is held, dragging SHALL
extend the selection to the current cell. Releasing the button SHALL copy the
selected text to the system clipboard and clear the selection highlight. The
selection SHALL span the rendered viewport cells, so it tracks the article's
current scroll position and wrapping. The system SHALL render the selected
cells with an inverted (highlighted) style distinct from unselected text, and
the inverted style SHALL remain continuous across hyperlinks and styled text
within the selection rather than resetting at the first link on a line. Copied
text SHALL be the plain text characters within the selected region, preserving
the line breaks of the wrapped rendering, with all ANSI styling and OSC 8
hyperlink escape sequences stripped. A press with an empty or zero-width
selection SHALL copy nothing. Starting a new selection SHALL replace the
previous selection. Leaving the article view SHALL clear the selection.

#### Scenario: Press anchors selection

- **WHEN** the user presses the left button on a cell in the article content area
- **THEN** a selection is anchored at that cell with no highlight yet

#### Scenario: Drag extends selection

- **WHEN** the user holds the left button and drags across cells in the article content area
- **THEN** the selection spans from the anchor cell to the current cell and those cells are rendered inverted

#### Scenario: Release copies selection

- **WHEN** the user releases the left button after dragging a selection
- **THEN** the selected text is copied to the system clipboard and the selection highlight is cleared

#### Scenario: Selected text is plain with escapes stripped

- **WHEN** the selected region includes text rendered with ANSI styling or OSC 8 hyperlink sequences
- **THEN** the copied text contains the visible characters with no ANSI or OSC 8 escape bytes

#### Scenario: Selection highlights continuously across a link

- **WHEN** the selected region spans a hyperlink and the surrounding text on the same line
- **THEN** the inverted highlight covers the entire selected region including the link, with no un-highlighted gap and no early reset at the link boundary

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

## MODIFIED Requirements

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown and then render that
markdown to styled terminal text using a theme-aware markdown renderer that
wraps its output to the content width. The renderer SHALL follow the configured
display theme: in `dark` mode it SHALL use the dark theme, in `light` mode the
light theme, and in `auto` mode the theme SHALL be chosen from the terminal's
preferred background. Inline `<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available; the alt text
SHALL be rendered literally, so markdown special characters in the alt text
(such as `*`, `_`, `[`, or `]`) SHALL NOT be interpreted as styling or link
syntax. A publisher-attached lead image (from `<enclosure>` or
`media:thumbnail`) is rendered separately as an inline halfblocks image block
above the rendered article content and is not affected by this placeholder rule.
`<strong>` and `<b>` SHALL render with the theme's bold styling and
`<em>`/`<i>` with its italic styling, not as literal markdown markers.
Hyperlinks SHALL render as the styled link text followed by the URL
(`text url`) and SHALL be wrapped in OSC 8 hyperlink escape sequences carrying
the URL. Structured HTML SHALL render as styled markdown: ordered and unordered
lists (including nested lists) with correct bullet markers and indentation,
blockquotes, code blocks, and tables. When a link's rendered text does not fit
on the current line, the renderer SHALL wrap it onto a new line. When a URL is
too long to fit on a single line by itself, the renderer SHALL wrap it across
lines; it SHALL NOT truncate the displayed URL. Non-link text SHALL be
unaffected. When HTML-to-markdown conversion fails, the system SHALL render the
article body as plain text with HTML markup removed rather than displaying the
raw HTML source. When markdown rendering fails, the system SHALL display the
article text without markdown formatting markers (no literal `**` or
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

#### Scenario: Link rendered as styled text followed by the URL

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the reader displays "world" styled as a link followed by `https://example.com/post`, with the link wrapped in OSC 8 hyperlink escapes

#### Scenario: Long URL wraps onto a new line

- **WHEN** a link's rendered text does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL wraps across lines

- **WHEN** a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is wrapped across multiple lines and is not truncated or elided

#### Scenario: Conversion failure renders plain text, not raw HTML

- **WHEN** HTML-to-markdown conversion of the article content fails
- **THEN** the reader displays the article's plain text with HTML markup removed, and no `<p>`/`<a>`/`<div>` tag text is shown

#### Scenario: Markdown render failure shows text without markdown markers

- **WHEN** the markdown renderer fails to render the article content
- **THEN** the reader displays the article text without literal markdown formatting markers such as `**` or `[text](url)`

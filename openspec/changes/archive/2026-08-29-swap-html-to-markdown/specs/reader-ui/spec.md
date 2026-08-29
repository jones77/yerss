## MODIFIED Requirements

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown with embedded terminal
escape sequences for inline formatting, preserving paragraph breaks. Inline
`<img>` elements SHALL be rendered as `[alt]` using the image's alt text, or
`[image]` if no alt text is available. `<strong>` and `<b>` elements SHALL
render as ANSI bold (SGR `\x1b[1m` … `\x1b[22m`) so the text appears bold in
the terminal rather than as literal markdown asterisks. Hyperlinks SHALL be
rendered in markdown style as `[text](url)`: the link text inside the square
brackets wrapped in OSC 8 hyperlink escape sequences carrying the URL, followed
by the URL in parentheses as visible plain-text fallback. Structured HTML SHALL
render as readable markdown: ordered and unordered lists (including nested
lists) with correct bullet markers and indentation, blockquotes with `> `
prefixes, and code blocks with fenced boundaries. When a link's markdown
rendering does not fit on the current line, the system SHALL break the URL onto
a new line. When a URL is too long to fit on a single line by itself, the
system SHALL truncate the displayed URL with an ellipsis glyph (`…` in Unicode
mode, `...` in ASCII fallback mode) so it fits on one line. Non-link text SHALL
be unaffected.

#### Scenario: HTML content rendered as markdown with paragraph breaks

- **WHEN** an article with HTML content is displayed in the reader view
- **THEN** the content is rendered as markdown text with paragraph breaks preserved

#### Scenario: Bold text rendered as terminal bold

- **WHEN** an article contains `<p><strong>bold text</strong> and <b>more</b> plain</p>`
- **THEN** "bold text" and "more" are wrapped in ANSI bold escape sequences and appear bold in the terminal, and no literal markdown asterisks are displayed

#### Scenario: Bold inside a link stays bold

- **WHEN** an article contains `<a href="https://example.com/x"><strong>bold link</strong></a>`
- **THEN** the link text "bold link" is both ANSI bold and wrapped in OSC 8 hyperlink escapes, with no literal markdown asterisks

#### Scenario: Nested list renders with correct indentation

- **WHEN** an article contains a `<ul>` with a nested `<ul>` inside one of its `<li>` elements
- **THEN** top-level items render with a bullet marker at the leftmost column and nested items render indented under their parent with their own bullet markers

#### Scenario: List item containing a paragraph renders on one line

- **WHEN** an article contains `<ul><li><p>first point</p></li></ul>` (as emitted by Substack and similar publishers)
- **THEN** the bullet marker and the item text appear on the same line, with no orphaned bullet marker on its own line

#### Scenario: Blockquote rendered with prefix

- **WHEN** an article contains a `<blockquote>` element
- **THEN** the quoted text is rendered with a `> ` prefix marking it as a blockquote

#### Scenario: Image rendered as alt text

- **WHEN** an article contains an `<img>` element with alt text "photo of a cat"
- **THEN** the reader displays `[photo of a cat]` in place of the image

#### Scenario: Link rendered as markdown

- **WHEN** an article contains `<a href="https://example.com/post">world</a>` and is displayed in the reader view
- **THEN** the reader displays `[world](https://example.com/post)` with the link text wrapped in OSC 8 hyperlink escapes

#### Scenario: Long URL breaks onto a new line

- **WHEN** a link's markdown rendering does not fit on the current wrapped line
- **THEN** the URL portion is placed on a new line

#### Scenario: Overlong URL is truncated with an ellipsis

- **WHEN** a URL is too long to fit on a single display line by itself
- **THEN** the displayed URL is truncated to fit one line and ends with `…` (or `...` in ASCII fallback mode)

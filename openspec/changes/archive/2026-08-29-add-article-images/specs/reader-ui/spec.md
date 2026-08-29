## MODIFIED Requirements

### Requirement: Article content rendering

The system SHALL render article HTML content as markdown and then render that
markdown to styled terminal text using a theme-aware markdown renderer that
wraps its output to the content width. The renderer SHALL follow the configured
display theme: in `dark` mode it SHALL use the dark theme, in `light` mode the
light theme, and in `auto` mode the theme SHALL be chosen from the terminal's
preferred background. Inline `<img>` elements SHALL be rendered as `[alt]` using
the image's alt text, or `[image]` if no alt text is available. A
publisher-attached lead image (from `<enclosure>` or `media:thumbnail`) is
rendered separately as an inline halfblocks image block above the rendered
article content and is not affected by this placeholder rule. `<strong>` and
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

## ADDED Requirements

### Requirement: Article lead image rendering

When image rendering is enabled and an article has a publisher-attached image URL (captured from `<enclosure>` or `media:thumbnail` during feed parsing), the system SHALL render the image inline at the top of the article body, above the rendered article content and below any top padding, as a block of content that scrolls with the article text. The image SHALL be rendered using a Unicode halfblocks representation (half-block characters with ANSI color) sized to fit the content width while preserving aspect ratio. The image block height SHALL be capped so that the full image plus at least one line of body text fits within the viewport height; when the image is taller than this cap, it SHALL be scaled down to fit. When image rendering is disabled (config `off` or `--ascii` mode), no image block SHALL be rendered and the article SHALL render exactly as it does without an image URL.

#### Scenario: Lead image rendered inline above body

- **WHEN** image rendering is enabled and an article with an enclosure image URL is opened
- **THEN** the image is rendered as a halfblocks block at the top of the article body, above the title text, sized to the content width with aspect ratio preserved

#### Scenario: Lead image scrolls with content

- **WHEN** a rendered lead image is visible and the user scrolls down
- **THEN** the image block scrolls upward with the body text as a single scrolling content region

#### Scenario: Lead image capped to viewport

- **WHEN** an article's enclosure image, when aspect-fit to the content width, would be taller than the available viewport height minus one line
- **THEN** the image is scaled down so that the full image and at least one line of body text are visible without scrolling

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block is rendered and the article displays only the title, metadata, and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block is rendered and the article displays only the title, metadata, and body text

### Requirement: Atomic image scroll behavior

When a lead image block is rendered inline, the system SHALL ensure the image is either fully visible within the viewport or fully scrolled out of view; it SHALL never display a partially clipped image. When a downward scroll would move the viewport offset into the interior of the image block (the range strictly between the image's first and last line), the system SHALL snap the offset to the line immediately after the image block, causing the image to vanish in a single keystroke. When an upward scroll would move the offset into the image interior, the system SHALL snap the offset to the image block's first line, revealing the full image. This snapping SHALL apply to all scroll operations: line, page, half-page, and goto-bottom.

#### Scenario: Downward scroll skips past image

- **WHEN** the image block occupies content lines 4 through 9, the viewport offset is 3, and the user scrolls down one line
- **THEN** the viewport offset snaps to 10 (the line after the image) and the image is no longer visible

#### Scenario: Upward scroll reveals full image

- **WHEN** the image block occupies content lines 4 through 9, the viewport offset is 10, and the user scrolls up one line
- **THEN** the viewport offset snaps to 4 (the image's first line) and the full image is visible

#### Scenario: Page-down skips past image

- **WHEN** the image block is partially within the viewport after a page-down operation
- **THEN** the offset snaps past the image block so the image is fully scrolled out

#### Scenario: No snapping when image fully out of view

- **WHEN** the viewport offset is below the image block and the user scrolls further down
- **THEN** no snapping occurs and scrolling behaves normally

### Requirement: Async image load lifecycle

The system SHALL fetch article lead images asynchronously so that opening an article never blocks the UI on a network request. Opening an article with an image URL SHALL first attempt to load the image from the cache hierarchy (in-memory decoded image, then stored database bytes); only on a cache miss SHALL it issue a network fetch command. On a successful load (from cache or network), the system decodes the image and produces a load message that triggers a re-render with the image block inserted; when the load came from a network fetch, the raw image bytes SHALL be persisted to the database. On failure (network error, timeout, non-image content type, or exceeded byte limit), the system SHALL produce a failure message and render the article without an image block. The decoded image SHALL be cached in memory by URL for the session so that resizing the viewport does not re-fetch or re-decode from the database. When the terminal is resized while an article with a loaded image is open, the system SHALL re-render the image block to the new content width using the cached decoded image without re-fetching. When an image loads after the user has scrolled past the insertion point, the system SHALL preserve the user's reading position by adjusting the viewport offset by the number of inserted image-block lines.

#### Scenario: Article opens with image loading asynchronously

- **WHEN** the user opens an article that has an enclosure image URL and image rendering is enabled
- **THEN** the article text renders immediately and the image fetch runs in the background

#### Scenario: Image loads and appears inline

- **WHEN** the background image fetch succeeds while the article is open and the user is at the top of the article
- **THEN** the image block is inserted at the top of the content and becomes visible

#### Scenario: Image fetch fails gracefully

- **WHEN** the background image fetch fails or times out
- **THEN** the article renders without an image block and no error is surfaced to the user

#### Scenario: Resize re-renders image without re-fetch

- **WHEN** the terminal is resized while an article with a loaded image is open
- **THEN** the image block is re-rendered to the new content width using the cached image and no network request is made

#### Scenario: Scroll position preserved on late image load

- **WHEN** an image loads after the user has scrolled into the body text past the image insertion point
- **THEN** the viewport offset is increased by the number of image-block lines so the user's reading position is unchanged

#### Scenario: Re-opened article uses cached image

- **WHEN** the user closes and re-opens an article whose image was previously fetched
- **THEN** the image block renders immediately from the in-memory cache without a network request

#### Scenario: Restarted article uses stored image bytes

- **WHEN** the application is restarted and the user opens an article whose image was previously fetched and stored in the database
- **THEN** the image block renders from the stored bytes without making a network request to the image host

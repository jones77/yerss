## MODIFIED Requirements

### Requirement: Article lead image rendering

When image rendering is enabled and an article has a publisher-attached image
URL (captured from `<enclosure>` or `media:thumbnail` during feed parsing), the
system SHALL render the image inline below the reader header (the URL, title,
and author lines) and above the rendered article content, as a block of content
that scrolls with the article text. On a terminal without native image support
the image SHALL be rendered using a Unicode halfblocks representation
(half-block characters with ANSI color) sized to fit the content width while
preserving aspect ratio, and SHALL be horizontally centered within the content
width. The halfblock block SHALL be served from the stored text block when one
exists at the current content width, re-rendered from the in-memory decoded
image when one is cached, or re-fetched and re-rendered when neither is
available. On a full-image-capable terminal the halfblock block SHALL be shown
first as an immediate placeholder and SHALL be replaced by a native inline
photo render once the real photo loads. Directly beneath the image or photo,
with no blank line between them, the system SHALL render a photo attribution
constrained to a standard caption width — a fixed fraction of the content width
(such as 70%) that does not depend on the photo's own width — wrapped onto
additional centered lines when the credit is longer and centered beneath it; a
blank line SHALL follow the attribution before the article content resumes. The
attribution text SHALL be the photo's credit extracted from the article's
stored HTML when present — a `figcaption` associated with the figure containing
the lead image, used only when the lead image is the figure's last `<img>` in
document order (a figure's caption belongs to its last image), otherwise an
element whose class indicates a credit line — and SHALL fall back to `photo:
<source>`, where `<source>` is derived by the same source-identifier rules used
for the list view's organization column; when the lead image sits inside a
captioned figure whose caption belongs to a later image, no attribution SHALL
render at all, with no `photo: <source>` fallback. The combined
image-plus-attribution block height SHALL be capped so that the full
block plus at least one line of body text fits within the viewport height; the
height fit SHALL reserve the wrapped line count of the caption the block
actually composes with, so a long caption cannot overflow the reserved space and
inflate the block past the viewport; when the block is taller than this cap it
SHALL be scaled down to fit. When image rendering is disabled
(config `off` or `--ascii` mode), no image block and no attribution SHALL be
rendered and the article SHALL render exactly as it does without an image URL.

#### Scenario: Lead image rendered below the header and above the body

- **WHEN** image rendering is enabled and an article with an enclosure image URL is opened
- **THEN** the image is rendered as a halfblocks block positioned below the URL/title/author header lines and above the body text, sized to the content width with aspect ratio preserved

#### Scenario: Image is horizontally centered

- **WHEN** a rendered lead image occupies fewer columns than the content width
- **THEN** the image block is centered with equal left and right padding within the content area

#### Scenario: Attribution centered directly beneath the image

- **WHEN** a lead image is rendered with an attribution
- **THEN** the attribution line is centered beneath the image at the standard caption width and there is no blank line between the image and the attribution

#### Scenario: Attribution wraps at the standard caption width

- **WHEN** the attribution text is longer than the standard caption width (for example a long credit beneath a narrow photo)
- **THEN** the attribution wraps onto additional centered lines within the standard caption width — a fixed fraction of the content width that does not depend on the photo's own width — and the photo is re-rendered smaller so the whole block plus at least one line of body text still fits the viewport

#### Scenario: Attribution rendered at a standard caption width

- **WHEN** a lead image narrower than the content width renders with a long credit
- **THEN** the attribution wraps to the standard caption width (a fixed fraction of the content width), centered beneath the image, rather than wrapping to the photo's own width

#### Scenario: Long caption does not inflate the block past the viewport

- **WHEN** an image's credit is long enough that wrapping it at the standard caption width takes several lines
- **THEN** the image is scaled down so the image plus the full wrapped attribution and one line of body text fit within the viewport height

#### Scenario: Attribution taken from the article HTML

- **WHEN** the article's HTML contains a `figcaption` credit associated with the lead image's figure
- **THEN** the attribution line shows that credit text, centered beneath the photo

#### Scenario: Lead image in a shared captioned figure renders without a caption

- **WHEN** the lead image is not the last `<img>` of a figure whose `figcaption` labels a later image in the same figure
- **THEN** no attribution renders beneath the lead image, with no `photo: <source>` fallback

#### Scenario: Attribution falls back to the source organization

- **WHEN** the article's HTML contains no credit for the lead image and the article link host is `theintercept.com`
- **THEN** the attribution line shows `photo: intercept`

#### Scenario: Blank line after the attribution

- **WHEN** a lead image with attribution is rendered
- **THEN** exactly one blank line separates the attribution from the first body line

#### Scenario: Lead image scrolls with content

- **WHEN** a rendered lead image is visible and the user scrolls down
- **THEN** the image block scrolls upward with the body text as a single scrolling content region

#### Scenario: Lead image capped to viewport

- **WHEN** an article's enclosure image, when aspect-fit to the content width, would make the combined image-plus-attribution block taller than the available viewport height minus one line
- **THEN** the image is scaled down so that the full block and at least one line of body text are visible without scrolling

#### Scenario: Stored block renders on open

- **WHEN** image rendering is enabled and an article with a stored image block at the current content width is opened
- **THEN** the stored block renders immediately as the lead image without a network request

#### Scenario: Native terminal swaps block for the photo

- **WHEN** a full-image-capable terminal opens an article and the fetched photo loads
- **THEN** the halfblock placeholder block is replaced by a native inline photo render with the attribution still beneath

#### Scenario: No image block when rendering disabled

- **WHEN** image rendering is disabled and an article with an enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text

#### Scenario: No image block when URL absent

- **WHEN** image rendering is enabled and an article with no enclosure image URL is opened
- **THEN** no image block and no attribution line are rendered and the article displays only the header and body text
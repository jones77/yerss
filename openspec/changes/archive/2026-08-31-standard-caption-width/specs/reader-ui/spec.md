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
the lead image, otherwise an element whose class indicates a credit line — and
SHALL fall back to `photo: <source>`, where `<source>` is derived by the same
source-identifier rules used for the list view's organization column. The
combined image-plus-attribution block height SHALL be capped so that the full
block plus at least one line of body text fits within the viewport height; the
height fit SHALL reserve the wrapped line count of the caption the block
actually composes with, so a long caption cannot overflow the reserved space
and inflate the block past the viewport; when the block is taller than this cap
it SHALL be scaled down to fit. When image rendering is disabled (config `off`
or `--ascii` mode), no image block and no attribution SHALL be rendered and the
article SHALL render exactly as it does without an image URL.

#### Scenario: Attribution rendered at a standard caption width

- **WHEN** a lead image narrower than the content width renders with a long credit
- **THEN** the attribution wraps to the standard caption width (a fixed fraction of the content width), centered beneath the image, rather than wrapping to the photo's own width

#### Scenario: Long caption does not inflate the block past the viewport

- **WHEN** an image's credit is long enough that wrapping it at the standard caption width takes several lines
- **THEN** the image is scaled down so the image plus the full wrapped attribution and one line of body text fit within the viewport height

#### Scenario: Attribution centered directly beneath the image

- **WHEN** an image renders with a credit
- **THEN** the attribution is centered beneath the image with no blank line between them

## ADDED Requirements

### Requirement: Inline image caption width

When an inline image renders in the article body with an attribution, the
system SHALL render the attribution at the same standard caption width as the
lead image attribution — a fixed fraction of the content width that does not
depend on the inline photo's own width — wrapped onto additional centered lines
when the caption is longer and centered beneath the image. Inline image blocks
SHALL reserve the wrapped line count of the caption they compose with in their
height fit, so an inline block (image plus caption) stays within the viewport
height budget even for a long caption shared by a gallery of narrow images.

#### Scenario: Long shared gallery caption wraps compactly

- **WHEN** an article renders two inline photos that share a long identical caption and each photo is narrower than the content width
- **THEN** each attribution wraps to the standard caption width beneath its own photo, and neither image-plus-caption block exceeds the viewport height budget

#### Scenario: Inline caption centered at the standard width

- **WHEN** an inline image with a short caption renders
- **THEN** the caption is centered at the standard caption width beneath the image, not stretched to the photo's width
# inline-image-rendering Specification

## Purpose

Renders inline `<img>` images from article content in place, interleaved with the
surrounding text, with attribution and the same native/halfblock rendering and
persistence as the lead image.

## Requirements

### Requirement: Inline image discovery

The system SHALL discover inline `<img>` elements in an article's HTML content
and extract their source URLs in document order, so each can be rendered in
place. Discovery SHALL be performed when the article is opened, not at refresh
time. The system SHALL NOT treat inline images as the lead image; the lead image
continues to come from the publisher-attached enclosure or media:thumbnail
metadata.

#### Scenario: Inline images discovered in document order

- **WHEN** an article's content contains two `<img>` elements with sources `a.jpg` and `b.jpg` in that order
- **THEN** the system discovers `a.jpg` followed by `b.jpg` as inline images

#### Scenario: No inline images

- **WHEN** an article's content contains no `<img>` elements
- **THEN** the system discovers no inline images and renders the article as before

### Requirement: Inline image rendering with attribution

The system SHALL render each discovered inline image in place within the article
body — interleaved with the surrounding text — using the same rendering path as
the lead image: a native inline photo on a full-image-capable terminal (with a
halfblock placeholder first), and halfblock art otherwise, sized to the content
width and centered, capped so the block fits the viewport. Each inline image
SHALL have its own attribution centered directly beneath it, taken from the
image's alt text when present, otherwise from the caption or credit of the
figure whose image source matches the image (a src-matched `<figure>`),
otherwise from the text of any markdown link that wraps the image (promo images
like a site's banner render as a linked image whose link text labels them),
otherwise falling back to `photo: <source>` derived by the source-identifier
rules. An image SHALL NOT display another image's caption: there is no
document-global credit fallback — neither the first figure with a caption nor
the first element whose class indicates a credit — for the lead image or any
inline image, so an image that has no caption of its own renders without one
rather than borrowing another image's.

#### Scenario: Inline image rendered in place

- **WHEN** an article has body text before and after an inline image, and the image loads
- **THEN** the image block appears between the surrounding text paragraphs, not at the top of the article

#### Scenario: Inline image attribution

- **WHEN** an inline image renders
- **THEN** a centered attribution appears directly beneath it, using the alt text, the src-matched figure caption, the wrapping link's text, or the `photo: <source>` fallback in that order

#### Scenario: Inline image without its own caption has none

- **WHEN** an inline image has no alt text and is not inside a figure with a caption or credit, while another image elsewhere in the article is
- **THEN** the inline image renders without an attribution rather than displaying the other image's caption

#### Scenario: Linked promo image uses its link text as the caption

- **WHEN** an article's markup wraps an inline image and banner text in a hyperlink (`<a href="/promo"><img src="logo.jpg"><p>Read Our Complete Coverage</p>…</a>`)
- **THEN** the image renders with the banner text ("Read Our Complete Coverage…") centered beneath it as its caption, and no leftover `[`, link text, or `](…)` fragments render around the block

#### Scenario: Inline image loads asynchronously

- **WHEN** an article with an inline image is opened
- **THEN** the article text renders immediately and the inline image loads in the background

### Requirement: Linked inline images render without link fragments

The system SHALL render an inline image that is the content of a hyperlink
(`<a>` wrapping `<img>`) as the image block alone: the markdown link the
converter wraps around the image sentinel SHALL be consumed whole — including
any link text between the sentinel and the destination — so no stray `[`, link
text, or `](destination)` text renders before or after the image block.

#### Scenario: Linked inline image shows no bracket artifacts

- **WHEN** an article contains `<a href="…"><img src="logo.jpg" alt=""></a>` and the image renders
- **THEN** the image block renders in place with no visible `[` or `](…)` fragments around it

#### Scenario: Link text between a linked image and its destination is consumed

- **WHEN** an article contains `<a href="/promo"><img src="logo.jpg"><span>Promo label</span></a>` and the image renders
- **THEN** the whole `[image Promo label](/promo)` construct renders as just the image block, with the label available as the image's caption fallback

#### Scenario: Linked promo image is listed in the links popup

- **WHEN** an article contains a promo banner rendered as a linked image (`<a href="/promo"><img …>…</a>`) and the user opens the links popup
- **THEN** the popup lists the link with its destination resolved to an absolute URL (a relative target like `/promo` against the article's origin) and its link text free of the image sentinel and link-fragment markup

### Requirement: Inline image persistence

The system SHALL persist each inline image's rendered block and full photo bytes
in the article's image table (at positions following the lead image), so a later
open re-renders inline images offline without a network request.

#### Scenario: Inline images persisted

- **WHEN** an inline image is fetched
- **THEN** its rendered block and full photo bytes are stored in the article's image table at a position after the lead image

#### Scenario: Inline images render from storage

- **WHEN** the application is restarted and an article with previously stored inline images is opened
- **THEN** the inline images render from the stored bytes without a network request
## MODIFIED Requirements

### Requirement: Inline image rendering with attribution

The system SHALL render each discovered inline image in place within the article
body — interleaved with the surrounding text — using the same rendering path as
the lead image: a native inline photo on a full-image-capable terminal (with a
halfblock placeholder first), and halfblock art otherwise, sized to the content
width and centered, capped so the block fits the viewport. Each inline image
SHALL have its own attribution centered directly beneath it, taken first from
the caption of the figure that directly contains the image and carries its own
figcaption (so a photo gallery whose photos each sit in their own figure with
their own figcaption renders each figcaption centered beneath its photo), then
from the image's alt text, then from the text of any markdown link that wraps
the image (promo images like a site's banner render as a linked image whose link
text labels them), otherwise falling back to `photo: <source>` derived by the
source-identifier rules. An image's figcaption SHALL be its centered
attribution, and the article body SHALL NOT also render that caption text as a
separate left-aligned paragraph: the caption is not duplicated as body content.
A figure whose single figcaption labels multiple `<img>` elements (a photo-grid)
attributes the caption only to its last image in document order — a figure's
caption belongs to its last image, so an image earlier in a shared captioned
figure has no caption of its own. An image SHALL NOT display another image's
caption: there is no document-global credit fallback — neither the first figure
with a caption nor the first element whose class indicates a credit — for the
lead image or any inline image, so an image that has no caption of its own
renders without one rather than borrowing another image's. When an image sits
inside a captioned figure whose caption belongs to a later image in that figure
(a photo-grid), the image SHALL render with no caption at all: it SHALL NOT
borrow the figure's caption, and it SHALL NOT fall back to `photo: <source>`,
because the fallback applies only to images with no caption source of their own.

#### Scenario: Inline image rendered in place

- **WHEN** an article has body text before and after an inline image, and the image loads
- **THEN** the image block appears between the surrounding text paragraphs, not at the top of the article

#### Scenario: Inline image attribution

- **WHEN** an inline image renders
- **THEN** a centered attribution appears directly beneath it, using the directly-containing figure's caption, the alt text, the wrapping link's text, or the `photo: <source>` fallback in that order

#### Scenario: Gallery photo shows its own figure's caption

- **WHEN** a photo gallery contains photos each inside its own figure with its own figcaption (nested figures within a gallery wrapper, as in a WordPress `wp-block-gallery`), and the article renders
- **THEN** each photo renders its own figcaption centered directly beneath it, and the figcaption text does not also appear as a left-aligned body paragraph elsewhere in the article

#### Scenario: Inline image without its own caption has none

- **WHEN** an inline image has no alt text and is not inside a figure with a caption or credit, while another image elsewhere in the article is
- **THEN** the inline image renders without an attribution rather than displaying the other image's caption

#### Scenario: Photo-grid first image renders without a caption

- **WHEN** an inline image is the first of two `<img>` elements inside a `<figure>` whose `figcaption` labels the second (`<figure><img src="a.jpg"><img src="b.jpg"><figcaption>…</figcaption></figure>`)
- **THEN** the first image (a.jpg) renders with no attribution at all — it does not display the figure's caption and does not fall back to `photo: <source>`

#### Scenario: Photo-grid last image shows the figure's caption

- **WHEN** an inline image is the last of multiple `<img>` elements inside a `<figure>` whose `figcaption` labels it
- **THEN** the last image renders with the figure's caption centered beneath it

#### Scenario: Linked promo image uses its link text as the caption

- **WHEN** an article's markup wraps an inline image and banner text in a hyperlink (`<a href="/promo"><img src="logo.jpg"><p>Read Our Complete Coverage</p>…</a>`)
- **THEN** the image renders with the banner text ("Read Our Complete Coverage…") centered beneath it as its caption, and no leftover `[`, link text, or `](…)` fragments render around the block

#### Scenario: Inline image loads asynchronously

- **WHEN** an article with an inline image is opened
- **THEN** the article text renders immediately and the inline image loads in the background
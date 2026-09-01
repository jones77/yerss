## Why

A photo-grid — one `<figure>` containing several `<img>` elements and a single
`figcaption` — currently copies that caption onto **every** image in the figure:
`figureCredit` matches any figure whose subtree contains the image's URL, so all
grid photos display the same combined caption (e.g. the Intercept's
`Left: … Right: …` label, observed on both photos of the hurricane article).
The caption belongs to the **last** image in the figure; earlier images should
render with no caption at all.

## What Changes

- A figure's `figcaption` is attributed only to the **last** `<img>` in that
  figure (in document order). An `<img>` earlier in a shared figure no longer
  receives the figure's caption.
- An image that sits in a captioned figure whose caption belongs to a later
  image renders with **no caption whatsoever** — it does not borrow the
  figure's caption, and it does not receive the `photo: <source>` fallback
  (that fallback is only for images with no caption source of their own).
- Applies uniformly to the lead image and inline images, since both use the
  same figure-caption extraction.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `inline-image-rendering`: the "Inline image rendering with attribution"
  requirement changes so a src-matched figure's caption belongs only to the
  figure's last image, and an earlier image in a shared captioned figure
  renders with no caption rather than borrowing the figure's or falling back
  to the source label.
- `reader-ui`: the lead-image credit requirement changes the same way — the
  `figcaption` associated with the lead image's figure is used only when the
  lead image is the figure's last image; otherwise no credit is rendered for it
  (no `photo: <source>` fallback for a shared-figure non-owner).

## Impact

- `convert/credit.go`: `figureCredit`/`ImageCredit` — attribute the caption to
  the figure's last `<img>` only; expose whether the image is in a captioned
  figure so callers can distinguish "no figure" from "shared figure, caption
  belongs elsewhere".
- `internal/ui/image.go` (`articleAttribution`) and `internal/ui/article.go`
  (`inlineAttribution`): return no caption for a shared-figure non-owner
  instead of falling back to `photo: <source>`.
- `convert/credit_test.go` and `internal/ui/article_image_test.go`: new cases
  for a multi-image figure (first image → no caption, last image → caption).
- The Intercept hurricane article in the local store (id 75) is the repro: the
  `Swannanoa-Bridgestone-2024` photo must render captionless while the 2026
  photo keeps the combined caption.
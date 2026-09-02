## Why

Drop Site News (and other Substack-style feeds) publish an enclosure image that
duplicates the article's first body image. Because the enclosure URL and the
body `<img src>` are different CDN transform variants of the same source photo,
the reader's exact-URL dedup does not catch them, so the same photo renders
twice — once as the lead block at the top and again as the first inline image —
and its bytes are fetched and stored twice.

## What Changes

- When an article's enclosure (lead) image is the same underlying source photo
  as any inline image in the article body, the reader SHALL NOT render the lead
  image as a top-of-article block; it SHALL render the body's copy in place at
  its natural position in the article flow.
- The reader SHALL decide "same photo" by comparing canonical source URLs, not
  exact fetch URLs and not decoded bytes: a CDN transform prefix is stripped to
  recover the underlying source, so `.../fetch/$s_!X!,f_auto,.../src` and
  `.../fetch/$s_!X!,w_2400,c_limit,.../src` are recognized as the same image.
- The existing inline-image dedup (which currently keeps the lead and drops the
  body copy) is inverted for the duplicated case: the first body copy is kept
  and the lead is suppressed.
- Because the suppressed lead is no longer composed, its image is not fetched
  and persisted at position 0, eliminating the duplicate download and the
  identical duplicate `article_images` rows.

## Capabilities

### New Capabilities

- `lead-image-dedup`: Recognizing when an article's enclosure/lead image is the
  same source photo as a body inline image and suppressing the redundant top
  lead block in favor of the body copy.

### Modified Capabilities

- `inline-image-rendering`: The lead/body dedup currently suppresses inline
  duplicates of the lead; the behavior inverts so that when the lead is
  duplicated in the body, the body copy renders instead of the lead.
- `image-block-storage`: Positions and storage no longer assume the lead image
  always occupies position 0; a suppressed lead stores its bytes only once, at
  the body image's position.

## Impact

- `internal/ui/article.go` — `newArticleState` decides whether to compose the
  top lead block and how to dedup inline images.
- `internal/ui/compose/sentinel.go` — `DedupeInline` / `StripDuplicateSentinels`
  gain canonical-source matching.
- `internal/ui/compose/image.go` and `internal/ui/image.go` — attribution and
  native-render caption resolution treat a suppressed lead URL as an inline
  image.
- A new canonical-source extraction helper (URL normalization) shared by the
  compose package.
- Tests: `internal/ui/article_image_test.go` (`TestInlineImageLeadDuplicateSkipped`
  and friends) and compose tests need updating to the new semantics.

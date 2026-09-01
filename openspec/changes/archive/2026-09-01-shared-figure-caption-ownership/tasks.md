## 1. Figure caption ownership in convert

- [x] 1.1 Change `figureCredit` in `convert/credit.go` to attribute a captioned figure's `figcaption` only to its last `<img>` in document order, and add `ImageCaption(htmlStr, url) (caption string, matched bool)` where `matched` reports whether the image sits in a captioned figure at all
- [x] 1.2 Keep `ImageCredit` as a thin wrapper over `ImageCaption` returning just the caption, so existing callers and tests keep their signatures

## 2. Attribution paths render no caption for shared-figure non-owners

- [x] 2.1 Update `inlineAttribution` in `internal/ui/article.go`: an image in a captioned figure whose caption belongs to a later image (matched but empty) returns "" — bypassing the alt, link-text, and `photo: <source>` fallbacks
- [x] 2.2 Update `articleAttribution` in `internal/ui/image.go` for the lead image the same way (no `photo: <source>` fallback for a shared-figure non-owner)

## 3. Tests and verification

- [x] 3.1 Add convert tests for a multi-image figure: the first image gets no caption, the last image gets the figure's caption, and an image in a captioned figure that is not the owner is reported as matched-but-empty
- [x] 3.2 Add a ui test exercising the grid scenario end-to-end: an article with a two-photo figure and a combined caption composes the first photo with no attribution and the last with the caption
- [x] 3.3 Confirm the existing credit tests pass unmodified, then run the full test suite, the linter, and a build; verify `git status` shows only the intended files changed
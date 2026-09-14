# Proposal: photo-stable-link-toggle

## Why

Toggling compact/expanded link display with `x` re-composes the open article,
and a photo on screen reacts oddly: it sometimes shrinks or grows and sometimes
scrolls off the screen. Toggling a link mode should affect the text — not the
photos next to it. Photo size should also not depend on what the text is doing,
matching what a reader naturally expects.

## What Changes

- Make the image-block height fit mode-independent. The fit caps an image's
  height against a per-article header budget; that budget currently uses the
  link-mode-dependent rendered header's line count, so re-fitting after a
  toggle re-sizes the photo. The budget will be the tallest header the article
  presents across both link modes, fixed for the article, so a toggle never
  re-fits a photo differently.
- Make the link-mode toggle anchor on a photo block. When the first visible
  content line lies inside an image block (photo rows or its caption), the
  toggle will re-position the same photo region at the start of the viewport
  instead of falling back to the total line-count delta (which is only correct
  for top insertions) and letting the photo drift away or the total-delta jump.
- Keep `recomposeArticle`'s image-load recompose path and bottom-pinning
  (`GotoBottom`) semantics unchanged.

## Capabilities

### Modified Capabilities

- **compact-link-mode** — the "Link mode toggle" requirement gains photo
  stability: toggling link display must neither re-size visible photos nor move
  a visible photo off the screen, for photo rows and caption rows alike.

## Impact

- `internal/ui/article.go` — derivation gains a mode-independent header budget
  (`fitHeaderLines`) used wherever the image block height fit is computed.
- `internal/ui/image.go` — `toggleLinkExpand` gains a photo-block anchor step;
  `nativeRenderCmd` and `composeImageBlock`/`fitBlock` use the mode-independent
  budget.
- Tests in `internal/ui/{link_expand_test.go,article_image_test.go}` — new
  toggle-photo tests plus updated expectations where the fit budget changed.
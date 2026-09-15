## Why

Three rough edges degrade the article reading experience. In expanded link
mode, glamour pushes a URL that does not fit next to its link name onto its own
line, so the paragraph reads as a dangling URL fragment and the name line stays
half empty. In compact mode, hiding a wrapped URL can leave a whitespace-only
residual line that splits the surrounding paragraph during re-wrap. And on
restart, the reading position is restored by a raw scroll offset, so image
blocks that have not materialized yet (and the hardcoded 80x24 size the restore
composes at before the real terminal size arrives) land the reader on the wrong
line.

## What Changes

- **Expanded link-mode URL inline reflow**: when glamour pushes an unfittable
  URL onto its own line, the paragraph re-flows so the URL starts on the link
  name's line and wraps at the content width, with each wrapped line remaining
  an OSC 8 clickable hyperlink.
- **Compact-mode wrapped-href residual fix**: hiding a wrapped URL now also
  swallows the trailing fill padding and wrap newline, so no whitespace-only
  line survives to split the paragraph during re-wrap.
- **Content-anchored selection restore**: on quit the app records the reading
  position as a content anchor (the image block URL, or the first text line's
  leading words plus the gap to the viewport top); on start it re-finds that
  anchor in the freshly composed article, and the first real window-size
  message re-applies the anchor at the true terminal size.
- **Below-fold image-load scroll fix**: recomposing after an image load bumps
  the scroll offset only by lines inserted above the viewport top; a block that
  materializes below the fold no longer shifts the reading position.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `compact-link-mode`: expanded mode re-flows pushed-out URLs inline so the URL
  starts on the name's line and wraps within the content width; compact mode
  leaves no whitespace-only residual line after hiding a wrapped URL.
- `article-recomposition`: recompose preserves the reading position by counting
  only lines inserted above the viewport top, so a block landing below the fold
  does not shift the reader.
- `reader-ui`: restore-reader-state-on-start becomes content-anchored — a
  block URL or leading-text anchor with gap — instead of a raw scroll offset,
  so the reopened article lands on the same line even before image blocks
  materialize and at the true terminal size.

## Impact

- `internal/ui/compose/inline.go` (new): `InlineLinkURLs` paragraph re-flow for
  expanded link mode.
- `internal/ui/compose/links.go`: `collectLinkSpans`/`groupLinkUnits` split and
  the wrapped-href drop extending through trailing fill and wrap newline.
- `internal/ui/articleview/article.go`: expanded-link branch applies
  `InlineLinkURLs`.
- `internal/ui/articleview/method.go`: `RecomposeInPlace` insertion accounting,
  plus `AnchorAt`/`RestoreAt`.
- `internal/ui/model.go`: `pendingRestore` re-apply on the first
  `WindowSizeMsg`.
- `internal/ui/selection.go`: `persistSelection` records the anchor;
  `restoreSelection` re-finds it.
- `internal/store/store.go` (already committed): `LastSelection` gains
  `ArticleBlock`, `ArticleAnchor`, `ArticleGap`.
- Tests: `compose/links_test.go`, `internal/ui/selection_test.go`,
  `internal/ui/article_image_test.go`.
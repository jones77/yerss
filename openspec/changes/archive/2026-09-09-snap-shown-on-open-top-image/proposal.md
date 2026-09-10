## Why

A top-of-article photo whose URL is both the feed lead image and the first
inline image in the body (a common ProPublica pattern) renders once, directly
below the header, and is shown on open — but the scroll snap treats it as a
mid-article inline image that entered from below. Scrolling down walks the
header line by line and only skips the photo once its top reaches the viewport
top, and scrolling up walks its caption line by line instead of snapping the
photo flush to the top of the screen.

## What Changes

- `leadShown` is now derived from geometry rather than only from a genuine,
  non-suppressed lead: any block composed directly below the header
  (`ImgStart == headerLines+1`) is shown on open, so its snap entry stages are
  pre-consumed. A suppressed lead whose URL opens the body as its first inline
  image therefore rises flush to the viewport top on the next down press and
  then skips its whole photo-and-caption block past its caption, never
  scrolling past line by line. A suppressed lead reused mid-article (a promo
  banner) still snaps like a regular inline image.
- The upward top-snap now applies to the shown-on-open top image too: as soon
  as the block's last line enters the window from above, the whole photo and
  caption snap flush to the viewport top rather than the caption scrolling up
  line by line.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `reader-ui`: the Atomic image scroll behavior requirement gains the
  shown-on-open top image's down-rise-and-skip transition and its upward
  top-snap, replacing the previous inline-style (suppressed-lead) treatment
  and the line-by-line caption reveal.

## Impact

- `internal/ui/article.go`: `composeArticle` computes `leadShown` from
  `blocks[0].ImgStart` vs `headerLines`.
- `internal/ui/compose/image.go`: `SnapYOffset` applies the upward top-snap to
  every block including the shown-on-open top image.
- `internal/ui/article_image_test.go`: updated snap expectations plus a
  regression test for the shown-on-open suppressed lead.
- `openspec/specs/reader-ui/spec.md`: the Atomic image scroll behavior
  requirement text is updated to describe the new transitions.
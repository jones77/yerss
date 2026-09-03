## Why

The article-padding-snap-flush change made the vertical padding trailing blank
content rows, so the 1-space vertical padding is only visible at the very end
of an article; while reading, the content fills the interior and there is no
vertical margin at all. The intent is instead a permanent, configurable margin
between the content viewport and the borders (`padding_y`, default 1): the
caption's last line snaps to the viewport's last line, with the margin always
sitting between the viewport and the border. The margin applies symmetrically —
one row above the content (below the top border) and one row below it (above
the bottom border). This restores the original permanent-margin layout that the
trailing-padding change replaced, extended with the top margin.

## What Changes

- The article view again reserves `padding_y` blank rows as a permanent margin
  between the content viewport and the borders, applied symmetrically above and
  below the content (default 1, configurable).
- The content viewport is `h - 2*padding_y - 2` rows tall; the trailing-content
  padding appended to the article is removed.
- An inline image's BOTTOM boundary puts the caption's last line on the
  viewport's last line, directly above the bottom margin — not flush against
  the border.
- `padding_x` (default 2) and `padding_y` (default 1) remain separate,
  independently overridable options.
- The scrollbar thumb and the bottom-border `N/M` line indicator track the
  smaller viewport again.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: the "Article reader border rendering" requirement reverts to the
  fixed vertical margin around the content viewport (above and below); the two
  trailing-padding scenarios are removed and the "content fills the interior"
  wording is reverted. The "Atomic image scroll behavior" bottom-boundary
  sentence about the viewport filling the interior is reverted.
- `configuration`: the "Configurable article padding" requirement reverts to
  the vertical padding applied as a permanent margin above and below the
  article content rather than trailing content rows.

## Impact

- Reverts the implementation from article-padding-snap-flush:
  `internal/ui/render/geometry.go` (`ContentGeom` restores `padY`/`effPadY`),
  `internal/ui/render/border.go` (`ClampPadY`, `RenderArticleBorder` `padY`
  param, below-content blank-row loop), `internal/ui/article.go` (trailing
  append removed, callers), `internal/ui/image.go`,
  `internal/ui/article_selection.go`.
- Adds the symmetric top margin: `ContentGeom` computes `viewportH =
  h - 2*effPadY - 2`, `RenderArticleBorder` renders `effPadY` blank rows below
  the top border before the content, and `contentRect` starts the content area
  at `row 1 + effPadY`.
- Tests restored and updated: `internal/ui/geometry_test.go`,
  `internal/ui/render_safety_test.go`, `internal/ui/render_test.go`,
  `internal/ui/mouse_selection_test.go`, `internal/ui/article_image_test.go`
  (including the trailing-padding tests added by article-padding-snap-flush).
- Main specs `openspec/specs/reader-ui/spec.md` and
  `openspec/specs/configuration/spec.md` reverted to the original wording with
  the top margin added.
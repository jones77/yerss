## Why

The article view reserves `padding_y` blank rows as a fixed strip below the
content viewport, outside the scrollable area. When an inline image snaps to
its BOTTOM boundary, its caption's last line lands on the viewport's last
content row — but that fixed strip still sits between it and the bottom border,
so the caption never reaches the last line of the viewport. The snap contract
says the bottom boundary is the block's last line at the viewport's last row;
the fixed padding strip makes that position visually a line short. The padding
belongs inside the scrollable content, not in a fixed strip the content can
never reach.

## What Changes

- The article content viewport fills the full interior between the top and
  bottom borders (`viewportH = h - 2`); the frame renders no fixed blank rows
  below the content.
- The vertical padding (`padding_y`, default 1) is applied as that many blank
  rows appended to the end of the article's scrollable content — "vertical
  padding below the article content", scrollable rather than fixed. A
  bottom-snapped inline image caption's last line is therefore the viewport's
  last line (flush at the bottom border), satisfying both phrasings of the
  intent.
- `padding_x` (default 2) and `padding_y` (default 1) remain separate,
  independently overridable config options.
- The in-content blank separator line after each image caption (the padding
  between the caption and the next paragraph/photo) is unchanged; it is the
  image block's own trailing blank line, not the vertical padding option.
- The scrollbar thumb and the bottom-border `N/M` line indicator now track the
  larger viewport (`h - 2`), so the thumb's travel matches the content that
  fills the interior.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: the "Article reader border rendering" requirement changes — the
  content area has no fixed vertical padding below the viewport; content fills
  the interior to the bottom border. The "Atomic image scroll behavior"
  requirement's bottom boundary is unchanged in wording but now lands flush at
  the viewport's last line.
- `configuration`: the "Configurable article padding" requirement is clarified
  — horizontal and vertical padding remain separately configurable (defaults 2
  and 1); the vertical padding is applied as trailing blank content rows rather
  than a fixed strip below the content.

## Impact

- `internal/config/config.go`: unchanged (both `PaddingX` and `PaddingY` stay).
- `internal/ui/render/geometry.go`: `ContentGeom` drops the `padY`/`effPadY`
  outputs; `viewportH = h - 2`.
- `internal/ui/render/border.go`: `RenderArticleBorder` drops the `padY` param
  and the below-content blank-row loop; `ClampPadY` is removed.
- `internal/ui/article.go`: `composeArticle` appends `padding_y` blank rows to
  the end of the article content; the `ContentGeom`/`RenderArticleBorder`
  callers lose the `padY` argument.
- `internal/ui/image.go`, `internal/ui/article_selection.go`: `ContentGeom`
  callers updated.
- Tests: `internal/config/config_test.go` (kept as-is), `internal/ui/geometry_test.go`,
  `internal/ui/render_safety_test.go`, `internal/ui/render_test.go` (scrollbar
  thumb geometry and `N/M` labels), plus a new article-view test asserting a
  bottom-snapped caption's last line is the last visible row.
- Specs: `openspec/specs/reader-ui/spec.md` and
  `openspec/specs/configuration/spec.md` updated.
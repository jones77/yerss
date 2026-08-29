## 1. Scroll state model

- [x] 1.1 Define a scroll-state value in `internal/ui` carrying `{totalH, viewportH, offset}` with a `thumb(trackH) (top, height int)` method implementing the design math: `scrollableRange = totalH - viewportH`; `thumbH = clamp(trackH*viewportH/totalH, 1, max(1, trackH-1))`; `thumbTop = (trackH-thumbH)*offset/scrollableRange`
- [x] 1.2 Add a `percent() int` method to the scroll state returning `100` when `totalH <= viewportH` (fits viewport) and `round(100*offset/scrollableRange)` otherwise, matching the existing `ScrollPercent()*100` label value
- [x] 1.3 Make `thumb` report the fits-viewport case (`totalH <= viewportH`) as a full-track fill: `top = 0`, `height = trackH`, so callers can render every track row as accent

## 2. Border thumb rendering

- [x] 2.1 In `renderArticleBorder` (`internal/ui/border.go`), replace the `percent int` parameter with the scroll-state value and delete the `fillH := interiorH * percent / 100` computation
- [x] 2.2 Render the right edge per track row from the thumb geometry: accent `fill` for rows in `[thumbTop, thumbTop+thumbH)` and dim `unfill` for the rest; in the fits-viewport case every track row is accent `fill`
- [x] 2.3 Derive the bottom-border percentage internally from the scroll state and pass it to `bottomBorder`; keep the ` %d%% scrolled ` format unchanged

## 3. Article model wiring

- [x] 3.1 In `renderArticle` (`internal/ui/article.go`), build the scroll state from the live viewport (`TotalLineCount()`, `Height`, `YOffset`) and pass it to `renderArticleBorder`
- [x] 3.2 Remove the cached `articleState.percent` field and the `updateArticlePercent` function; drop its call sites in `newArticleState` and `updateArticle`

## 4. Tests

- [x] 4.1 Update `TestRenderArticleBorderHeightClamp` (`internal/ui/render_safety_test.go`) to pass the new scroll-state argument; confirm the `len(lines) <= h` assertion still holds for all `h`/`padY` cases
- [x] 4.2 Add a thumb-on-open test: open a scrollable article at offset 0 and assert a thumb at least one row tall is rendered in the accent style at the top of the right border, the rest of the track is dim, and the bottom border reads "0% scrolled"
- [x] 4.3 Extend the thumb-position test to ~42% and 100% scroll: assert the thumb's top row moves to the rounded 42% travel position at 42%, and sits at the bottom of the track (`thumbTop = trackH - thumbH`) at 100%, with bottom-border labels "42% scrolled" and "100% scrolled"
- [x] 4.4 Add a thumb-bounds test across several total/viewport size combinations: assert the thumb height stays within `[1, trackH-1]` for every scrollable case
- [x] 4.5 Keep `TestRenderArticleBorder` passing: a short article renders a fully accent-filled right border with the bottom border reading "100% scrolled"

## 5. Verification

- [x] 5.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`; resolve any fallout from the `renderArticleBorder` signature change and the removed `percent` field
- [x] 5.2 Visually confirm in a real terminal that opening a long article shows the thumb immediately, that it moves (not grows) while scrolling, and that it never fills the entire right border until the article fits the viewport

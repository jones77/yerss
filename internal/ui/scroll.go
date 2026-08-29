package ui

// scrollState is the article viewport's live scroll position. It carries the
// total content height, the visible viewport height, and the current scroll
// offset so the border renderer can compute both the thumb geometry and the
// percent label from a single source of truth.
type scrollState struct {
	totalH    int
	viewportH int
	offset    int
}

// fitsViewport reports whether the entire content fits without scrolling.
func (s scrollState) fitsViewport() bool {
	return s.totalH <= s.viewportH
}

// thumb returns the accent thumb's top row and height within a track of the
// given height. When the content fits the viewport there is nothing to scroll
// and the whole track is filled (top 0, height trackH). For scrollable content
// the thumb is proportional to the visible fraction, clamped to a minimum of 1
// row and a maximum of trackH-1 rows (or 1 row when trackH is 1), and its top
// row is positioned by the scroll offset within the thumb's travel range.
func (s scrollState) thumb(trackH int) (top, height int) {
	if s.fitsViewport() {
		return 0, trackH
	}
	thumbH := trackH * s.viewportH / s.totalH
	thumbH = min(max(thumbH, 1), max(1, trackH-1))
	scrollableRange := s.totalH - s.viewportH
	thumbTop := (trackH - thumbH) * s.offset / scrollableRange
	return thumbTop, thumbH
}

// percent returns the scroll percentage for the bottom-border label: 100 when
// the content fits the viewport, otherwise the offset's share of the scrollable
// range rounded to the nearest percent.
func (s scrollState) percent() int {
	if s.fitsViewport() {
		return 100
	}
	scrollableRange := s.totalH - s.viewportH
	return int(float64(s.offset)/float64(scrollableRange)*100 + 0.5)
}

// bottomLine returns the line number of the last viewable line at the bottom
// of the viewport, clamped to the total line count.
func (s scrollState) bottomLine() int {
	last := s.offset + s.viewportH
	if last > s.totalH {
		last = s.totalH
	}
	return last
}
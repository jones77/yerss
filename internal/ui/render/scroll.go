package render

// ScrollState is the article viewport's live scroll position. It carries the
// total content height, the visible viewport height, and the current scroll
// offset so the border renderer can compute both the thumb geometry and the
// percent label from a single source of truth.
type ScrollState struct {
	TotalH    int
	ViewportH int
	Offset    int
}

// FitsViewport reports whether the entire content fits without scrolling.
func (s ScrollState) FitsViewport() bool {
	return s.TotalH <= s.ViewportH
}

// Thumb returns the accent thumb's top row and height within a track of the
// given height. When the content fits the viewport there is nothing to scroll
// and the whole track is filled (top 0, height trackH), though the renderer
// omits the thumb entirely in that case. For scrollable content the thumb is
// proportional to the visible fraction, clamped to a minimum of 1 row and a
// maximum of trackH-1 rows (or 1 row when trackH is 1), and its top row is
// positioned by the scroll offset within the thumb's travel range.
func (s ScrollState) Thumb(trackH int) (top, height int) {
	if s.FitsViewport() {
		return 0, trackH
	}
	thumbH := trackH * s.ViewportH / s.TotalH
	thumbH = min(max(thumbH, 1), max(1, trackH-1))
	scrollableRange := s.TotalH - s.ViewportH
	thumbTop := (trackH - thumbH) * s.Offset / scrollableRange
	return thumbTop, thumbH
}

// Percent returns the scroll percentage for the bottom-border label: 100 when
// the content fits the viewport, otherwise the offset's share of the scrollable
// range rounded to the nearest percent.
func (s ScrollState) Percent() int {
	if s.FitsViewport() {
		return 100
	}
	scrollableRange := s.TotalH - s.ViewportH
	return (s.Offset*100 + scrollableRange/2) / scrollableRange
}

// BottomLine returns the line number of the last viewable line at the bottom
// of the viewport, clamped to the total line count.
func (s ScrollState) BottomLine() int {
	last := s.Offset + s.ViewportH
	if last > s.TotalH {
		last = s.TotalH
	}
	return last
}

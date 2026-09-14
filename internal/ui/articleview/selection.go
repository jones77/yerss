package articleview

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

// Selection tracks a mouse text selection in the article viewport. Cells are
// in content-area coordinates (relative to the visible viewport window, i.e.
// 0..viewportH rows and 0..contentW columns). Active means there is a range to
// highlight; Tracking means the mouse button is currently held and dragging
// extends the range.
type Selection struct {
	Active   bool
	Tracking bool
	AnchorX  int
	AnchorY  int
	CurX     int
	CurY     int
}

// Cell is a content-area coordinate.
type Cell struct {
	X, Y int
}

// rangeFor normalizes the selection into the upper-left and lower-right cells.
func (s Selection) rangeFor() (lo, hi Cell) {
	lo = Cell{s.AnchorX, s.AnchorY}
	hi = Cell{s.CurX, s.CurY}
	if lo.Y > hi.Y || (lo.Y == hi.Y && lo.X > hi.X) {
		lo, hi = hi, lo
	}
	return lo, hi
}

// selectionRange returns the column span of the selection on a given row, or
// (0, 0) when the row is outside the selection's vertical extent.
func selectionRange(lo, hi Cell, row, lineWidth int) (from, to int) {
	if lo.Y == hi.Y {
		return lo.X, hi.X
	}
	switch {
	case row == lo.Y:
		return lo.X, lineWidth
	case row == hi.Y:
		return 0, hi.X
	case row > lo.Y && row < hi.Y:
		return 0, lineWidth
	}
	return 0, 0
}

// ContentRect returns the content-area rectangle (x0, y0, w, h): the region
// inside the article border where the viewport content draws. The content
// begins one column past the left padding and one row past the effective top
// padding (the top border's vertical margin plus the padding).
func (s *State) ContentRect() (x0, y0, w, h int) {
	padX := s.sess.Config().Display.PaddingX
	w, h, effPadY := render.ContentGeom(s.width, s.height, padX, s.sess.Config().Display.PaddingY)
	return 1 + padX, 1 + effPadY, w, h
}

// ContentCell maps a terminal cell to a content-area cell, returning ok=false
// when it falls outside the content area (in the border or padding).
func (s *State) ContentCell(x, y int) (Cell, bool) {
	x0, y0, w, h := s.ContentRect()
	if x < x0 || y < y0 || x >= x0+w || y >= y0+h {
		return Cell{}, false
	}
	return Cell{X: x - x0, Y: y - y0}, true
}

// ContentCellClamped maps a terminal cell to a content-area cell, clamping
// out-of-bounds coordinates into the content area.
func (s *State) ContentCellClamped(x, y int) Cell {
	x0, y0, w, h := s.ContentRect()
	if x < x0 {
		x = x0
	}
	if x >= x0+w {
		x = x0 + w - 1
	}
	if y < y0 {
		y = y0
	}
	if y >= y0+h {
		y = y0 + h - 1
	}
	return Cell{X: x - x0, Y: y - y0}
}

// LinkSpan is a clickable OSC 8 hyperlink's visible column span in one line.
type LinkSpan struct {
	Start int
	End   int
	URL   string
}

// LinkAtContentCell returns the URL of the OSC 8 hyperlink whose visible span
// covers the content-area cell, or "" when the cell is not on a link.
func (s *State) LinkAtContentCell(x, y int) string {
	visible := strings.Split(s.Viewport.View(), "\n")
	if y < 0 || y >= len(visible) {
		return ""
	}
	for _, span := range ParseLinkSpans(visible[y]) {
		if x >= span.Start && x < span.End {
			return span.URL
		}
	}
	return ""
}

// ParseLinkSpans walks an ANSI/OSC-8 styled line and returns the column spans
// of its hyperlinks. An unterminated OSC 8 (escaped or truncated) closes the
// current span; a new terminated OSC 8 opens one at the current column.
func ParseLinkSpans(line string) []LinkSpan {
	var spans []LinkSpan
	x := 0
	flush := func(url string, start int) {
		if url != "" && x > start {
			spans = append(spans, LinkSpan{Start: start, End: x, URL: url})
		}
	}
	activeURL := ""
	activeStart := 0
	for _, seg := range compose.ScanANSI(line) {
		switch seg.Kind {
		case compose.SegText:
			x += compose.WidthIn(line, seg.Start, seg.End)
		case compose.SegOSC:
			payload, terminated := seg.OSC(line)
			if !terminated {
				flush(activeURL, activeStart)
				activeURL = ""
				return spans
			}
			flush(activeURL, activeStart)
			activeURL = ""
			activeStart = 0
			if !strings.HasPrefix(payload, "8;") {
				continue
			}
			rest := payload[2:]
			u := rest
			if i := strings.LastIndex(rest, ";"); i >= 0 {
				u = rest[i+1:]
			}
			if u != "" {
				activeURL = u
				activeStart = x
			}
		}
	}
	flush(activeURL, activeStart)
	return spans
}

// SelectedText returns the text covered by the current selection, stripped of
// ANSI styles, with trailing whitespace trimmed on every row but the last.
func (s *State) SelectedText() string {
	if !s.Sel.Active {
		return ""
	}
	visible := strings.Split(s.Viewport.View(), "\n")
	lo, hi := s.Sel.rangeFor()
	var rows []string
	for y := lo.Y; y <= hi.Y && y < len(visible); y++ {
		line := ansi.Strip(visible[y])
		width := ansi.StringWidth(line)
		from, to := selectionRange(lo, hi, y, width)
		if from < to {
			if slice := ansi.Cut(line, from, to); slice != "" {
				if y < hi.Y {
					rows = append(rows, strings.TrimRight(slice, " "))
				} else {
					rows = append(rows, slice)
				}
			}
		}
	}
	return strings.Join(rows, "\n")
}

// MarkRead marks the article read in the store, once per State: the flag
// guards against repeat writes for the same composed article.
func (s *State) MarkRead(st *store.Store) {
	if s.readMarked || s.article.Read {
		s.readMarked = true
		return
	}
	s.readMarked = true
	s.article.Read = true
	_ = st.SetRead(s.id, true)
}

// highlightSelection applies the inverted selection style to the selected cell
// range of each visible content line. Lines outside the range are returned
// unchanged; ANSI and OSC 8 escape sequences on a line are preserved by
// ansi.Cut/ansi.Truncate. The reverse-video SGR is emitted directly so the
// highlight works regardless of the terminal profile.
func highlightSelection(lines []string, sel Selection) []string {
	if !sel.Active {
		return lines
	}
	lo, hi := sel.rangeFor()
	for i := lo.Y; i <= hi.Y && i < len(lines); i++ {
		line := lines[i]
		width := ansi.StringWidth(line)
		from, to := selectionRange(lo, hi, i, width)
		if to > width {
			to = width
		}
		if from >= to {
			continue
		}
		lines[i] = compose.InvertRange(line, from, to)
	}
	return lines
}
package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

// textSelection tracks a mouse text selection in the article viewport. Cells
// are in content-area coordinates (relative to the visible viewport window,
// i.e. 0..viewportH rows and 0..contentW columns). active means there is a
// range to highlight; tracking means the mouse button is currently held and
// dragging extends the range.
type textSelection struct {
	active   bool
	tracking bool
	anchorX  int
	anchorY  int
	curX     int
	curY     int
}

// cell is a content-area coordinate.
type cell struct {
	x, y int
}

// rangeFor returns the normalized selection range: lo is the upper-left cell
// and hi the lower-right cell of the selected block.
func (s textSelection) rangeFor() (lo, hi cell) {
	lo = cell{x: s.anchorX, y: s.anchorY}
	hi = cell{x: s.curX, y: s.curY}
	if lo.y > hi.y || (lo.y == hi.y && lo.x > hi.x) {
		lo, hi = hi, lo
	}
	return lo, hi
}

// selectionRange returns the selected column range [from, to) on one row: the
// anchor/cur cells when the selection spans a single row, the full line width
// on the first and last rows of a multi-row selection, and the full line in
// between. lineWidth must be the display width of the row's visible text. It is
// the single source of truth shared by the copy and highlight paths so they
// cannot drift apart.
func selectionRange(lo, hi cell, row, lineWidth int) (from, to int) {
	if lo.y == hi.y {
		return lo.x, hi.x
	}
	if row == lo.y {
		return lo.x, lineWidth
	}
	if row == hi.y {
		return 0, hi.x
	}
	return 0, lineWidth
}

func (m *Model) contentRect() (x0, y0, w, h int) {
	padX := m.sess.Config().Display.PaddingX
	w, h, effPadY := render.ContentGeom(m.width, m.height, padX, m.sess.Config().Display.PaddingY)
	return 1 + padX, 1 + effPadY, w, h
}

// contentCell maps a screen cell to a content-area cell, reporting ok=false
// when the press is outside the content area (border, padding, status bar).
func (m *Model) contentCell(x, y int) (cell, bool) {
	x0, y0, w, h := m.contentRect()
	if x < x0 || x >= x0+w || y < y0 || y >= y0+h {
		return cell{}, false
	}
	return cell{x: x - x0, y: y - y0}, true
}

// contentCellClamped maps a screen cell to a content-area cell, clamping the
// result into the content area so a drag that leaves it stays in range.
func (m *Model) contentCellClamped(x, y int) cell {
	x0, y0, w, h := m.contentRect()
	cx := x - x0
	if cx < 0 {
		cx = 0
	}
	if cx >= w {
		cx = w - 1
	}
	cy := y - y0
	if cy < 0 {
		cy = 0
	}
	if cy >= h {
		cy = h - 1
	}
	return cell{x: cx, y: cy}
}

// linkSpan is a hyperlink covering the display columns [start, end) on a
// single rendered line.
type linkSpan struct {
	start, end int
	url        string
}

// linkAtContentCell reports the URL of any hyperlink covering content cell
// (x, y), or "" when the cell is not on a link. The cell is indexed against the
// viewport's visible lines, matching the layout the user sees.
func (m *Model) linkAtContentCell(x, y int) string {
	visible := strings.Split(m.article.viewport.View(), "\n")
	if y < 0 || y >= len(visible) {
		return ""
	}
	for _, sp := range parseLinkSpans(visible[y]) {
		if x >= sp.start && x < sp.end {
			return sp.url
		}
	}
	return ""
}

// parseLinkSpans scans a rendered ANSI line for OSC 8 hyperlink escapes and
// returns the display-column span each URL covers. It tracks the current
// column as it walks printable runes (honoring wide characters), clips spans
// when a link is reset, and skips CSI/SGR and other escape sequences. The
// terminator may be BEL or the ESC-backslash string terminator.
func parseLinkSpans(line string) []linkSpan {
	var spans []linkSpan
	url := ""
	start := 0
	x := 0
	flush := func(col int) {
		if url != "" {
			spans = append(spans, linkSpan{start: start, end: col, url: url})
		}
	}
	for _, seg := range compose.ScanANSI(line) {
		switch seg.Kind {
		case compose.SegOSC:
			payload, terminated := seg.OSC(line)
			if !terminated {
				flush(x)
				return spans
			}
			flush(x)
			if strings.HasPrefix(payload, "8;") {
				rest := payload[2:]
				u := rest
				if idx := strings.LastIndex(rest, ";"); idx >= 0 {
					u = rest[idx+1:]
				}
				if u == "" {
					url = ""
				} else {
					url = u
					start = x
				}
			}
		case compose.SegText:
			x += compose.WidthIn(line, seg.Start, seg.End)
		}
	}
	flush(x)
	return spans
}

// selectedText returns the plain text within the selection range, stripped of
// ANSI styling and OSC 8 escapes, preserving the wrapped line breaks. It
// returns "" for an inactive or zero-width selection.
func (s *articleState) selectedText() string {
	if !s.sel.active {
		return ""
	}
	visible := strings.Split(s.viewport.View(), "\n")
	lo, hi := s.sel.rangeFor()
	var b strings.Builder
	for row := lo.y; row <= hi.y; row++ {
		if row < 0 || row >= len(visible) {
			continue
		}
		line := ansi.Strip(visible[row])
		from, to := selectionRange(lo, hi, row, ansi.StringWidth(line))
		if from < to {
			slice := ansi.Cut(line, from, to)
			if row < hi.y {
				slice = strings.TrimRight(slice, " ")
			}
			b.WriteString(slice)
		}
		if row < hi.y {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// markRead implements the read-status rules: an article that fits the
// viewport is marked read on open; a longer article is marked read on the
// first downward scroll.
func (s *articleState) markRead(st *store.Store) {
	if s.readMarked || s.article.Read {
		s.readMarked = true
		return
	}
	s.readMarked = true
	s.article.Read = true
	_ = st.SetRead(s.id, true)
}

func highlightSelection(lines []string, sel textSelection) []string {
	if !sel.active {
		return lines
	}
	lo, hi := sel.rangeFor()
	const revOn, revOff = "\x1b[7m", "\x1b[0m"
	out := make([]string, len(lines))
	for i, line := range lines {
		if i < lo.y || i > hi.y {
			out[i] = line
			continue
		}
		width := ansi.StringWidth(line)
		from, to := selectionRange(lo, hi, i, width)
		if to > width {
			to = width
		}
		if from >= to {
			out[i] = line
			continue
		}
		before := ansi.Truncate(line, from, "")
		selSeg := ansi.Cut(line, from, to)
		after := ansi.Cut(line, to, width)
		out[i] = before + revOn + selSeg + revOff + after
	}
	return out
}

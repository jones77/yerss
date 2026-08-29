package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// borderGlyphs holds the glyphs used to draw the article border. In ASCII
// mode box-drawing characters are replaced with plain equivalents. The right
// edge acts as a scrollbar: the unfilled track is a grey single line and the
// filled thumb is a bright double line (`║` Unicode, `|` ASCII).
type borderGlyphs struct {
	tl, bl, tr, br string
	h, v           string
	fill, unfill   string
	ellipsis       string
	bullet         string
	expand, collapse string
}

func glyphsFor(ascii bool) borderGlyphs {
	if ascii {
		return borderGlyphs{tl: "+", bl: "+", tr: "+", br: "+", h: "-", v: ":", fill: "|", unfill: ":", ellipsis: "...", bullet: ".", expand: "v", collapse: ">"}
	}
	return borderGlyphs{tl: "┌", bl: "└", tr: "┐", br: "┘", h: "─", v: "│", fill: "║", unfill: "│", ellipsis: "…", bullet: "·", expand: "▾", collapse: "▸"}
}

// renderArticleBorder draws the article reader frame: a thin border with the
// date and title inline in the top edge, a percent-scrolled indicator in the
// bottom edge, and a double-line right edge that acts as a scrollbar. A
// contiguous accent thumb positioned by the scroll offset represents the
// currently visible portion of the content, and the rest of the track is dim.
// Content is padded inside the border by padX columns on each side and padY
// rows above and below.
func renderArticleBorder(w, h, padX, padY int, g borderGlyphs, p Palette, date, title string, sc scrollState, content []string) string {
	if h < 1 {
		h = 1
	}
	// Clamp padY so the frame (top + padY + viewport + padY + bottom = 2 +
	// 2*padY + viewportH) never exceeds the terminal height. This only
	// activates on degenerate terminals where h - 2*padY - 2 < 1.
	effPadY := padY
	if maxPad := (h - 3) / 2; effPadY > maxPad {
		effPadY = maxPad
	}
	if effPadY < 0 {
		effPadY = 0
	}
	interiorW := w - 2
	if interiorW < 1 {
		interiorW = 1
	}
	textW := interiorW - 2*padX
	if textW < 1 {
		textW = 1
	}
	viewportH := h - 2*effPadY - 2
	if viewportH < 1 {
		viewportH = 1
	}

	interiorH := viewportH + 2*effPadY
	if interiorH < 1 {
		interiorH = 1
	}
	thumbTop, thumbH := sc.thumb(interiorH)

	border := lipgloss.NewStyle().Foreground(p.Border)
	thumb := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff"))

	ipad := strings.Repeat(" ", padX)
	space := strings.Repeat(" ", interiorW)

	right := func(row int) string {
		if row >= thumbTop && row < thumbTop+thumbH {
			return thumb.Render(g.fill)
		}
		return border.Render(g.unfill)
	}

	var lines []string
	lines = append(lines, topBorder(w, g, p, date, title))

	row := 0
	blank := func() string {
		return border.Render(g.v) + space + right(row)
	}
	for i := 0; i < effPadY; i++ {
		lines = append(lines, blank())
		row++
	}
	for i := 0; i < viewportH; i++ {
		var line string
		if i < len(content) {
			line = truncate(content[i], textW)
		}
		line = padRight(line, textW)
		lines = append(lines, border.Render(g.v)+ipad+line+ipad+right(row))
		row++
	}
	for i := 0; i < effPadY; i++ {
		lines = append(lines, blank())
		row++
	}
	lines = append(lines, bottomBorder(w, g, p, sc))

	// Defensive clamp to exactly h lines on the most degenerate sizes.
	if len(lines) > h {
		lines = lines[:h]
	} else if len(lines) < h {
		for len(lines) < h {
			lines = append(lines, blank())
			row++
		}
	}
	return strings.Join(lines, "\n")
}

// topBorder renders the article frame's top edge as `leftRail + core +
// rightRail`, where the rails are corner-plus-dash on the left (`┌─`) and
// dash-plus-corner on the right (`─╖`). The date is always fully visible; when
// the title does not fit, it is truncated and ends with the ellipsis glyph and
// the right rail keeps its single dash so the two ends mirror each other. When
// the title fits, leftover space is filled with horizontal dashes.
func topBorder(w int, g borderGlyphs, p Palette, date, title string) string {
	leftRail := g.tl + g.h
	rightRail := g.h + g.tr
	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}
	coreW := innerW - 2
	if coreW < 1 {
		coreW = 1
	}
	prefix := " " + date + " " + g.bullet + " "
	if date == "" {
		prefix = " "
	}
	titleW := coreW - runewidth.StringWidth(prefix) - 1
	if titleW < 0 {
		titleW = 0
	}

	style := lipgloss.NewStyle().Foreground(p.Border)

	if runewidth.StringWidth(title) <= titleW {
		core := prefix + title + " "
		fill := coreW - runewidth.StringWidth(core)
		if fill < 0 {
			fill = 0
		}
		return style.Render(leftRail + core + strings.Repeat(g.h, fill) + rightRail)
	}

	cut := truncate(title, titleW-runewidth.StringWidth(g.ellipsis))
	core := prefix + cut + g.ellipsis + " "
	return style.Render(leftRail + core + rightRail)
}

// bottomBorder renders the article frame's bottom edge as `leftCorner +
// horizontal line + core + horizontal line + rightCorner`, mirroring the top
// border. The core is a left-aligned `o: open in browser` hint and a
// right-aligned `<percent>% · <bottomLine>/<totalLines>` position indicator,
// each inset one space from the horizontal line (like the top border's date
// and title) and separated from the dash fill by a space. The bullet between
// the percent and the line ratio is g.bullet.
func bottomBorder(w int, g borderGlyphs, p Palette, sc scrollState) string {
	hint := "o: open in browser"
	indicator := fmt.Sprintf("%d%% %s %d/%d", sc.percent(), g.bullet, sc.bottomLine(), sc.totalH)
	inner := w - 2
	if inner < 1 {
		inner = 1
	}
	// The fixed chrome is 8 columns: corner+h+space on each side plus a space
	// either side of the dash fill, leaving inner-6 for hint+indicator+fill.
	// When that is too tight, the indicator (then the hint) is truncated.
	avail := inner - 6
	if avail < 0 {
		avail = 0
	}
	hint = truncate(hint, avail)
	hw := ansi.StringWidth(hint)
	if avail-hw < 0 {
		indicator = ""
	} else {
		indicator = truncate(indicator, avail-hw)
	}
	iw := ansi.StringWidth(indicator)
	fill := inner - 6 - hw - iw
	if fill < 0 {
		fill = 0
	}
	style := lipgloss.NewStyle().Foreground(p.Border)
	return style.Render(g.bl + g.h + " " + hint + " " + strings.Repeat(g.h, fill) + " " + indicator + " " + g.h + g.br)
}
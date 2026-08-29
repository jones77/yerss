package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// borderGlyphs holds the glyphs used to draw the article border. In ASCII
// mode box-drawing characters are replaced with plain equivalents and the
// double-line right edge uses # for filled and : for unfilled segments.
type borderGlyphs struct {
	tl, bl, tr, br string
	h, v           string
	fill, unfill   string
	ellipsis       string
}

func glyphsFor(ascii bool) borderGlyphs {
	if ascii {
		return borderGlyphs{tl: "+", bl: "+", tr: "+", br: "+", h: "-", v: "|", fill: "#", unfill: ":", ellipsis: "..."}
	}
	return borderGlyphs{tl: "┌", bl: "└", tr: "╖", br: "╜", h: "─", v: "│", fill: "║", unfill: "║", ellipsis: "…"}
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

	accent := lipgloss.NewStyle().Foreground(p.Accent)
	dim := lipgloss.NewStyle().Foreground(p.Dim)

	ipad := strings.Repeat(" ", padX)
	space := strings.Repeat(" ", interiorW)

	right := func(row int) string {
		if row >= thumbTop && row < thumbTop+thumbH {
			return accent.Render(g.fill)
		}
		return dim.Render(g.unfill)
	}

	var lines []string
	lines = append(lines, topBorder(w, g, p, date, title))

	row := 0
	blank := func() string {
		return g.v + space + right(row)
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
		lines = append(lines, g.v+ipad+line+ipad+right(row))
		row++
	}
	for i := 0; i < effPadY; i++ {
		lines = append(lines, blank())
		row++
	}
	lines = append(lines, bottomBorder(w, g, p, sc.percent()))

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
	prefix := " " + date + " · "
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

func bottomBorder(w int, g borderGlyphs, p Palette, percent int) string {
	text := fmt.Sprintf(" %d%% scrolled ", percent)
	inner := w - 2
	if inner < 1 {
		inner = 1
	}
	text = truncate(text, inner)
	tw := runewidth.StringWidth(text)
	total := inner - tw
	if total < 0 {
		total = 0
	}
	left, right := total/2, total-total/2
	style := lipgloss.NewStyle().Foreground(p.Border)
	return style.Render(g.bl + strings.Repeat(g.h, left) + text + strings.Repeat(g.h, right) + g.br)
}
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
}

func glyphsFor(ascii bool) borderGlyphs {
	if ascii {
		return borderGlyphs{tl: "+", bl: "+", tr: "+", br: "+", h: "-", v: "|", fill: "#", unfill: ":"}
	}
	return borderGlyphs{tl: "┌", bl: "└", tr: "╖", br: "╜", h: "─", v: "│", fill: "║", unfill: "║"}
}

// renderArticleBorder draws the article reader frame: a thin border with the
// date and title inline in the top edge, a percent-scrolled indicator in the
// bottom edge, and a double-line right edge that acts as a scrollbar. A
// contiguous accent thumb positioned by the scroll offset represents the
// currently visible portion of the content, and the rest of the track is dim.
// The frame is inset from the terminal by padX columns on each side and content
// is padded inside by padY rows above and below.
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
	contentW := w - 2*padX - 2
	if contentW < 1 {
		contentW = 1
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

	pad := strings.Repeat(" ", padX)
	space := strings.Repeat(" ", contentW)

	right := func(row int) string {
		if row >= thumbTop && row < thumbTop+thumbH {
			return accent.Render(g.fill)
		}
		return dim.Render(g.unfill)
	}
	left := func() string {
		return pad + g.v
	}

	var lines []string
	lines = append(lines, topBorder(w, padX, g, p, date, title))

	row := 0
	for i := 0; i < effPadY; i++ {
		lines = append(lines, left()+space+right(row))
		row++
	}
	for i := 0; i < viewportH; i++ {
		var line string
		if i < len(content) {
			line = truncate(content[i], contentW)
		}
		line = padRight(line, contentW)
		lines = append(lines, left()+line+right(row))
		row++
	}
	for i := 0; i < effPadY; i++ {
		lines = append(lines, left()+space+right(row))
		row++
	}
	lines = append(lines, bottomBorder(w, padX, g, p, sc.percent()))

	// Defensive clamp to exactly h lines on the most degenerate sizes.
	if len(lines) > h {
		lines = lines[:h]
	} else if len(lines) < h {
		for len(lines) < h {
			lines = append(lines, left()+space+right(row))
			row++
		}
	}
	return strings.Join(lines, "\n")
}

func topBorder(w, padX int, g borderGlyphs, p Palette, date, title string) string {
	text := " " + date + " · " + title + " "
	text = truncate(text, w-2*padX-3)
	fill := (w - 2*padX - 2) - 1 - runewidth.StringWidth(text)
	if fill < 0 {
		fill = 0
	}
	style := lipgloss.NewStyle().Foreground(p.Border)
	return strings.Repeat(" ", padX) + style.Render(g.tl+g.h+text+strings.Repeat(g.h, fill)+g.tr)
}

func bottomBorder(w, padX int, g borderGlyphs, p Palette, percent int) string {
	text := fmt.Sprintf(" %d%% scrolled ", percent)
	inner := (w - 2*padX - 2)
	text = truncate(text, inner)
	tw := runewidth.StringWidth(text)
	total := inner - tw
	if total < 0 {
		total = 0
	}
	left, right := total/2, total-total/2
	style := lipgloss.NewStyle().Foreground(p.Border)
	return strings.Repeat(" ", padX) + style.Render(g.bl+strings.Repeat(g.h, left)+text+strings.Repeat(g.h, right)+g.br)
}
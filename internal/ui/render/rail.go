package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RailGlyph returns the tree-rail corner for day group i of n: `┌` on the
// first day group, `└` on the last when it is collapsed (the terminal visible
// row), and `├` everywhere else — including an expanded last group, whose
// articles continue the rail below the header. Corners are computed against
// the full group list, not the scroll window, so they do not move as the
// window scrolls. Only day headers render a rail; article rows carry no tree
// glyph.
func RailGlyph(g BorderGlyphs, collapsed bool, n, i int) string {
	if i == 0 {
		return g.TL
	}
	if i == n-1 && collapsed {
		return g.BL
	}
	return g.Tee
}

// DayHeader renders a collapsible day-group header: the rail corner followed
// by four connector glyphs, a space, and the label, all in the given style.
func DayHeader(style lipgloss.Style, g BorderGlyphs, corner, label string) string {
	connector := strings.Repeat(g.H, 4)
	return style.Render(corner + connector + " " + label)
}
package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Overlay draws pop over base, centered by display width. Centering is computed
// from visible columns with ANSI escapes excluded, so styled popups are centered
// rather than glued to the left edge. Base content outside the popup region is
// preserved.
func Overlay(base, pop string) string {
	baseLines := strings.Split(base, "\n")
	popLines := strings.Split(pop, "\n")

	bw := lipgloss.Width(base)
	bh := len(baseLines)
	pw := lipgloss.Width(pop)
	ph := len(popLines)
	if pw == 0 || ph == 0 {
		return base
	}
	if pw > bw {
		pw = bw
	}
	if ph > bh {
		ph = bh
		popLines = popLines[:bh]
	}

	bx := (bw - pw) / 2
	if bx < 0 {
		bx = 0
	}
	by := (bh - ph) / 2
	if by < 0 {
		by = 0
	}

	out := make([]string, len(baseLines))
	copy(out, baseLines)
	for i, l := range popLines {
		t := by + i
		if t < 0 || t >= len(out) {
			continue
		}
		out[t] = spliceStyled(out[t], l, bx, pw)
	}
	return strings.Join(out, "\n")
}

// spliceStyled overlays popRow onto baseRow so that popRow's visible content
// begins at display column bx. The base row's content before bx and after
// bx+pw is preserved, and ANSI escape sequences in both rows are kept intact.
func spliceStyled(baseRow, popRow string, bx, pw int) string {
	left := ansi.Truncate(baseRow, bx, "")
	if ansi.StringWidth(left) < bx {
		left = PadRight(left, bx)
	}
	right := ansi.Cut(baseRow, bx+pw, ansi.StringWidth(baseRow))
	return left + popRow + right
}
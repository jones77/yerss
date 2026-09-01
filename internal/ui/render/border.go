package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// BorderGlyphs holds the glyphs used to draw the article border. In ASCII
// mode box-drawing characters are replaced with plain equivalents. The right
// edge acts as a scrollbar: the unfilled track is a grey single line and the
// filled thumb is a bright single line (`│` Unicode, `|` ASCII); the model's
// glyphs helper swaps in the double-line thumb (`║`) when configured.
type BorderGlyphs struct {
	TL, BL, TR, BR string
	Tee            string
	H, V           string
	Fill, Unfill   string
	Ellipsis       string
	Bullet         string
	Half           string
}

// GlyphsFor returns the glyph set for the given mode. In ASCII mode every
// box-drawing character is replaced with a plain equivalent.
func GlyphsFor(ascii bool) BorderGlyphs {
	if ascii {
		return BorderGlyphs{TL: "+", BL: "+", TR: "+", BR: "+", Tee: "+", H: "-", V: ":", Fill: "|", Unfill: ":", Ellipsis: "...", Bullet: ".", Half: "1/2"}
	}
	return BorderGlyphs{TL: "┌", BL: "└", TR: "┐", BR: "┘", Tee: "├", H: "─", V: "│", Fill: "│", Unfill: "│", Ellipsis: "…", Bullet: "·", Half: "½"}
}

// ClampPadY clamps the vertical content padding so the article frame (top +
// 2*padY + viewport + bottom) never exceeds the terminal height on degenerate
// terminals. It is shared by the content geometry so the border renderer and
// the mouse coordinate mapping agree on the content area layout.
func ClampPadY(padY, h int) int {
	if maxPad := (h - 3) / 2; padY > maxPad {
		padY = maxPad
	}
	if padY < 0 {
		return 0
	}
	return padY
}

// RenderArticleBorder draws the article reader frame: a thin border with the
// date and title inline in the top edge, a percent-scrolled indicator in the
// bottom edge, and a double-line right edge that acts as a scrollbar. A
// contiguous accent thumb positioned by the scroll offset represents the
// currently visible portion of the content, and the rest of the track is dim.
// Content is padded inside the border by padX columns on each side and padY
// rows above and below.
func RenderArticleBorder(w, h, padX, padY int, g BorderGlyphs, p Palette, date, title string, sc ScrollState, content []string) string {
	if h < 1 {
		h = 1
	}
	textW, viewportH, effPadY := ContentGeom(w, h, padX, padY)
	interiorW := max(1, w-2)
	interiorH := max(1, h-2)
	thumbTop, thumbH := sc.Thumb(interiorH)

	border := lipgloss.NewStyle().Foreground(p.Dim)
	thumb := lipgloss.NewStyle().Bold(true).Foreground(p.StatusBar)

	ipad := strings.Repeat(" ", padX)
	space := strings.Repeat(" ", interiorW)

	right := func(row int) string {
		if !sc.FitsViewport() && row >= thumbTop && row < thumbTop+thumbH {
			return thumb.Render(g.Fill)
		}
		return border.Render(g.Unfill)
	}

	var lines []string
	lines = append(lines, TopBorder(w, g, p, date, title))

	// Content starts directly beneath the top border with no padding above it;
	// the vertical padding pads only the bottom.
	row := 0
	blank := func() string {
		return border.Render(g.V) + space + right(row)
	}
	for i := 0; i < viewportH; i++ {
		var line string
		if i < len(content) {
			line = Truncate(content[i], textW)
		}
		line = PadRight(line, textW)
		lines = append(lines, border.Render(g.V)+ipad+line+ipad+right(row))
		row++
	}
	for i := 0; i < effPadY; i++ {
		lines = append(lines, blank())
		row++
	}
	lines = append(lines, BottomBorder(w, g, p, sc))

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

// TopBorder renders the article frame's top edge as `leftRail + core +
// rightRail`, where the rails are corner-plus-dash on the left (`┌─`) and
// dash-plus-corner on the right (`─╖`). The date is always fully visible; when
// the title does not fit, it is truncated and ends with the ellipsis glyph and
// the right rail keeps its single dash so the two ends mirror each other. When
// the title fits, leftover space is filled with horizontal dashes. The date,
// title, and ellipsis are styled in bright blue (12); the bullet and the rails
// stay in the grey role (8).
func TopBorder(w int, g BorderGlyphs, p Palette, date, title string) string {
	leftRail := g.TL + g.H
	rightRail := g.H + g.TR
	innerW := max(1, w-2)
	coreW := max(1, innerW-2)
	prefix := " " + date + " " + g.Bullet + " "
	if date == "" {
		prefix = " "
	}
	titleW := coreW - runewidth.StringWidth(prefix) - 1
	if titleW < 0 {
		titleW = 0
	}

	style := lipgloss.NewStyle().Foreground(p.Dim)
	textStyle := lipgloss.NewStyle().Foreground(p.StatusBar)
	hLeft := style.Render(leftRail)
	hRight := style.Render(rightRail)

	if runewidth.StringWidth(title) <= titleW {
		fill := coreW - runewidth.StringWidth(prefix+title) - 1
		if fill < 0 {
			fill = 0
		}
		right := style.Render(strings.Repeat(g.H, fill) + rightRail)
		if date == "" {
			return hLeft + " " + textStyle.Render(title) + " " + right
		}
		return hLeft + textStyle.Render(" "+date+" ") + style.Render(g.Bullet) + " " + textStyle.Render(title) + " " + right
	}

	cut := Truncate(title, titleW-runewidth.StringWidth(g.Ellipsis))
	if date == "" {
		return hLeft + " " + textStyle.Render(cut+g.Ellipsis) + " " + hRight
	}
	return hLeft + textStyle.Render(" "+date+" ") + style.Render(g.Bullet) + " " + textStyle.Render(cut+g.Ellipsis) + " " + hRight
}

// BottomBorder renders the article frame's bottom edge as `leftCorner +
// horizontal line + core + horizontal line + rightCorner`, mirroring the top
// border. The core is a left-aligned `o: open article in browser` hint and a
// right-aligned `?: help · <percent>% · <bottomLine>/<totalLines>` position
// indicator (the help affordance first), each inset one space from the
// horizontal line (like the top border's date and title) and separated from
// the dash fill by a space. The bullets separating the affordance, percent,
// and line ratio are g.Bullet. The hint, affordance, percent, and ratio are
// styled in bright blue (12); the bullets and the rails stay in the grey role
// (8).
func BottomBorder(w int, g BorderGlyphs, p Palette, sc ScrollState) string {
	style := lipgloss.NewStyle().Foreground(p.Dim)
	textStyle := lipgloss.NewStyle().Foreground(p.StatusBar)

	// The left-aligned hint is the open hint; the right-aligned position
	// indicator is led by the `?: help` affordance, each separated by the
	// bullet (dim).
	hint := textStyle.Render("o: open article in browser")
	percentStr := fmt.Sprintf("%d%%", sc.Percent())
	ratioStr := fmt.Sprintf("%d/%d", sc.BottomLine(), sc.TotalH)
	indicator := "?: help" + " " + g.Bullet + " " + percentStr + " " + g.Bullet + " " + ratioStr
	inner := max(1, w-2)
	// The fixed chrome is 8 columns: corner+h+space on each side plus a space
	// either side of the dash fill, leaving inner-6 for hint+indicator+fill.
	// When that is too tight, the indicator (then the hint) is truncated.
	avail := inner - 6
	if avail < 0 {
		avail = 0
	}
	hint = Truncate(hint, avail)
	hw := ansi.StringWidth(hint)
	if avail-hw < 0 {
		indicator = ""
	} else {
		indicator = Truncate(indicator, avail-hw)
	}
	iw := ansi.StringWidth(indicator)
	fill := inner - 6 - hw - iw
	if fill < 0 {
		fill = 0
	}
	var ind strings.Builder
	if indicator != "" {
		for i, part := range strings.Split(indicator, g.Bullet) {
			if i > 0 {
				ind.WriteString(style.Render(g.Bullet))
			}
			ind.WriteString(textStyle.Render(part))
		}
	}
	dash := style.Render(g.H)
	return style.Render(g.BL) + dash + " " + hint + " " + style.Render(strings.Repeat(g.H, fill)) + " " + ind.String() + " " + dash + style.Render(g.BR)
}

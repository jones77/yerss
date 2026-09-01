package ui

import (
	"yerss/internal/ui/render"
)

// glyphs returns the glyph set for the current mode, applying the configured
// scrollbar thumb style: single line by default, or a double line when the
// scrollbar option is set to "double" (Unicode only; ASCII fallback has no
// double-line glyph).
func (m *Model) glyphs() render.BorderGlyphs {
	g := render.GlyphsFor(m.ascii)
	if m.sess.Config().Display.Scrollbar == "double" && !m.ascii {
		g.Fill = "║"
	}
	return g
}

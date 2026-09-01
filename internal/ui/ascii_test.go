package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/ui/render"
)

func TestGlyphsForASCIIAndUnicodePairs(t *testing.T) {
	u, a := render.GlyphsFor(false), render.GlyphsFor(true)
	if a.TL != "+" || a.BL != "+" || a.TR != "+" || a.BR != "+" || a.Tee != "+" {
		t.Errorf("ascii corner/tee glyphs = %q%q%q%q%q, want all +", a.TL, a.BL, a.TR, a.BR, a.Tee)
	}
	if u.TL != "┌" || u.BL != "└" || u.TR != "┐" || u.BR != "┘" || u.Tee != "├" {
		t.Errorf("unicode corner/tee glyphs = %q%q%q%q%q", u.TL, u.BL, u.TR, u.BR, u.Tee)
	}
	if a.H != "-" || a.Bullet != "." || a.Ellipsis != "..." || a.Half != "1/2" {
		t.Errorf("ascii misc glyphs = %q %q %q %q, want - . ... 1/2", a.H, a.Bullet, a.Ellipsis, a.Half)
	}
	if u.H != "─" || u.Bullet != "·" || u.Ellipsis != "…" || u.Half != "½" {
		t.Errorf("unicode misc glyphs = %q %q %q %q", u.H, u.Bullet, u.Ellipsis, u.Half)
	}
}

func TestScrollbarConfigSelectsThumbGlyph(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(false)
	if got := m.glyphs().Fill; got != "│" {
		t.Errorf("default unicode thumb = %q, want single line │", got)
	}
	m.sess.Config().Display.Scrollbar = "double"
	if got := m.glyphs().Fill; got != "║" {
		t.Errorf("double unicode thumb = %q, want ║", got)
	}
	m.SetAscii(true)
	if got := m.glyphs().Fill; got != "|" {
		t.Errorf("ascii thumb = %q, want |", got)
	}
	m.sess.Config().Display.Scrollbar = "single"
	if got := m.glyphs().Fill; got != "|" {
		t.Errorf("ascii single thumb = %q, want |", got)
	}
}

func TestSetAsciiForcesAsciiGlyphs(t *testing.T) {
	m, _ := newTestModel(t)
	// Force ascii off regardless of what config or terminal detection chose.
	m.SetAscii(false)
	m.article = m.newArticleState(store.Article{Title: "x", Content: "<p>y</p>"})

	before := ansi.Strip(m.renderArticle())
	if !strings.HasPrefix(before, "┌") {
		t.Fatalf("expected a unicode top border before SetAscii(true), got %q", before)
	}

	m.SetAscii(true)
	m.article.viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	after := ansi.Strip(m.renderArticle())
	if !strings.HasPrefix(after, "+") {
		t.Errorf("expected an ASCII top border after SetAscii(true), got %q", after)
	}
	if !strings.Contains(after, "-") || !strings.Contains(after, "|") {
		t.Errorf("expected ASCII glyphs (-, |) after SetAscii(true), got %q", after)
	}
}

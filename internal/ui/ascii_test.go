package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

func TestGlyphsForASCIIAndUnicodePairs(t *testing.T) {
	u, a := glyphsFor(false), glyphsFor(true)
	if a.tl != "+" || a.bl != "+" || a.tr != "+" || a.br != "+" || a.tee != "+" {
		t.Errorf("ascii corner/tee glyphs = %q%q%q%q%q, want all +", a.tl, a.bl, a.tr, a.br, a.tee)
	}
	if u.tl != "┌" || u.bl != "└" || u.tr != "┐" || u.br != "┘" || u.tee != "├" {
		t.Errorf("unicode corner/tee glyphs = %q%q%q%q%q", u.tl, u.bl, u.tr, u.br, u.tee)
	}
	if a.h != "-" || a.bullet != "." || a.ellipsis != "..." || a.half != "1/2" {
		t.Errorf("ascii misc glyphs = %q %q %q %q, want - . ... 1/2", a.h, a.bullet, a.ellipsis, a.half)
	}
	if u.h != "─" || u.bullet != "·" || u.ellipsis != "…" || u.half != "½" {
		t.Errorf("unicode misc glyphs = %q %q %q %q", u.h, u.bullet, u.ellipsis, u.half)
	}
}

func TestScrollbarConfigSelectsThumbGlyph(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(false)
	if got := m.glyphs().fill; got != "│" {
		t.Errorf("default unicode thumb = %q, want single line │", got)
	}
	m.cfg.Display.Scrollbar = "double"
	if got := m.glyphs().fill; got != "║" {
		t.Errorf("double unicode thumb = %q, want ║", got)
	}
	m.SetAscii(true)
	if got := m.glyphs().fill; got != "|" {
		t.Errorf("ascii thumb = %q, want |", got)
	}
	m.cfg.Display.Scrollbar = "single"
	if got := m.glyphs().fill; got != "|" {
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
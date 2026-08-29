package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

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
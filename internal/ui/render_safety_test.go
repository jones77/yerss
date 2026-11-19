package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"yerss/internal"
	"yerss/internal/ui/render"
)

func TestRenderArticleZeroHeight(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 0
	m.height = 0
	m.article.SetSize(m.width, m.height)
	m.article = m.article.Recompose(internal.Article{
		Title:   "x",
		Content: "<p>hello world</p>",
	})
	if s := m.article.Render(); s == "" {
		t.Error("expected non-empty render")
	}
}

func TestTruncateLongTitle(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := render.Truncate(long, 40)
	if runewidth.StringWidth(got) > 40 {
		t.Errorf("truncate exceeded width: %q", got)
	}
}

func TestTruncateCJKTitle(t *testing.T) {
	title := "日本語のタイトルです"
	if runewidth.StringWidth(title) != 20 {
		t.Fatalf("expected 20 display columns for test title, got %d", runewidth.StringWidth(title))
	}
	got := render.Truncate(title, 10)
	if runewidth.StringWidth(got) != 10 {
		t.Errorf("expected exactly 10 columns, got %d: %q", runewidth.StringWidth(got), got)
	}
}

func TestRenderArticleBorderHeightClamp(t *testing.T) {
	g := render.GlyphsFor(true)
	p := render.DarkPalette()
	cases := []struct {
		name    string
		h, padY int
	}{
		{"h3 padY2", 3, 2},
		{"h2 padY2", 2, 2},
		{"h1 padY2", 1, 2},
		{"h3 padY5", 3, 5},
	}
	for _, c := range cases {
		got := render.RenderArticleBorder(80, c.h, 2, c.padY, g, p, "", "t", "", "o: open article in browser", render.ScrollState{TotalH: 100, ViewportH: 10, Offset: 0}, []string{"x"})
		if n := len(strings.Split(got, "\n")); n > c.h {
			t.Errorf("%s: emitted %d lines, want <= %d", c.name, n, c.h)
		}
	}
}

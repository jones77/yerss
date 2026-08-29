package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"yerss/internal/store"
)

func TestRenderListZeroHeight(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 0
	m.height = 0
	s := m.renderList()
	if got := len(strings.Split(s, "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s)
	}
}

func TestRenderListEmptyHeightOne(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 0
	m.height = 1
	s := m.renderList()
	if got := len(strings.Split(s, "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s)
	}
}

func TestRenderListNonEmptyHeightOne(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.width = 0
	m.height = 1
	s := m.renderList()
	if got := len(strings.Split(s, "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s)
	}
}

func TestRenderListZeroWidth(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.width = 0
	m.height = 24
	if s := m.renderList(); s == "" {
		t.Error("expected non-empty render")
	}
}

func TestRenderArticleZeroHeight(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 0
	m.height = 0
	m.article = m.newArticleState(store.Article{
		Title:   "x",
		Content: "<p>hello world</p>",
	})
	if s := m.renderArticle(); s == "" {
		t.Error("expected non-empty render")
	}
}

func TestListWindow(t *testing.T) {
	cases := []struct {
		name              string
		n, cursor, height int
		wantStart, wantEnd int
	}{
		{"all fit", 5, 2, 10, 0, 5},
		{"empty list", 0, 0, 10, 0, 0},
		{"exact fit", 10, 5, 10, 0, 10},
		{"cursor at top", 20, 0, 10, 0, 10},
		{"cursor middle", 20, 10, 10, 5, 15},
		{"cursor at bottom", 20, 19, 10, 10, 20},
		{"height one at top", 20, 0, 1, 0, 1},
		{"height one at bottom", 20, 19, 1, 19, 20},
	}
	for _, c := range cases {
		start, end := listWindow(c.n, c.cursor, c.height)
		if start != c.wantStart || end != c.wantEnd {
			t.Errorf("%s: listWindow(%d,%d,%d) = (%d,%d), want (%d,%d)",
				c.name, c.n, c.cursor, c.height, start, end, c.wantStart, c.wantEnd)
		}
	}
}

func TestTruncateLongTitle(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := truncate(long, 40)
	if runewidth.StringWidth(got) > 40 {
		t.Errorf("truncate exceeded width: %q", got)
	}
}

func TestTruncateCJKTitle(t *testing.T) {
	title := "日本語のタイトルです"
	if runewidth.StringWidth(title) != 20 {
		t.Fatalf("expected 20 display columns for test title, got %d", runewidth.StringWidth(title))
	}
	got := truncate(title, 10)
	if runewidth.StringWidth(got) != 10 {
		t.Errorf("expected exactly 10 columns, got %d: %q", runewidth.StringWidth(got), got)
	}
}

func TestRenderArticleBorderHeightClamp(t *testing.T) {
	g := glyphsFor(true)
	p := darkPalette()
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
		got := renderArticleBorder(80, c.h, 2, c.padY, g, p, "", "t", 0, []string{"x"})
		if n := len(strings.Split(got, "\n")); n > c.h {
			t.Errorf("%s: emitted %d lines, want <= %d", c.name, n, c.h)
		}
	}
}
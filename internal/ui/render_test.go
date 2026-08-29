package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

func TestRenderList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	insertArticle(t, st, "two", nil)
	m.loadList()
	s := m.renderList()
	if !strings.Contains(s, "one") || !strings.Contains(s, "two") {
		t.Errorf("list render missing articles: %q", s)
	}
	if !strings.Contains(s, "2 articles") {
		t.Errorf("status bar missing count: %q", s)
	}
}

func TestRenderArticleBorder(t *testing.T) {
	m, st := newTestModel(t)
	a := store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "g1",
		Title:       "Hello world article title",
		Link:        "https://example.com/1",
		PublishedAt: mustParseTime(t, "2026-01-02T15:04:05Z"),
		Content:     "<p>a short paragraph</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.newArticleState(a)

	s := m.renderArticle()
	lines := strings.Split(s, "\n")
	if len(lines) != m.height {
		t.Fatalf("expected %d lines, got %d", m.height, len(lines))
	}
	if !strings.Contains(lines[0], "Hello world article title") {
		t.Errorf("top border missing title: %q", lines[0])
	}
	if !strings.Contains(lines[0], "2026-01-02") {
		t.Errorf("top border missing date: %q", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "100% scrolled") {
		t.Errorf("bottom border missing percent: %q", lines[len(lines)-1])
	}
}

func TestRenderArticleBorderASCII(t *testing.T) {
	m, _ := newTestModel(t)
	m.ascii = true
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g2",
		Title:   "ascii",
		Content: "<p>x</p>",
	}
	m.article = m.newArticleState(a)
	s := m.renderArticle()
	lines := strings.Split(s, "\n")
	if !strings.Contains(lines[0], "+") {
		t.Errorf("ascii top border missing +: %q", lines[0])
	}
}

func TestRenderTagPopup(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{
		tags: []store.TagCount{{Name: "tech", Total: 3, Unread: 1}},
	}
	s := m.renderTagPopup()
	if !strings.Contains(s, "tech") {
		t.Errorf("popup missing tag: %q", s)
	}
}

func TestOverlayCentersByDisplayWidth(t *testing.T) {
	base := "abcdefghijklmnopqrstuvwxyz" // 26 wide
	pop := "XXX"                          // 3 wide
	got := overlay(base, pop)
	want := "abcdefghijk" + "XXX" + "opqrstuvwxyz"
	if got != want {
		t.Errorf("overlay = %q, want %q", got, want)
	}
}

func TestOverlayPreservesBaseOutsidePopup(t *testing.T) {
	base := strings.Join([]string{
		strings.Repeat("a", 20),
		strings.Repeat("b", 20),
		strings.Repeat("c", 20),
	}, "\n")
	pop := "YY"
	got := overlay(base, pop)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != strings.Repeat("a", 20) {
		t.Errorf("top line modified: %q", lines[0])
	}
	if lines[2] != strings.Repeat("c", 20) {
		t.Errorf("bottom line modified: %q", lines[2])
	}
	// popup width 2, base width 20 → centered at columns 9..10.
	if want := "bbbbbbbbb" + "YY" + "bbbbbbbbb"; lines[1] != want {
		t.Errorf("middle line not centered: %q, want %q", lines[1], want)
	}
}

func TestOverlayIgnoresANSIWidth(t *testing.T) {
	base := "aaaaaaaaaa" // 10 wide
	pop := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("XX")
	got := overlay(base, pop)
	stripped := ansi.Strip(got)
	if stripped != "aaaaXXaaaa" {
		t.Errorf("styled overlay stripped = %q, want %q", stripped, "aaaaXXaaaa")
	}
}

func mustParseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

func TestRenderArticleHeaderURLIsOSC8(t *testing.T) {
	m, _ := newTestModel(t)
	url := "https://example.com/post"
	m.article = m.newArticleState(store.Article{Title: "t", Link: url, Content: "<p>x</p>"})

	rendered := strings.Join(m.article.lines, "\n")
	if !strings.Contains(rendered, "\x1b]8;") {
		t.Errorf("header URL not wrapped in OSC 8: %q", rendered)
	}
	if !strings.Contains(ansi.Strip(rendered), url) {
		t.Errorf("visible header URL text missing: %q", ansi.Strip(rendered))
	}
}

func TestTruncateWidthWithOSC8(t *testing.T) {
	url := "https://example.com/post"
	linked := ansi.SetHyperlink(url) + "a-very-long-link-text-that-exceeds" + ansi.ResetHyperlink()
	got := truncate(linked, 10)
	if ansi.StringWidth(got) != 10 {
		t.Errorf("truncated width = %d, want 10: %q", ansi.StringWidth(got), got)
	}
	if !strings.Contains(got, ansi.SetHyperlink(url)) {
		t.Errorf("truncate dropped the OSC 8 open sequence: %q", got)
	}
	if !strings.Contains(got, ansi.ResetHyperlink()) {
		t.Errorf("truncate dropped the OSC 8 reset sequence: %q", got)
	}
}

func TestPadRightWidthWithOSC8(t *testing.T) {
	url := "https://example.com/post"
	linked := ansi.SetHyperlink(url) + "world" + ansi.ResetHyperlink()
	got := padRight(linked, 10)
	if ansi.StringWidth(got) != 10 {
		t.Errorf("padded width = %d, want 10: %q", ansi.StringWidth(got), got)
	}
}

func TestArticleOSC8LinksRenderInASCIIMode(t *testing.T) {
	m, _ := newTestModel(t)
	m.ascii = true
	url := "https://example.com/post"
	m.article = m.newArticleState(store.Article{
		Title:   "ascii links",
		Link:    url,
		Content: `<p>see <a href="` + url + `">post</a> now</p>`,
	})

	s := m.renderArticle()
	if !strings.Contains(s, "\x1b]8;") {
		t.Errorf("ascii mode should still emit OSC 8: %q", s)
	}
	if !strings.Contains(ansi.Strip(s), "post") {
		t.Errorf("link text missing in ascii mode: %q", ansi.Strip(s))
	}
	if !strings.Contains(ansi.Strip(s), "+") {
		t.Errorf("ascii mode should use ASCII borders: %q", ansi.Strip(s))
	}
}
package ui

import (
	"reflect"
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

func TestParseLinkSpans(t *testing.T) {
	url := "https://example.com/a"
	open := ansi.SetHyperlink(url)
	reset := ansi.ResetHyperlink()
	two := ansi.SetHyperlink("https://example.com/b")
	cases := []struct {
		name string
		in   string
		want []linkSpan
	}{
		{"single", open + "link" + reset, []linkSpan{{0, 4, url}}},
		{"two spans", open + "ab" + reset + "xy" + two + "cd" + reset,
			[]linkSpan{{0, 2, url}, {4, 6, "https://example.com/b"}}},
		{"wide chars", open + "界" + reset, []linkSpan{{0, 2, url}}},
		{"none", "plain text", nil},
		{"st terminator", "\x1b]8;;" + url + "\x1b\\" + "hi" + "\x1b]8;;\x1b\\", []linkSpan{{0, 2, url}}},
	}
	for _, c := range cases {
		got := parseLinkSpans(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: parseLinkSpans(%q) = %+v, want %+v", c.name, c.in, got, c.want)
		}
	}
}

func TestArticleClickOpensLink(t *testing.T) {
	url := "https://example.com/body"
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{
		Title:   "t",
		Content: `<p>go <a href="` + url + `">there</a> now</p>`,
	})

	var hitX, hitY int
	var found bool
	for y, ln := range strings.Split(m.article.viewport.View(), "\n") {
		for _, sp := range parseLinkSpans(ln) {
			if sp.url == url {
				hitX, hitY = sp.start, y
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no body link found in rendered article")
	}
	if got := m.linkAtContentCell(hitX, hitY); got != url {
		t.Errorf("linkAtContentCell(%d,%d) = %q, want %q", hitX, hitY, got, url)
	}

	// A left click on the link cell should return an open command and must not
	// start a text selection.
	x0, y0, _, _ := m.contentRect()
	cmd := m.updateArticleMouse(mouseClick(x0+hitX, y0+hitY))
	if cmd == nil {
		t.Fatal("clicking a link should return an open command")
	}
	if m.article.sel.active {
		t.Error("clicking a link should not start a text selection")
	}
}

func TestArticleClickOffLinkSelects(t *testing.T) {
	url := "https://example.com/body"
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{
		Title:   "t",
		Content: `<p>go <a href="` + url + `">there</a> now</p>`,
	})

	x0, y0, _, _ := m.contentRect()
	// Click the leading "go " text, which precedes the link, so no link is hit.
	m.updateArticleMouse(mouseClick(x0+0, y0+1))
	if !m.article.sel.active {
		t.Error("clicking non-link text should start a selection")
	}
}
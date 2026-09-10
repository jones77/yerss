package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/ui/render"
)

func TestRenderArticleHeaderURLIsOSC8(t *testing.T) {
	m, _ := newTestModel(t)
	url := "https://example.com/post"
	m.article = m.newArticleState(store.Article{Title: "t", Link: url, Content: "<p>x</p>"})

	rendered := strings.Join(m.article.lines, "\n")
	if !strings.Contains(rendered, "\x1b]8;") {
		t.Errorf("header URL not wrapped in OSC 8: %q", rendered)
	}
	visible := ansi.Strip(rendered)
	stripped := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if !strings.Contains(visible, stripped) {
		t.Errorf("visible header URL text missing: %q", visible)
	}
	if n := strings.Count(visible, stripped); n != 1 {
		t.Errorf("header URL should be rendered exactly once, got %d: %q", n, visible)
	}
}

func TestReaderHeaderOrder(t *testing.T) {
	m, _ := newTestModel(t)
	url := "https://example.com/post"
	m.article = m.newArticleState(store.Article{
		Title:   "Headline",
		Author:  "Jane",
		Link:    url,
		Content: "<p>x</p>",
	})

	// Header: URL, blank line, bold title, author line directly beneath;
	// then a blank line before the body. Title and author are joined
	// adjacent because the user asked for no space between them.
	lines := strippedLines(m.article.lines)
	if strings.TrimRight(lines[0], " ") != "example.com/post" {
		t.Errorf("first content line = %q, want the URL", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "" {
		t.Errorf("second content line = %q, want a blank line", lines[1])
	}
	if !strings.Contains(m.article.lines[2], "Headline") || !strings.Contains(m.article.lines[2], ";1m") {
		t.Errorf("title line = %q, want bold-rendered title", m.article.lines[2])
	}
	if strings.TrimRight(lines[3], " ") != "by Jane" {
		t.Errorf("author line = %q, want it directly beneath the title: %q", lines[3], "by Jane")
	}
	if strings.TrimSpace(lines[4]) != "" {
		t.Errorf("line after the header = %q, want a blank line", lines[4])
	}
}

func TestReaderHeaderOrderWithoutAuthor(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{
		Title:   "Headline",
		Link:    "https://example.com/post",
		Content: "<p>x</p>",
	})

	lines := strippedLines(m.article.lines)
	if strings.TrimRight(lines[0], " ") != "example.com/post" || strings.TrimSpace(lines[1]) != "" {
		t.Errorf("header should start with the URL and a blank line: %q", lines[:3])
	}
	if !strings.Contains(lines[2], "Headline") || strings.HasPrefix(strings.TrimRight(lines[3], " "), "by ") {
		t.Errorf("no-author header should have title and no author line: %q", lines[:4])
	}
}

func TestReaderHeaderURLWraps(t *testing.T) {
	// A URL longer than the content width wraps onto subsequent lines so the
	// whole display string is visible, each wrapped line an OSC 8 hyperlink
	// carrying the full URL.
	m, _ := newTestModel(t)
	m.width = 40
	url := "https://example.com/a/very/long/path/that/extends/beyond/the/available/width/and/must/wrap"
	m.article = m.newArticleState(store.Article{
		Title:   "t",
		Link:    url,
		Content: "<p>x</p>",
	})
	contentW := m.article.viewport.Width

	lines := strippedLines(m.article.lines)
	var urlLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			break
		}
		urlLines = append(urlLines, strings.TrimRight(l, " "))
	}
	if len(urlLines) < 2 {
		t.Fatalf("expected the URL to wrap to multiple lines, got %d: %q", len(urlLines), urlLines)
	}
	if got := strings.Join(urlLines, ""); got != stripURLScheme(url) {
		t.Errorf("wrapped header URL = %q, want the whole stripped URL %q", got, stripURLScheme(url))
	}
	for i, l := range urlLines {
		if w := ansi.StringWidth(l); w > contentW {
			t.Errorf("URL line %d width %d exceeds content width %d: %q", i, w, contentW, l)
		}
	}
	// Each wrapped line stays clickable: exactly one OSC 8 span per line,
	// targeting the full URL.
	for i := 0; i < len(urlLines); i++ {
		spans := parseLinkSpans(m.article.lines[i])
		if len(spans) != 1 || spans[0].url != url {
			t.Errorf("URL line %d spans = %+v, want one span for %q", i, spans, url)
		}
	}
}

func TestReaderHeaderWithoutLinkStartsWithTitle(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "Only title", Content: "<p>x</p>"})

	lines := ansi.Strip(strings.Join(m.article.lines, "\n"))
	if !strings.Contains(lines, "Only title") || !strings.HasPrefix(strings.Split(lines, "\n")[0], "Only title") {
		t.Errorf("link-less header should start with the title: %q", strings.Split(lines, "\n")[0])
	}
}

// strippedLines strips ANSI escapes from each rendered line.
func strippedLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
}

func TestTruncateWidthWithOSC8(t *testing.T) {
	url := "https://example.com/post"
	linked := ansi.SetHyperlink(url) + "a-very-long-link-text-that-exceeds" + ansi.ResetHyperlink()
	got := render.Truncate(linked, 10)
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
	got := render.PadRight(linked, 10)
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

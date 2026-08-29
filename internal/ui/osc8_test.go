package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

func TestHTMLToTextWrapsLinksInOSC8(t *testing.T) {
	url := "https://example.com/post"
	html := `<p>Hello <a href="` + url + `">world</a>.</p>`
	out := HTMLToText(html)

	if !strings.Contains(out, ansi.SetHyperlink(url)) {
		t.Errorf("link text not wrapped in OSC 8 carrying %q: %q", url, out)
	}
	if !strings.Contains(out, ansi.ResetHyperlink()) {
		t.Errorf("missing OSC 8 hyperlink reset: %q", out)
	}
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "world ( "+url+" )") {
		t.Errorf("visible text should keep link text and URL fallback: %q", stripped)
	}
	if !strings.Contains(stripped, "Hello") {
		t.Errorf("non-link text should be unaffected: %q", stripped)
	}
}

func TestHTMLToTextSingleQuotedHref(t *testing.T) {
	url := "https://example.com/post"
	html := `<p>See <a href='` + url + `'>post</a> now.</p>`
	out := HTMLToText(html)
	if !strings.Contains(out, ansi.SetHyperlink(url)) {
		t.Errorf("single-quoted href not wrapped in OSC 8: %q", out)
	}
}

func TestHTMLToTextPlainParagraphHasNoEscapes(t *testing.T) {
	out := HTMLToText("<p>just plain text</p>")
	if strings.Contains(out, "\x1b") {
		t.Errorf("plain text should not contain escape sequences: %q", out)
	}
}

func TestRenderArticleHeaderURLIsOSC8(t *testing.T) {
	m, _ := newTestModel(t)
	url := "https://example.com/post"
	m.article = m.newArticleState(store.Article{Title: "t", Link: url, Content: "<p>x</p>"})

	text := renderArticleContent(*m.article.article)
	if !strings.Contains(text, ansi.SetHyperlink(url)) {
		t.Errorf("header URL not wrapped in OSC 8: %q", text)
	}
	if !strings.Contains(text, ansi.ResetHyperlink()) {
		t.Errorf("header URL missing OSC 8 reset: %q", text)
	}
	if !strings.Contains(ansi.Strip(text), url) {
		t.Errorf("visible header URL text missing: %q", ansi.Strip(text))
	}
}

func TestWrapTextWidthWithOSC8(t *testing.T) {
	url := "https://example.com/post"
	linked := ansi.SetHyperlink(url) + "world" + ansi.ResetHyperlink()
	out := wrapText(linked+" ", 10)
	for _, l := range strings.Split(out, "\n") {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line width %d > 10: %q", ansi.StringWidth(l), l)
		}
	}
	if strings.Count(out, "\n") != 0 {
		t.Errorf("a short OSC 8 link should fit one line, got %q", out)
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
	if !strings.Contains(s, ansi.SetHyperlink(url)) {
		t.Errorf("ascii mode should still emit OSC 8: %q", s)
	}
	if !strings.Contains(ansi.Strip(s), "post") {
		t.Errorf("link text missing in ascii mode: %q", ansi.Strip(s))
	}
	if !strings.Contains(ansi.Strip(s), "+") {
		t.Errorf("ascii mode should use ASCII borders: %q", ansi.Strip(s))
	}
}
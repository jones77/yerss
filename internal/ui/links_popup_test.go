package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal"
	"yerss/internal/ui/articleview"
	"yerss/internal/ui/popup"
)

func linkyArticle() internal.Article {
	return internal.Article{
		Title:   "Linked",
		Link:    "https://example.com/post",
		Content: `<p>see <a href="https://example.com/one">one</a> and <a href="https://example.com/two">two</a> again <a href="https://example.com/one">one</a></p>`,
	}
}

func articleWithLinks(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	m.article = m.article.Recompose(linkyArticle())
	m.view = viewArticle
	return m
}

func TestHarvestLinksDeduplicatesInDocumentOrder(t *testing.T) {
	links := articleview.HarvestLinks("intro [a](https://x.com/1) mid [b](<https://x.com/2>) end [c](https://x.com/1)")
	want := []string{"https://x.com/1", "https://x.com/2"}
	if len(links) != len(want) {
		t.Fatalf("harvested %d links, want %d: %+v", len(links), len(want), links)
	}
	for i, l := range links {
		if l.URL != want[i] {
			t.Errorf("links[%d].url = %q, want %q", i, l.URL, want[i])
		}
	}
	if links[0].Text != "a" || links[1].Text != "b" {
		t.Errorf("link texts = %q, %q; want a, b", links[0].Text, links[1].Text)
	}
}

func TestResolveLinkAgainstArticleOrigin(t *testing.T) {
	a := internal.Article{Link: "https://theintercept.com/2026/08/31/iran-war/", FeedURL: "https://theintercept.com/feed"}
	cases := map[string]string{
		"/tomdispatch":               "https://theintercept.com/tomdispatch",
		"https://example.com/abs":    "https://example.com/abs",
		"tomdispatch":                "https://theintercept.com/2026/08/31/iran-war/tomdispatch",
		"https://theintercept.com/x": "https://theintercept.com/x",
	}
	for in, want := range cases {
		if got := articleview.ResolveLink(a, in); got != want {
			t.Errorf("articleview.ResolveLink(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHarvestArticleLinksResolvesRelativeTargets(t *testing.T) {
	a := internal.Article{
		Link:    "https://theintercept.com/2026/08/31/iran-war/",
		Content: `<p><a href="/tomdispatch">TomDispatch</a></p>`,
	}
	links := articleview.HarvestArticleLinks(a)
	if len(links) != 1 || links[0].URL != "https://theintercept.com/tomdispatch" {
		t.Errorf("links = %+v, want the relative target resolved to https://theintercept.com/tomdispatch", links)
	}
}

func TestHarvestArticleLinksLinkedImageBanner(t *testing.T) {
	// A promo banner renders as a linked image whose link wraps the sentinel
	// and the banner text; the harvest must still surface the link, resolved
	// to an absolute target, with the sentinel stripped from the link text.
	a := internal.Article{
		Link: "https://theintercept.com/2026/08/31/iran-war/",
		Content: `<aside class="promote-banner"><a class="promote-banner__link" href="/tomdispatch">` +
			`<span class="promote-banner__image"><img decoding="async" src="https://example.com/logo.jpg" alt=""></span>` +
			`<div class="promote-banner__text"><p class="promote-banner__eyebrow">Read Our Complete Coverage</p>` +
			`<h2 class="promote-banner__title">TomDispatch</h2></div></a></aside>`,
	}
	links := articleview.HarvestArticleLinks(a)
	if len(links) != 1 {
		t.Fatalf("links = %+v, want the banner link", links)
	}
	if links[0].URL != "https://theintercept.com/tomdispatch" {
		t.Errorf("banner url = %q, want https://theintercept.com/tomdispatch", links[0].URL)
	}
	if strings.Contains(links[0].Text, "\x00") || strings.Contains(links[0].Text, "img:") {
		t.Errorf("banner link text still carries the image sentinel: %q", links[0].Text)
	}
	if !strings.Contains(links[0].Text, "Read Our Complete Coverage") || !strings.Contains(links[0].Text, "TomDispatch") {
		t.Errorf("banner link text = %q, want the banner text", links[0].Text)
	}
}

func TestArticleStateHarvestsBodyLinks(t *testing.T) {
	m := articleWithLinks(t)
	urls := []string{"https://example.com/one", "https://example.com/two"}
	if len(m.article.Links()) != len(urls) {
		t.Fatalf("links = %+v, want %d body links (header excluded, deduped)", m.article.Links(), len(urls))
	}
	for i, l := range m.article.Links() {
		if l.URL != urls[i] {
			t.Errorf("links[%d].url = %q, want %q", i, l.URL, urls[i])
		}
	}
}

func TestLinkPopupOpensViaLAndRight(t *testing.T) {
	m := articleWithLinks(t)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.popupData.Kind() != popup.KindLinks {
		t.Fatalf("l should open the links popup, kind = %d", m.popupData.Kind())
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyEsc})
	if m.popupData.Kind() != popup.KindNone {
		t.Fatalf("esc should close the popup, kind = %d", m.popupData.Kind())
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRight})
	if m.popupData.Kind() != popup.KindLinks {
		t.Fatalf("right should open the links popup, kind = %d", m.popupData.Kind())
	}
}

func TestLinkPopupEnterOpensSelected(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	m.popupData.SetCursor(1) // https://example.com/two

	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return an open command for the selected link")
	}
	if m.popupData.Kind() != popup.KindNone {
		t.Errorf("popup should close after confirming, kind = %d", m.popupData.Kind())
	}
}

func TestLinkPopupEscClosesWithoutNavigating(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("esc should not navigate, got %T", cmd)
	}
	if m.popupData.Kind() != popup.KindNone || m.view != viewArticle {
		t.Errorf("esc should close the popup and stay in the article, kind = %d view = %d", m.popupData.Kind(), m.view)
	}
}

func TestLinkPopupEmptyArticle(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.article.Recompose(internal.Article{Title: "bare", Content: "<p>no links</p>"})
	m.openLinksPopup()
	if len(m.popupData.Links()) != 0 {
		t.Fatalf("expected no links, got %+v", m.popupData.Links())
	}
	s := m.popupData.Render()
	if !strings.Contains(s, "Links") {
		t.Errorf("empty popup should still render: %q", s)
	}
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Errorf("Enter on an empty popup should do nothing, got %T", cmd)
	}
}

func TestLinkPopupOOpensArticleURL(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if cmd == nil {
		t.Fatal("o with the links popup open should open the article URL")
	}
	if m.popupData.Kind() != popup.KindLinks {
		t.Errorf("o should keep the popup open, kind = %d", m.popupData.Kind())
	}
}

func TestTagPopupListOOpensNothing(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData.SetKind(popup.KindTagsList)
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if cmd != nil {
		t.Errorf("o over the list tag popup should not open anything, got %T", cmd)
	}
}

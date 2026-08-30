package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

func linkyArticle() store.Article {
	return store.Article{
		Title:   "Linked",
		Link:    "https://example.com/post",
		Content: `<p>see <a href="https://example.com/one">one</a> and <a href="https://example.com/two">two</a> again <a href="https://example.com/one">one</a></p>`,
	}
}

func articleWithLinks(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	m.article = m.newArticleState(linkyArticle())
	m.view = viewArticle
	return m
}

func TestHarvestLinksDeduplicatesInDocumentOrder(t *testing.T) {
	links := harvestLinks("intro [a](https://x.com/1) mid [b](<https://x.com/2>) end [c](https://x.com/1)")
	want := []string{"https://x.com/1", "https://x.com/2"}
	if len(links) != len(want) {
		t.Fatalf("harvested %d links, want %d: %+v", len(links), len(want), links)
	}
	for i, l := range links {
		if l.url != want[i] {
			t.Errorf("links[%d].url = %q, want %q", i, l.url, want[i])
		}
	}
	if links[0].text != "a" || links[1].text != "b" {
		t.Errorf("link texts = %q, %q; want a, b", links[0].text, links[1].text)
	}
}

func TestArticleStateHarvestsBodyLinks(t *testing.T) {
	m := articleWithLinks(t)
	urls := []string{"https://example.com/one", "https://example.com/two"}
	if len(m.article.links) != len(urls) {
		t.Fatalf("links = %+v, want %d body links (header excluded, deduped)", m.article.links, len(urls))
	}
	for i, l := range m.article.links {
		if l.url != urls[i] {
			t.Errorf("links[%d].url = %q, want %q", i, l.url, urls[i])
		}
	}
}

func TestLinkPopupOpensViaLAndRight(t *testing.T) {
	m := articleWithLinks(t)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.popup != popupLinks {
		t.Fatalf("l should open the links popup, popup = %d", m.popup)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyEsc})
	if m.popup != noPopup {
		t.Fatalf("esc should close the popup, popup = %d", m.popup)
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRight})
	if m.popup != popupLinks {
		t.Fatalf("right should open the links popup, popup = %d", m.popup)
	}
}

func TestLinkPopupRowsShowTextAndDimURL(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	m.popupData.cursor = 1

	s := m.renderLinksPopup()
	if !strings.Contains(s, "Links") {
		t.Errorf("popup missing title: %q", s)
	}
	stripped := ansi.Strip(s)
	for _, want := range []string{"one", "two", "https://example.com/one", "https://example.com/two"} {
		if !strings.Contains(stripped, want) {
			t.Errorf("popup missing %q: %q", want, stripped)
		}
	}
	if strings.Contains(stripped, "post") {
		t.Errorf("popup should not list the article's own URL: %q", stripped)
	}
	if !strings.Contains(s, "> ") {
		t.Errorf("popup missing the cursor marker: %q", s)
	}
}

func TestLinkPopupNavigationWraps(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	n := len(m.popupData.links)

	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.popupData.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m.popupData.cursor)
	}
	for i := 0; i < n-1; i++ {
		m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	if m.popupData.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", m.popupData.cursor)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyUp})
	if m.popupData.cursor != n-1 {
		t.Errorf("cursor should wrap up to %d, got %d", n-1, m.popupData.cursor)
	}
}

func TestLinkPopupEnterOpensSelected(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	m.popupData.cursor = 1 // https://example.com/two

	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return an open command for the selected link")
	}
	if m.popup != noPopup {
		t.Errorf("popup should close after confirming, popup = %d", m.popup)
	}
}

func TestLinkPopupEscClosesWithoutNavigating(t *testing.T) {
	m := articleWithLinks(t)
	m.openLinksPopup()
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Errorf("esc should not navigate, got %T", cmd)
	}
	if m.popup != noPopup || m.view != viewArticle {
		t.Errorf("esc should close the popup and stay in the article, popup = %d view = %d", m.popup, m.view)
	}
}

func TestLinkPopupEmptyArticle(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "bare", Content: "<p>no links</p>"})
	m.openLinksPopup()
	if len(m.popupData.links) != 0 {
		t.Fatalf("expected no links, got %+v", m.popupData.links)
	}
	s := m.renderLinksPopup()
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
	if m.popup != popupLinks {
		t.Errorf("o should keep the popup open, popup = %d", m.popup)
	}
}

func TestTagPopupListOOpensNothing(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	_, cmd := m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	if cmd != nil {
		t.Errorf("o over the list tag popup should not open anything, got %T", cmd)
	}
}

func TestHelpListsLinkPopup(t *testing.T) {
	m, _ := newTestModel(t)
	s := ansi.Strip(m.renderHelp())
	if !strings.Contains(s, "Link Popup") {
		t.Errorf("help should list the link_popup action: %q", s)
	}
	if !strings.Contains(s, "l, right") {
		t.Errorf("help should show the link_popup keys: %q", s)
	}
}

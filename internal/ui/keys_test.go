package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// sendKey routes a single rune through the active view's update handler.
func sendKey(m *Model, r rune) tea.Cmd {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	if m.view == viewList {
		_, cmd := m.updateList(msg)
		return cmd
	}
	if m.view == viewArticle {
		_, cmd := m.updateArticle(msg)
		return cmd
	}
	return nil
}

func TestBKeyAliasBackFromList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetFilter("tech")
	m.list.SetCursor(0)

	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Errorf("b in list should stay in list, view = %d", m.view)
	}
	if m.list.Filter() != "" {
		t.Errorf("b in list with active filter should clear it, filter = %q", m.list.Filter())
	}
}

func TestBKeyAliasBackFromArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetCursor(1) // first article row
	m.openArticle()
	if m.view != viewArticle {
		t.Fatalf("expected article view, got %d", m.view)
	}

	sendKey(m, 'b')
	if m.view != viewList {
		t.Errorf("b in article view should return to list, view = %d", m.view)
	}
}

func TestEnterKeyClosesArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetCursor(1) // first article row
	m.openArticle()
	if m.view != viewArticle {
		t.Fatalf("expected article view, got %d", m.view)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != viewList {
		t.Errorf("enter in article view should return to list, view = %d", m.view)
	}
}

func TestOneKeyGoesToTopOfList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	insertArticle(t, st, "two", nil)
	m.list.Load()
	m.list.SetCursor(2) // last row
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.list.Cursor() != 0 {
		t.Errorf("1 in list should move to the top, cursor = %d", m.list.Cursor())
	}
}

func TestOneKeyGoesToTopOfArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetCursor(1)
	m.openArticle()
	m.article.Viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	m.article.Viewport.GotoBottom()
	if m.article.Viewport.YOffset == 0 {
		t.Fatal("expected to be scrolled down before pressing 1")
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.article.Viewport.YOffset != 0 {
		t.Errorf("1 in article should scroll to the top, offset = %d", m.article.Viewport.YOffset)
	}
}

func TestOKeyAliasOpensArticleFromList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	if m.list.RowCount() != 2 { // header + article
		t.Fatalf("expected 2 visible rows, got %d", m.list.RowCount())
	}
	m.list.SetCursor(1) // article row

	sendKey(m, 'o')
	if m.view != viewArticle {
		t.Errorf("o in list on article row should open the article, view = %d", m.view)
	}
	if m.article.ID() == 0 {
		t.Error("o should have opened an article")
	}
}

func TestOKeyAliasOpenURLFromArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetCursor(1) // first article row
	m.openArticle()
	if m.view != viewArticle {
		t.Fatalf("expected article view, got %d", m.view)
	}
	m.article.Article().Link = "https://example.com/post"

	cmd := sendKey(m, 'o')
	if cmd == nil {
		t.Error("o in article view should dispatch the open-url command")
	}
}

func TestExistingKeysUnchanged(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()
	m.list.SetCursor(1)

	for _, r := range []rune{'l', '\r'} {
		m.list.SetCursor(1)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		if r == '\r' {
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		}
		m.updateList(msg)
		if m.view != viewArticle {
			t.Errorf("%q in list should open the article, view = %d", string(r), m.view)
		}
		m.view = viewList
	}

	m.view = viewArticle
	for _, r := range []rune{'h'} {
		sendKey(m, r)
		if m.view != viewList {
			t.Errorf("h in article view should return to list, view = %d", m.view)
		}
		m.view = viewArticle
	}
}

func TestOOpensOnArticleRowNotHeader(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	m.list.Load()

	m.list.SetCursor(0) // header row
	sendKey(m, 'o')
	if m.view != viewList {
		t.Errorf("o on a day header should not open an article, view = %d", m.view)
	}
}


package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
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
	m.loadList()
	m.list.filter = "tech"
	m.list.cursor = 0

	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Errorf("b in list should stay in list, view = %d", m.view)
	}
	if m.list.filter != "" {
		t.Errorf("b in list with active filter should clear it, filter = %q", m.list.filter)
	}
}

func TestBKeyAliasBackFromArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1 // first article row
	m.openArticle()
	if m.view != viewArticle {
		t.Fatalf("expected article view, got %d", m.view)
	}

	sendKey(m, 'b')
	if m.view != viewList {
		t.Errorf("b in article view should return to list, view = %d", m.view)
	}
}

func TestOKeyAliasOpensArticleFromList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	if len(m.visibleRows()) != 2 { // header + article
		t.Fatalf("expected 2 visible rows, got %d", len(m.visibleRows()))
	}
	m.list.cursor = 1 // article row

	sendKey(m, 'o')
	if m.view != viewArticle {
		t.Errorf("o in list on article row should open the article, view = %d", m.view)
	}
	if m.article.id == 0 {
		t.Error("o should have opened an article")
	}
}

func TestOKeyAliasOpenURLFromArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1 // first article row
	m.openArticle()
	if m.view != viewArticle {
		t.Fatalf("expected article view, got %d", m.view)
	}
	m.article.article.Link = "https://example.com/post"

	cmd := sendKey(m, 'o')
	if cmd == nil {
		t.Error("o in article view should dispatch the open-url command")
	}
}

func TestExistingKeysUnchanged(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1

	for _, r := range []rune{'l', '\r'} {
		m.list.cursor = 1
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
	m.loadList()

	m.list.cursor = 0 // header row
	sendKey(m, 'o')
	if m.view != viewList {
		t.Errorf("o on a day header should not open an article, view = %d", m.view)
	}
}

func TestHelpListsEveryActionInOrder(t *testing.T) {
	m, _ := newTestModel(t)
	help := m.renderHelp()
	last := 0
	for i, a := range config.AllActions() {
		keys := strings.Join(m.cfg.Keybindings[a], ", ")
		line := fmt.Sprintf("  %-18s %s", actionLabel(a), keys)
		pos := strings.Index(help, line)
		if pos < 0 {
			t.Fatalf("help missing %q", line)
		}
		if i > 0 && pos < last {
			t.Errorf("action %q out of order in help", a)
		}
		last = pos
	}
}
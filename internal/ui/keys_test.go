package ui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
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

func TestEnterKeyClosesArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1 // first article row
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
	m.loadList()
	m.list.cursor = 2 // last row
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.list.cursor != 0 {
		t.Errorf("1 in list should move to the top, cursor = %d", m.list.cursor)
	}
}

func TestOneKeyGoesToTopOfArticle(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1
	m.openArticle()
	m.article.viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	m.article.viewport.GotoBottom()
	if m.article.viewport.YOffset == 0 {
		t.Fatal("expected to be scrolled down before pressing 1")
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.article.viewport.YOffset != 0 {
		t.Errorf("1 in article should scroll to the top, offset = %d", m.article.viewport.YOffset)
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

func TestHelpListsEveryActionGrouped(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(false)
	help := ansi.Strip(m.renderHelp())
	for _, a := range config.AllActions() {
		keys := strings.Join(displayKeys(m.sess.Config().Keybindings[a]), ", ")
		found := false
		for _, l := range strings.Split(help, "\n") {
			if strings.Contains(l, m.helpLabel(a)) && strings.Contains(l, keys) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("help missing row for %q (%s %s)", a, m.helpLabel(a), keys)
		}
	}
	for _, want := range []string{"Global", "List view", "Article view"} {
		if !strings.Contains(help, want) {
			t.Errorf("help missing section %q", want)
		}
	}
	gi := strings.Index(help, "Global")
	li := strings.Index(help, "List view")
	ai := strings.Index(help, "Article view")
	if gi >= li || li >= ai {
		t.Errorf("help sections out of order: Global=%d List view=%d Article view=%d", gi, li, ai)
	}
}

func TestHelpKeysStandOutInBright(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.SetAscii(false)
	help := m.renderHelp()
	if !strings.Contains(help, "\x1b[97mctrl+c\x1b[0m") {
		t.Errorf("keys should render in the bright role (ANSI 15): %q", help)
	}
	if !strings.Contains(help, "\x1b[1;94m Key Bindings \x1b[0m") {
		t.Errorf("Key Bindings title should render bold in the chrome role (ANSI 12): %q", help)
	}
	// Section headings stay in the chrome role but are no longer bold.
	if !strings.Contains(help, "\x1b[94mGlobal\x1b[0m") {
		t.Errorf("section headings should render plain in the chrome role: %q", help)
	}
	for _, l := range strings.Split(help, "\n") {
		if strings.Contains(l, "ctrl+c") {
			if !strings.Contains(l, "\x1b[0m, \x1b[97m") {
				t.Errorf("commas between keys should stay uncolored: %q", l)
			}
			break
		}
	}
}

func TestHelpShowsSpaceKeyAsWord(t *testing.T) {
	// Load normalizes "space" to a literal " " for runtime matching; the help
	// popup must still display it as the word "space".
	km, err := config.ParseKeybindings(config.DefaultKeybindings())
	if err != nil {
		t.Fatalf("ParseKeybindings: %v", err)
	}
	cfg := config.Default()
	cfg.Keybindings = km
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	m := New(app.New(cfg, st))
	m.width = 80
	m.height = 24
	help := ansi.Strip(m.renderHelp())
	if !strings.Contains(help, "pgdn, ctrl+f, space") {
		t.Errorf("help should render the space key as the word space: %q", help)
	}
}

func TestHelpTwoColumnWithinSeventyColumns(t *testing.T) {
	m, _ := newTestModel(t)
	lines := strings.Split(m.renderHelp(), "\n")
	if len(lines) < 2 {
		t.Fatalf("help should have content lines, got %d", len(lines))
	}
	// List view must sit in the second column, on the same row as Global.
	row := ""
	for _, l := range lines {
		if strings.Contains(l, "Global") {
			row = l
			break
		}
	}
	if row == "" {
		t.Fatalf("help should render a Global row")
	}
	if i := strings.Index(row, "Global"); i < 0 || strings.Index(row, "List view") <= i {
		t.Errorf("List view should be to the right of Global: %q", row)
	}
	// The popup including its border must not exceed 70 columns.
	for i, l := range lines {
		if w := ansi.StringWidth(l); w > 70 {
			t.Errorf("help line %d is %d columns, want <= 70: %q", i, w, l)
		}
	}
}
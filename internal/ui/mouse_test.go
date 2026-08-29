package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/store"
)

func mouseClick(x, y int) tea.MouseMsg {
	return tea.MouseMsg{
		X:      x,
		Y:      y,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
}

func mouseWheel(y int, down bool) tea.MouseMsg {
	btn := tea.MouseButtonWheelUp
	if down {
		btn = tea.MouseButtonWheelDown
	}
	return tea.MouseMsg{
		X:      0,
		Y:      y,
		Button: btn,
		Action: tea.MouseActionPress,
	}
}

// twoArticleModel returns a model whose list is [header, a, b] with the
// article IDs returned in visible-row order.
func twoArticleModel(t *testing.T) (*Model, []int64) {
	t.Helper()
	m, st := newTestModel(t)
	insertArticle(t, st, "a", nil)
	insertArticle(t, st, "b", nil)
	m.loadList()
	rows := m.visibleRows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 visible rows, got %d", len(rows))
	}
	ids := []int64{
		m.list.groups[rows[1].groupIdx].articles[rows[1].artIdx].ID,
		m.list.groups[rows[2].groupIdx].articles[rows[2].artIdx].ID,
	}
	return m, ids
}

func TestClickOpensArticleAndPreservesCursor(t *testing.T) {
	m, ids := twoArticleModel(t)
	if m.view != viewList {
		t.Fatalf("expected list view, got %d", m.view)
	}

	m.updateListMouse(mouseClick(0, 2)) // row index 2 = article b
	if m.view != viewArticle {
		t.Errorf("click on article row should open it, view = %d", m.view)
	}
	if m.article.id != ids[1] {
		t.Errorf("clicked article id = %d, want %d", m.article.id, ids[1])
	}
	if m.list.cursor != 2 {
		t.Errorf("cursor after click = %d, want 2", m.list.cursor)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Errorf("b should return to list, view = %d", m.view)
	}
	if m.list.cursor != 2 {
		t.Errorf("cursor not preserved on return = %d, want 2", m.list.cursor)
	}
}

func TestClickOnDayHeaderToggles(t *testing.T) {
	m, _ := twoArticleModel(t)
	if m.list.groups[0].collapsed {
		t.Fatal("group should start expanded")
	}

	m.updateListMouse(mouseClick(0, 0)) // row 0 = day header
	if m.view != viewList {
		t.Errorf("click on header should stay in list, view = %d", m.view)
	}
	if !m.list.groups[0].collapsed {
		t.Error("click on header should collapse the day")
	}

	m.updateListMouse(mouseClick(0, 0))
	if m.list.groups[0].collapsed {
		t.Error("second click on header should expand the day")
	}
}

func TestClickOnStatusBarIgnored(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseClick(0, m.height-1))
	if m.view != viewList {
		t.Errorf("click on status bar should be ignored, view = %d", m.view)
	}
	if m.list.cursor != 0 {
		t.Errorf("cursor changed by status-bar click = %d, want 0", m.list.cursor)
	}
}

func TestClickBeyondRowsIgnored(t *testing.T) {
	m, _ := twoArticleModel(t)
	// Rows 3..22 are blank filler; a click there must be a no-op.
	m.updateListMouse(mouseClick(0, 10))
	if m.view != viewList {
		t.Errorf("click on blank area should be ignored, view = %d", m.view)
	}
	if m.list.cursor != 0 {
		t.Errorf("cursor changed by blank click = %d, want 0", m.list.cursor)
	}
}

func TestWheelMovesListCursor(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseWheel(0, false)) // wheel up at cursor 0 wraps to last
	if m.list.cursor != 2 {
		t.Errorf("wheel up cursor = %d, want 2", m.list.cursor)
	}
	m.updateListMouse(mouseWheel(0, true)) // wheel down
	if m.list.cursor != 0 {
		t.Errorf("wheel down cursor = %d, want 0", m.list.cursor)
	}
}

func TestMouseIgnoredWhilePopupOpen(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{
		tags:   []store.TagCount{{Name: "tech", Total: 1, Unread: 1}},
		cursor: 0,
	}

	m.Update(mouseClick(0, 2))
	if m.view != viewList {
		t.Errorf("click during popup should not open article, view = %d", m.view)
	}
	if m.list.cursor != 0 {
		t.Errorf("click during popup changed cursor = %d", m.list.cursor)
	}
	if m.popup != popupTagsList {
		t.Errorf("click during popup should not close it, popup = %d", m.popup)
	}
}

func TestArticleWheelScrollsViewport(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "long", Content: "<p>x</p>"})
	m.article.viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	m.article.viewport.GotoTop()
	start := m.article.viewport.YOffset

	m.updateArticleMouse(mouseWheel(0, true)) // wheel down
	if m.article.viewport.YOffset != start+1 {
		t.Errorf("wheel down offset = %d, want %d", m.article.viewport.YOffset, start+1)
	}
	m.updateArticleMouse(mouseWheel(0, false)) // wheel up
	if m.article.viewport.YOffset != start {
		t.Errorf("wheel up offset = %d, want %d", m.article.viewport.YOffset, start)
	}
}

func TestArticleClickIsNoop(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "t", Content: "<p>x</p>"})
	start := m.article.viewport.YOffset

	m.updateArticleMouse(mouseClick(5, 5))
	if m.article.viewport.YOffset != start {
		t.Errorf("click in article view should be a no-op, offset = %d", m.article.viewport.YOffset)
	}
}
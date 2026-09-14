package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/store"
	"yerss/internal/ui/popup"
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

func rightClick(x, y int) tea.MouseMsg {
	return tea.MouseMsg{
		X:      x,
		Y:      y,
		Button: tea.MouseButtonRight,
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
	m.list.Load()
	if n := m.list.RowCount(); n != 3 {
		t.Fatalf("expected 3 visible rows, got %d", n)
	}
	ids := []int64{
		m.list.ArticleIDAtRow(1),
		m.list.ArticleIDAtRow(2),
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
	if m.article.ID() != ids[1] {
		t.Errorf("clicked article id = %d, want %d", m.article.ID(), ids[1])
	}
	if m.list.Cursor() != 2 {
		t.Errorf("cursor after click = %d, want 2", m.list.Cursor())
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Errorf("b should return to list, view = %d", m.view)
	}
	if m.list.Cursor() != 2 {
		t.Errorf("cursor not preserved on return = %d, want 2", m.list.Cursor())
	}
}

func TestClickOnDayHeaderToggles(t *testing.T) {
	m, _ := twoArticleModel(t)
	if m.list.Collapsed(0) {
		t.Fatal("group should start expanded")
	}

	m.updateListMouse(mouseClick(0, 0)) // row 0 = day header
	if m.view != viewList {
		t.Errorf("click on header should stay in list, view = %d", m.view)
	}
	if !m.list.Collapsed(0) {
		t.Error("click on header should collapse the day")
	}

	m.updateListMouse(mouseClick(0, 0))
	if m.list.Collapsed(0) {
		t.Error("second click on header should expand the day")
	}
}

func TestClickOnStatusBarIgnored(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseClick(0, m.height-1))
	if m.view != viewList {
		t.Errorf("click on status bar should be ignored, view = %d", m.view)
	}
	if m.list.Cursor() != 0 {
		t.Errorf("cursor changed by status-bar click = %d, want 0", m.list.Cursor())
	}
}

func TestClickBeyondRowsIgnored(t *testing.T) {
	m, _ := twoArticleModel(t)
	// Rows 3..22 are blank filler; a click there must be a no-op.
	m.updateListMouse(mouseClick(0, 10))
	if m.view != viewList {
		t.Errorf("click on blank area should be ignored, view = %d", m.view)
	}
	if m.list.Cursor() != 0 {
		t.Errorf("cursor changed by blank click = %d, want 0", m.list.Cursor())
	}
}

func TestWheelMovesListCursor(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseWheel(0, false)) // wheel up at cursor 0 wraps to last
	if m.list.Cursor() != 2 {
		t.Errorf("wheel up cursor = %d, want 2", m.list.Cursor())
	}
	m.updateListMouse(mouseWheel(0, true)) // wheel down
	if m.list.Cursor() != 0 {
		t.Errorf("wheel down cursor = %d, want 0", m.list.Cursor())
	}
}

func TestMouseIgnoredWhilePopupOpen(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.popupData.SetKind(popup.KindTagsList)
	m.popupData.SetTags([]store.TagCount{{Name: "tech", Total: 1, Unread: 1}})
	m.popupData.SetCursor(0)

	m.Update(mouseClick(0, 2))
	if m.view != viewList {
		t.Errorf("click during popup should not open article, view = %d", m.view)
	}
	if m.list.Cursor() != 0 {
		t.Errorf("click during popup changed cursor = %d", m.list.Cursor())
	}
	if m.popupData.Kind() != popup.KindTagsList {
		t.Errorf("click during popup should not close it, kind = %d", m.popupData.Kind())
	}
}

func TestArticleWheelScrollsViewport(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.article.Recompose(store.Article{Title: "long", Content: "<p>x</p>"})
	m.article.Viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	m.article.Viewport.GotoTop()
	start := m.article.Viewport.YOffset

	m.updateArticleMouse(mouseWheel(0, true)) // wheel down
	if m.article.Viewport.YOffset != start+1 {
		t.Errorf("wheel down offset = %d, want %d", m.article.Viewport.YOffset, start+1)
	}
	m.updateArticleMouse(mouseWheel(0, false)) // wheel up
	if m.article.Viewport.YOffset != start {
		t.Errorf("wheel up offset = %d, want %d", m.article.Viewport.YOffset, start)
	}
}

func TestRightClickReturnsToList(t *testing.T) {
	m, ids := twoArticleModel(t)
	m.updateListMouse(mouseClick(0, 1)) // open article a
	if m.view != viewArticle || m.article.ID() != ids[0] {
		t.Fatalf("expected article a open, view = %d id = %d", m.view, m.article.ID())
	}

	m.updateArticleMouse(rightClick(10, 5))
	if m.view != viewList {
		t.Errorf("right-click should return to the list, view = %d", m.view)
	}
	if m.list.Cursor() != 1 {
		t.Errorf("cursor after right-click = %d, want 1 (the article row)", m.list.Cursor())
	}
}

func TestRightClickDoesNotStartSelection(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseClick(0, 1))
	m.updateArticleMouse(rightClick(10, 5))
	if m.article.Sel.Active || m.article.Sel.Tracking {
		t.Errorf("right-click should not anchor a selection: %+v", m.article.Sel)
	}
}

func TestRightClickWithModifierIgnored(t *testing.T) {
	m, _ := twoArticleModel(t)
	m.updateListMouse(mouseClick(0, 1))
	msg := rightClick(10, 5)
	msg.Ctrl = true
	m.updateArticleMouse(msg)
	if m.view != viewArticle {
		t.Errorf("modifier right-click should be ignored, view = %d", m.view)
	}
}

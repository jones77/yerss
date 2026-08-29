package ui

import (
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/store"
)

// mouseDrag is a motion event while the left button is held (cell-motion mode
// reports motion only during a drag; the button field may be none).
func mouseDrag(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonNone, Action: tea.MouseActionMotion}
}

// mouseRelease is a left-button release event.
func mouseRelease(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonNone, Action: tea.MouseActionRelease}
}

// runCmd executes a tea.Cmd synchronously and returns its message.
func runCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	return cmd()
}

// clipboardMsgLabel runs a clipboard cmd and returns the success label, or ""
// when the platform helper is unavailable or the copy failed.
func clipboardMsgLabel(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	msg := runCmd(t, cmd)
	if msg == nil {
		return ""
	}
	am, ok := msg.(urlActionMsg)
	if !ok || am.err != nil {
		return ""
	}
	return am.label
}

func articleMouseModel(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "t", Content: "<p>x</p>"})
	return m
}

func TestArticlePressAnchorsSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(10, 10)) // content cell (7, 8)
	sel := m.article.sel
	if !sel.active || !sel.tracking {
		t.Fatalf("press should anchor an active tracked selection, got %+v", sel)
	}
	if sel.anchorX != 7 || sel.anchorY != 8 || sel.curX != 7 || sel.curY != 8 {
		t.Errorf("press anchor = (%d,%d) cur = (%d,%d), want (7,8)/(7,8)", sel.anchorX, sel.anchorY, sel.curX, sel.curY)
	}
}

func TestArticlePressOutsideContentIgnored(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(0, 0)) // top border
	if m.article.sel.active {
		t.Error("press on the top border should not start a selection")
	}
	m.updateArticleMouse(mouseClick(1, 2)) // left padding column
	if m.article.sel.active {
		t.Error("press in the left padding should not start a selection")
	}
	m.updateArticleMouse(mouseClick(5, 1)) // top padding row
	if m.article.sel.active {
		t.Error("press in the top padding should not start a selection")
	}
	m.updateArticleMouse(mouseClick(79, 23)) // bottom/right border area
	if m.article.sel.active {
		t.Error("press on the border should not start a selection")
	}
}

func TestArticleDragExtendsSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(3, 2)) // content cell (0, 0)
	m.updateArticleMouse(mouseDrag(30, 10))
	sel := m.article.sel
	if sel.anchorX != 0 || sel.anchorY != 0 {
		t.Errorf("drag must keep anchor at (0,0), got (%d,%d)", sel.anchorX, sel.anchorY)
	}
	if sel.curX != 27 || sel.curY != 8 {
		t.Errorf("drag current = (%d,%d), want (27,8)", sel.curX, sel.curY)
	}
}

func TestArticleDragClampsToContentArea(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(3, 2))
	m.updateArticleMouse(mouseDrag(1000, 1000)) // far outside content
	sel := m.article.sel
	if sel.curX != 73 || sel.curY != 19 {
		t.Errorf("drag outside content = (%d,%d), want clamped (73,19)", sel.curX, sel.curY)
	}
	m.updateArticleMouse(mouseDrag(-100, -100))
	sel = m.article.sel
	if sel.curX != 0 || sel.curY != 0 {
		t.Errorf("drag above content = (%d,%d), want clamped (0,0)", sel.curX, sel.curY)
	}
}

func TestArticleReleaseCopiesSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.article.viewport.SetContent("abcdefghij")
	m.article.lines = []string{"abcdefghij"}
	m.updateArticleMouse(mouseClick(3, 2)) // anchor (0,0)
	m.updateArticleMouse(mouseDrag(8, 2))  // current (5,0)
	cmd := m.updateArticleMouse(mouseRelease(8, 2))
	if cmd == nil {
		t.Fatal("release after a drag should produce a copy command")
	}
	if m.article.sel.active || m.article.sel.tracking {
		t.Error("released selection should be cleared after copying")
	}
	if runtime.GOOS == "darwin" {
		if label := clipboardMsgLabel(t, cmd); label != "copied selection" {
			t.Errorf("copy label = %q, want %q", label, "copied selection")
		}
	}
}

func TestArticleModifierPressIsNoop(t *testing.T) {
	m := articleMouseModel(t)
	ctrl := mouseClick(10, 10)
	ctrl.Ctrl = true
	m.updateArticleMouse(ctrl)
	if m.article.sel.active {
		t.Error("ctrl-click should not start a selection (OSC 8 passthrough)")
	}
	shift := mouseClick(10, 10)
	shift.Shift = true
	m.updateArticleMouse(shift)
	if m.article.sel.active {
		t.Error("shift-click should not start a selection")
	}
}

func TestArticleZeroWidthReleaseCopiesNothing(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(5, 5))
	cmd := m.updateArticleMouse(mouseRelease(5, 5))
	if cmd != nil {
		t.Error("zero-width press+release should copy nothing")
	}
}

func TestSelectionClearedOnBack(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(5, 5))
	m.updateArticleMouse(mouseDrag(30, 10))
	if !m.article.sel.active {
		t.Fatal("selection should be active before leaving")
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Fatalf("b should return to list, view = %d", m.view)
	}
	if m.article.sel.active {
		t.Error("selection should be cleared when leaving the article")
	}
}

func TestSelectionHighlightRendering(t *testing.T) {
	m := articleMouseModel(t)
	m.article.viewport.SetContent("abcdefghijklmnopqrstuvwxyz")
	m.article.lines = []string{"abcdefghijklmnopqrstuvwxyz"}
	m.article.sel = textSelection{active: true, anchorX: 0, anchorY: 0, curX: 6, curY: 0}

	render := m.renderArticle()
	if !strings.Contains(render, "\x1b[7mabcdef\x1b[0m") {
		t.Errorf("render missing inverted highlight over 'abcdef': %q", render)
	}
	if got := strings.Count(render, "\n") + 1; got != m.height {
		t.Errorf("highlight changed line count: %d, want %d", got, m.height)
	}

	m.ascii = true
	render = m.renderArticle()
	if !strings.Contains(render, "\x1b[7mabcdef\x1b[0m") {
		t.Errorf("ascii render missing inverted highlight: %q", render)
	}
}

func TestSelectedTextStripsEscapes(t *testing.T) {
	m := articleMouseModel(t)
	osc := ansi.SetHyperlink("https://example.com") + "[world]" + ansi.ResetHyperlink() + "(https://example.com)"
	m.article.viewport.SetContent(osc + "\nsecond line")
	m.article.lines = strings.Split(osc+"\nsecond line", "\n")

	m.article.sel = textSelection{active: true, anchorX: 0, anchorY: 0, curX: 1, curY: 1}
	got := m.article.selectedText()
	want := "[world](https://example.com)\ns"
	if got != want {
		t.Errorf("selectedText = %q, want %q", got, want)
	}
	if strings.Contains(got, "\x1b]8;") {
		t.Errorf("selectedText leaked OSC 8 escapes: %q", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("selectedText leaked ANSI escapes: %q", got)
	}
}

func TestSelectedTextZeroWidth(t *testing.T) {
	m := articleMouseModel(t)
	m.article.sel = textSelection{active: true, anchorX: 3, anchorY: 2, curX: 3, curY: 2}
	if got := m.article.selectedText(); got != "" {
		t.Errorf("zero-width selection text = %q, want empty", got)
	}
	m.article.sel = textSelection{}
	if got := m.article.selectedText(); got != "" {
		t.Errorf("inactive selection text = %q, want empty", got)
	}
}

func TestCKeyCopiesArticleText(t *testing.T) {
	m, _ := newTestModel(t)
	a := store.Article{Title: "headline", Link: "https://example.com/a", Content: "<p>body text</p>"}
	m.article = m.newArticleState(a)

	_, cmd := m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	if cmd == nil {
		t.Fatal("C should produce a copy command in the article view")
	}
	want := ansi.Strip(renderArticleContent(a))
	if !strings.Contains(want, "headline") || !strings.Contains(want, "body text") {
		t.Errorf("article text source missing content: %q", want)
	}
	if strings.Contains(want, "\x1b]8;") {
		t.Errorf("article text source leaked OSC 8: %q", want)
	}
	if runtime.GOOS == "darwin" {
		if label := clipboardMsgLabel(t, cmd); label != "copied article text" {
			t.Errorf("copy label = %q, want %q", label, "copied article text")
		}
	}

	// Lowercase c must still map to copy_url.
	cfg := config.Default()
	articleEff, err := cfg.Keybindings.EffectiveKeys(config.ViewArticle)
	if err != nil {
		t.Fatal(err)
	}
	if articleEff["c"] != config.CopyURL {
		t.Errorf("c should map to copy_url, got %v", articleEff["c"])
	}
}

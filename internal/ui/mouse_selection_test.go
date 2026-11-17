package ui

import (
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui/articleview"
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
	am, ok := msg.(articleview.URLActionMsg)
	if !ok || am.Err != nil {
		return ""
	}
	return am.Label
}

func articleMouseModel(t *testing.T) *Model {
	t.Helper()
	m, _ := newTestModel(t)
	m.article = m.article.Recompose(store.Article{Title: "t", Content: "<p>x</p>"})
	return m
}

func TestArticlePressAnchorsSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(10, 9)) // content cell (7, 7)
	sel := m.article.Sel
	if !sel.Active || !sel.Tracking {
		t.Fatalf("press should anchor an active tracked selection, got %+v", sel)
	}
	if sel.AnchorX != 7 || sel.AnchorY != 7 || sel.CurX != 7 || sel.CurY != 7 {
		t.Errorf("press anchor = (%d,%d) cur = (%d,%d), want (7,7)/(7,7)", sel.AnchorX, sel.AnchorY, sel.CurX, sel.CurY)
	}
}

func TestArticlePressOutsideContentIgnored(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(0, 0)) // top border
	if m.article.Sel.Active {
		t.Error("press on the top border should not start a selection")
	}
	m.updateArticleMouse(mouseClick(1, 2)) // left padding column
	if m.article.Sel.Active {
		t.Error("press in the left padding should not start a selection")
	}
	m.updateArticleMouse(mouseClick(79, 23)) // bottom/right border area
	if m.article.Sel.Active {
		t.Error("press on the border should not start a selection")
	}
}

func TestArticleDragExtendsSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(3, 2)) // content cell (0, 0)
	m.updateArticleMouse(mouseDrag(30, 9))
	sel := m.article.Sel
	if sel.AnchorX != 0 || sel.AnchorY != 0 {
		t.Errorf("drag must keep anchor at (0,0), got (%d,%d)", sel.AnchorX, sel.AnchorY)
	}
	if sel.CurX != 27 || sel.CurY != 7 {
		t.Errorf("drag current = (%d,%d), want (27,7)", sel.CurX, sel.CurY)
	}
}

func TestArticleDragClampsToContentArea(t *testing.T) {
	m := articleMouseModel(t)
	m.updateArticleMouse(mouseClick(3, 2))
	m.updateArticleMouse(mouseDrag(1000, 1000)) // far outside content
	sel := m.article.Sel
	if sel.CurX != 73 || sel.CurY != 19 {
		t.Errorf("drag outside content = (%d,%d), want clamped (73,19)", sel.CurX, sel.CurY)
	}
	m.updateArticleMouse(mouseDrag(-100, -100))
	sel = m.article.Sel
	if sel.CurX != 0 || sel.CurY != 0 {
		t.Errorf("drag above content = (%d,%d), want clamped (0,0)", sel.CurX, sel.CurY)
	}
}

func TestArticleReleaseCopiesSelection(t *testing.T) {
	m := articleMouseModel(t)
	m.article.Viewport.SetContent("abcdefghij")
	m.article.SetLines([]string{"abcdefghij"})
	m.updateArticleMouse(mouseClick(3, 2)) // anchor (0,0)
	m.updateArticleMouse(mouseDrag(8, 2))  // current (5,0)
	cmd := m.updateArticleMouse(mouseRelease(8, 2))
	if cmd == nil {
		t.Fatal("release after a drag should produce a copy command")
	}
	if m.article.Sel.Active || m.article.Sel.Tracking {
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
	if m.article.Sel.Active {
		t.Error("ctrl-click should not start a selection (OSC 8 passthrough)")
	}
	shift := mouseClick(10, 10)
	shift.Shift = true
	m.updateArticleMouse(shift)
	if m.article.Sel.Active {
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
	if !m.article.Sel.Active {
		t.Fatal("selection should be active before leaving")
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.view != viewList {
		t.Fatalf("b should return to list, view = %d", m.view)
	}
	if m.article.Sel.Active {
		t.Error("selection should be cleared when leaving the article")
	}
}

func TestSelectionHighlightRendering(t *testing.T) {
	m := articleMouseModel(t)
	m.article.Viewport.SetContent("abcdefghijklmnopqrstuvwxyz")
	m.article.SetLines([]string{"abcdefghijklmnopqrstuvwxyz"})
	m.article.Sel = articleview.Selection{Active: true, AnchorX: 0, AnchorY: 0, CurX: 6, CurY: 0}

	render := m.article.Render()
	if !strings.Contains(render, "\x1b[7mabcdef\x1b[0m") {
		t.Errorf("render missing inverted highlight over 'abcdef': %q", render)
	}
	if got := strings.Count(render, "\n") + 1; got != m.height {
		t.Errorf("highlight changed line count: %d, want %d", got, m.height)
	}

	m.SetAscii(true)
	render = m.article.Render()
	if !strings.Contains(render, "\x1b[7mabcdef\x1b[0m") {
		t.Errorf("ascii render missing inverted highlight: %q", render)
	}
}

func TestSelectionHighlightSpansLink(t *testing.T) {
	// A selection crossing the link's SGR reset must re-assert reverse video
	// instead of letting the link's reset end the highlight early.
	line := "\x1b[37mbefore \x1b[m\x1b]8;;https://example.com/x\x07\x1b[96;1mlink\x1b[m\x1b]8;;\x07\x1b[37m after\x1b[m"
	sel := articleview.Selection{Active: true, AnchorX: 3, AnchorY: 0, CurX: 13, CurY: 0}
	out := articleview.HighlightSelection([]string{line}, sel)
	if len(out) != 1 {
		t.Fatalf("highlightSelection returned %d lines, want 1", len(out))
	}
	got := out[0]
	if ansi.Strip(got) != ansi.Strip(line) {
		t.Fatalf("highlight changed the visible text:\n got %q\nwant %q", ansi.Strip(got), ansi.Strip(line))
	}
	if n := strings.Count(got, "\x1b[7m"); n < 2 {
		t.Errorf("selection across the link has %d reverse-on sequences, want at least 2 (re-asserted after the link reset): %q", n, got)
	}
	if !strings.Contains(got, "\x1b]8;;https://example.com/x") {
		t.Errorf("link open dropped inside the highlight: %q", got)
	}
}

func TestSelectedTextStripsEscapes(t *testing.T) {
	m := articleMouseModel(t)
	osc := ansi.SetHyperlink("https://example.com") + "[world]" + ansi.ResetHyperlink() + "(https://example.com)"
	m.article.Viewport.SetContent(osc + "\nsecond line")
	m.article.SetLines(strings.Split(osc+"\nsecond line", "\n"))

	m.article.Sel = articleview.Selection{Active: true, AnchorX: 0, AnchorY: 0, CurX: 1, CurY: 1}
	got := m.article.SelectedText()
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
	m.article.Sel = articleview.Selection{Active: true, AnchorX: 3, AnchorY: 2, CurX: 3, CurY: 2}
	if got := m.article.SelectedText(); got != "" {
		t.Errorf("zero-width selection text = %q, want empty", got)
	}
	m.article.Sel = articleview.Selection{}
	if got := m.article.SelectedText(); got != "" {
		t.Errorf("inactive selection text = %q, want empty", got)
	}
}

func TestCKeyCopiesArticleText(t *testing.T) {
	m, _ := newTestModel(t)
	a := store.Article{Title: "headline", Link: "https://example.com/a", Content: "<p>body text</p>"}
	m.article = m.article.Recompose(a)

	_, cmd := m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	if cmd == nil {
		t.Fatal("C should produce a copy command in the article view")
	}
	want := ansi.Strip(strings.Join(m.article.Lines(), "\n"))
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

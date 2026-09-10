package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
)

func (m *Model) updateArticle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	act, ok := m.keys[config.ViewArticle][msg.String()]
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	switch act {
	case config.Quit:
		m.persistSelection()
		return m, tea.Quit
	case config.Back, config.CloseArticle:
		m.backToList()
	case config.LinkPopup:
		m.openLinksPopup()
	case config.MoveDown:
		cmd = m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
		m.article.markRead(m.sess.Store())
	case config.MoveUp:
		cmd = m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	case config.PageDown:
		cmd = m.scrollArticle(func() { m.article.viewport.PageDown() }, false)
		m.article.markRead(m.sess.Store())
	case config.PageUp:
		cmd = m.scrollArticle(func() { m.article.viewport.PageUp() }, false)
	case config.HalfPageDown:
		cmd = m.scrollArticle(func() { m.article.viewport.HalfPageDown() }, false)
		m.article.markRead(m.sess.Store())
	case config.HalfPageUp:
		cmd = m.scrollArticle(func() { m.article.viewport.HalfPageUp() }, false)
	case config.Top:
		cmd = m.scrollArticle(func() { m.article.viewport.GotoTop() }, false)
	case config.Bottom:
		cmd = m.scrollArticle(func() { m.article.viewport.GotoBottom() }, false)
	case config.OpenURL:
		if m.article.article.Link != "" {
			cmd = openURLCmd(m.article.article.Link)
		}
	case config.CopyURL:
		if m.article.article.Link != "" {
			cmd = copyURLCmd(m.article.article.Link)
		}
	case config.CopyArticleText:
		cmd = copyTextCmd(ansi.Strip(strings.Join(m.article.lines, "\n")), "copied article text")
	case config.TagPopup:
		m.openTagPopup(popupTagsArticle)
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

// updateArticleMouse handles mouse events in the article view: wheel up/down
// scrolls the viewport by one line, with downward scroll marking the article
// read like the Down key. An unmodified left-button press anchors a text
// selection, drag extends it, and release copies the selected text to the
// clipboard and clears the selection. An unmodified right-button press
// returns to the list view like the Back keys. Presses carrying a Ctrl/Alt/
// Shift modifier are not interpreted, so the terminal can activate OSC 8
// hyperlinks natively.
func (m *Model) updateArticleMouse(msg tea.MouseMsg) tea.Cmd {
	if tea.MouseEvent(msg).IsWheel() && msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			return m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
		case tea.MouseButtonWheelDown:
			cmd := m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
			m.article.markRead(m.sess.Store())
			return cmd
		}
		return nil
	}
	if msg.Shift || msg.Alt || msg.Ctrl {
		return nil
	}
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonRight {
			m.backToList()
			return nil
		}
		if msg.Button != tea.MouseButtonLeft {
			return nil
		}
		if c, ok := m.contentCell(msg.X, msg.Y); ok {
			if url := m.linkAtContentCell(c.x, c.y); url != "" {
				return openURLCmd(url)
			}
			m.article.sel = textSelection{
				active:   true,
				tracking: true,
				anchorX:  c.x,
				anchorY:  c.y,
				curX:     c.x,
				curY:     c.y,
			}
		}
	case tea.MouseActionMotion:
		if !m.article.sel.tracking {
			return nil
		}
		c := m.contentCellClamped(msg.X, msg.Y)
		m.article.sel.curX, m.article.sel.curY = c.x, c.y
	case tea.MouseActionRelease:
		if !m.article.sel.tracking {
			return nil
		}
		text := m.article.selectedText()
		m.article.sel = textSelection{}
		if text != "" {
			return copyTextCmd(text, "copied selection")
		}
	}
	return nil
}

// backToList returns from the article view to the list view: the previously
// viewed article stays selected, any active mouse selection is cleared, and the
// deferred native-render set is dropped so a reopen starts fresh (the photos
// remain cached, so in-view renders re-fire through the open path).
func (m *Model) backToList() {
	m.view = viewList
	m.article.sel = textSelection{}
	m.nativePending = make(map[string]bool)
	m.loadList()
}

// contentRect returns the screen-space rectangle of the article content area
// (origin column/row plus width/height in cells), matching renderArticleBorder:
// content begins at column padX+1 and row 1+effPadY, beneath the top border and
// its vertical margin.

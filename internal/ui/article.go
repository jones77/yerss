package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/store"
)

type articleState struct {
	id         int64
	article    *store.Article
	lines      []string
	viewport   viewport.Model
	readMarked bool
}

func (m *Model) openArticle() {
	rows := m.visibleRows()
	if m.list.cursor < 0 || m.list.cursor >= len(rows) {
		return
	}
	row := rows[m.list.cursor]
	if row.kind != rowArticle {
		return
	}
	item := &m.list.groups[row.groupIdx].articles[row.artIdx]
	full, err := m.store.GetArticle(item.ID)
	if err == nil && full != nil {
		m.article = m.newArticleState(*full)
	} else {
		m.article = m.newArticleState(store.Article{ID: item.ID, Title: item.Title, Read: item.Read})
	}
	m.view = viewArticle
}

func (m *Model) newArticleState(a store.Article) articleState {
	padX, padY := m.cfg.Display.PaddingX, m.cfg.Display.PaddingY
	contentW := m.width - 2*padX - 2
	if contentW < 1 {
		contentW = 1
	}
	vpH := m.height - 2*padY - 2
	if vpH < 1 {
		vpH = 1
	}
	text := renderArticleContent(a)
	wrapped := wrapText(text, contentW, glyphsFor(m.ascii).ellipsis)
	vp := viewport.New(contentW, vpH)
	vp.SetContent(wrapped)
	st := articleState{
		id:       a.ID,
		article:  &a,
		lines:    strings.Split(wrapped, "\n"),
		viewport: vp,
	}
	if len(st.lines) <= vpH {
		st.markRead(m.store)
	}
	return st
}

func renderArticleContent(a store.Article) string {
	var b strings.Builder
	if a.Title != "" {
		b.WriteString(a.Title)
		b.WriteString("\n")
	}
	if a.Author != "" {
		b.WriteString("by " + a.Author)
		b.WriteString("\n")
	}
	if a.Link != "" {
		b.WriteString(ansi.SetHyperlink(a.Link) + a.Link + ansi.ResetHyperlink())
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(HTMLToText(a.Content))
	return b.String()
}

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
	case config.Back:
		m.view = viewList
		m.loadList()
	case config.MoveDown:
		m.article.viewport.ScrollDown(1)
		m.article.markRead(m.store)
	case config.MoveUp:
		m.article.viewport.ScrollUp(1)
	case config.PageDown:
		m.article.viewport.PageDown()
		m.article.markRead(m.store)
	case config.PageUp:
		m.article.viewport.PageUp()
	case config.HalfPageDown:
		m.article.viewport.HalfPageDown()
		m.article.markRead(m.store)
	case config.HalfPageUp:
		m.article.viewport.HalfPageUp()
	case config.Top:
		m.article.viewport.GotoTop()
	case config.Bottom:
		m.article.viewport.GotoBottom()
	case config.ToggleRead:
		m.article.toggleRead(m.store)
	case config.OpenURL:
		if m.article.article.Link != "" {
			cmd = openURLCmd(m.article.article.Link)
		}
	case config.CopyURL:
		if m.article.article.Link != "" {
			cmd = copyURLCmd(m.article.article.Link)
		}
	case config.TagPopup:
		m.openTagPopup(popupTagsArticle)
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

// updateArticleMouse handles mouse events in the article view: wheel up/down
// scrolls the viewport by one line, with downward scroll marking the article
// read like the Down key. Clicks are no-ops — links are handled natively by
// the terminal via OSC 8 hyperlinks.
func (m *Model) updateArticleMouse(msg tea.MouseMsg) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		m.article.viewport.ScrollUp(1)
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		m.article.viewport.ScrollDown(1)
		m.article.markRead(m.store)
	}
}

// markRead implements the read-status rules: an article that fits the
// viewport is marked read on open; a longer article is marked read on the
// first downward scroll.
func (s *articleState) markRead(st *store.Store) {
	if s.readMarked || s.article.Read {
		s.readMarked = true
		return
	}
	s.readMarked = true
	s.article.Read = true
	_ = st.SetRead(s.id, true)
}

func (s *articleState) toggleRead(st *store.Store) {
	s.article.Read = !s.article.Read
	_ = st.SetRead(s.id, s.article.Read)
}

func (m *Model) renderArticle() string {
	st := &m.article
	a := st.article
	date := ""
	if !a.PublishedAt.IsZero() {
		date = a.PublishedAt.Local().Format("2006-01-02 15:04:05")
	}
	title := a.Title
	if title == "" {
		title = "(untitled)"
	}
	g := glyphsFor(m.ascii)
	content := m.article.viewport.View()
	lines := strings.Split(content, "\n")
	sc := scrollState{
		totalH:    m.article.viewport.TotalLineCount(),
		viewportH: m.article.viewport.Height,
		offset:    m.article.viewport.YOffset,
	}
	return renderArticleBorder(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY,
		g, m.palette, date, title, sc, lines)
}
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal"
	"yerss/internal/config"
	"yerss/internal/ui/articleview"
	"yerss/internal/ui/popup"
)

func (m *Model) openArticle() tea.Cmd {
	item, ok := m.list.ArticleAtCursor()
	if !ok {
		return nil
	}
	full, err := m.sess.GetArticle(item.ID)
	if err == nil && full != nil {
		m.article = m.article.Recompose(*full)
		m.view = viewArticle
		return m.images.FireImageLoad(*full, &m.article)
	}
	m.article = m.article.Recompose(internal.Article{ID: item.ID, Title: item.Title, Read: item.Read})
	m.view = viewArticle
	return nil
}

func (m *Model) updateArticle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	act, cmd := m.article.UpdateKey(msg, m.keys[config.ViewArticle])
	switch act {
	case articleview.ActionQuit:
		m.persistSelection()
		return m, tea.Quit
	case articleview.ActionBack:
		m.backToList()
	case articleview.ActionOpenLinkPopup:
		m.openLinksPopup()
	case articleview.ActionOpenTagPopup:
		m.openTagPopup(popup.KindTagsArticle)
	case articleview.ActionToggleLinkExpand:
		m.toggleLinkExpand()
	case articleview.ActionHelp:
		m.popupData.SetKind(popup.KindHelp)
	}
	return m, cmd
}

func (m *Model) updateArticleMouse(msg tea.MouseMsg) tea.Cmd {
	act, cmd := m.article.UpdateMouse(msg)
	if act == articleview.ActionBack {
		m.backToList()
	}
	return cmd
}

func (m *Model) backToList() {
	m.view = viewList
	m.article.Sel = articleview.Selection{}
	m.images.ClearNativePending()
	if status := m.list.Load(); status != "" {
		m.setStatus(status)
	}
}

func (m *Model) toggleLinkExpand() {
	m.linkExpand = m.article.ToggleLinkExpand()
	m.article.LinkExpand = m.linkExpand
}

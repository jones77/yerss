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
	return m.openArticleByID(item.ID)
}

// openArticleByID loads the article with the given ID and switches the reader to
// it, resetting the transient navigation hint and recording the article's
// position within the shown list.
func (m *Model) openArticleByID(id int64) tea.Cmd {
	full, err := m.sess.GetArticle(id)
	if err == nil && full != nil {
		m.article = m.article.Recompose(*full)
		m.article.ResetNavHint()
		m.setArticlePosition(id)
		m.view = viewArticle
		return m.images.FireImageLoad(*full, &m.article)
	}
	m.article = m.article.Recompose(internal.Article{ID: id})
	m.article.ResetNavHint()
	m.setArticlePosition(id)
	m.view = viewArticle
	return nil
}

// setArticlePosition records the open article's 1-based position within the
// currently shown (possibly filtered) list for the reader's top-right counter.
// A zero position means the article is not in the list, so the counter is
// omitted.
func (m *Model) setArticlePosition(id int64) {
	ids := m.list.ArticleIDs()
	for i, aid := range ids {
		if aid == id {
			m.article.SetListPosition(i+1, len(ids))
			return
		}
	}
	m.article.SetListPosition(0, 0)
}

// navigateArticle opens the article one step away in the shown list order: step
// +1 is the older (previous) article down the newest-first list, and step -1 is
// the newer (next) article up it. It wraps at both ends and is a no-op when the
// current article is not in the list.
func (m *Model) navigateArticle(step int) tea.Cmd {
	ids := m.list.ArticleIDs()
	cur := -1
	for i, id := range ids {
		if id == m.article.ID() {
			cur = i
			break
		}
	}
	if cur < 0 {
		return nil
	}
	next := (cur + step + len(ids)) % len(ids)
	if next == cur {
		return nil
	}
	return m.openArticleByID(ids[next])
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
	case articleview.ActionPreviousArticle:
		cmd = m.navigateArticle(1)
	case articleview.ActionNextArticle:
		cmd = m.navigateArticle(-1)
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
	id := m.article.ID()
	m.view = viewList
	m.article.Sel = articleview.Selection{}
	m.images.ClearNativePending()
	if status := m.list.Load(); status != "" {
		m.setStatus(status)
	}
	m.list.SelectArticleID(id)
}

func (m *Model) toggleLinkExpand() {
	m.linkExpand = m.article.ToggleLinkExpand()
	m.article.LinkExpand = m.linkExpand
}

package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/store"
)

// persistSelection records the current list selection (and the open article,
// when in the article view) so a restart can restore it. It is called on quit.
func (m *Model) persistSelection() {
	sel := store.LastSelection{View: "list"}
	if m.view == viewArticle {
		sel.View = "article"
		if m.article.Article() != nil {
			sel.ArticleID = m.article.ID()
			sel.ArticleOffset = m.article.Viewport.YOffset
		}
	} else {
		sel.ArticleID, sel.HeaderKey = m.list.SelectionForCursor()
	}
	_ = m.sess.SaveSelection(sel)
}

// restoreSelection repositions the list cursor to the previously saved
// selection and, when the reader was in the article view, reopens that
// article. It is called once at startup after the list is loaded; a selection
// that no longer exists (article pruned, header gone) falls back to the clamped
// cursor. When an article is restored it returns the same lead-image load
// command openArticle fires, so a restarted session resolves the image from
// the cache hierarchy instead of staying imageless until re-opened.
func (m *Model) restoreSelection() tea.Cmd {
	sel, err := m.sess.LoadSelection()
	if err != nil {
		return nil
	}
	m.list.RestoreCursor(sel.ArticleID, sel.HeaderKey)
	if sel.View == "article" && sel.ArticleID != 0 {
		if full, err := m.sess.GetArticle(sel.ArticleID); err == nil && full != nil {
			m.article = m.article.Recompose(*full)
			m.article.Viewport.SetYOffset(sel.ArticleOffset)
			m.view = viewArticle
			return m.images.FireImageLoad(*full, &m.article)
		}
	}
	return nil
}
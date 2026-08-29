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
		if m.article.article != nil {
			sel.ArticleID = m.article.id
			sel.ArticleOffset = m.article.viewport.YOffset
		}
	} else if item, ok := m.articleAtCursor(); ok {
		sel.ArticleID = item.ID
	} else {
		rows := m.visibleRows()
		if m.list.cursor >= 0 && m.list.cursor < len(rows) && rows[m.list.cursor].kind == rowHeader {
			g := &m.list.groups[rows[m.list.cursor].groupIdx]
			sel.HeaderKey = dayKey(g.date)
		}
	}
	_ = m.store.SaveLastSelection(sel)
}

// restoreSelection repositions the list cursor to the previously saved
// selection and, when the reader was in the article view, reopens that
// article. It is called once at startup after the list is loaded; a selection
// that no longer exists (article pruned, header gone) falls back to the clamped
// cursor. When an article is restored it returns the same lead-image load
// command openArticle fires, so a restarted session resolves the image from
// the cache hierarchy instead of staying imageless until re-opened.
func (m *Model) restoreSelection() tea.Cmd {
	sel, err := m.store.LoadLastSelection()
	if err != nil {
		return nil
	}
	if sel.HeaderKey != "" {
		for gi := range m.list.groups {
			if dayKey(m.list.groups[gi].date) == sel.HeaderKey {
				rows := m.visibleRows()
				for i, r := range rows {
					if r.kind == rowHeader && r.groupIdx == gi {
						m.list.cursor = i
						break
					}
				}
				break
			}
		}
	} else if sel.ArticleID != 0 {
		rows := m.visibleRows()
		for i, r := range rows {
			if r.kind == rowArticle && m.list.groups[r.groupIdx].articles[r.artIdx].ID == sel.ArticleID {
				m.list.cursor = i
				break
			}
		}
	}
	if sel.View == "article" && sel.ArticleID != 0 {
		if full, err := m.store.GetArticle(sel.ArticleID); err == nil && full != nil {
			m.article = m.newArticleState(*full)
			m.article.viewport.SetYOffset(sel.ArticleOffset)
			m.view = viewArticle
			return m.fireImageLoad(*full)
		}
	}
	return nil
}
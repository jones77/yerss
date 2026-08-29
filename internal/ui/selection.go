package ui

import (
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
	} else {
		rows := m.visibleRows()
		if m.list.cursor >= 0 && m.list.cursor < len(rows) {
			r := rows[m.list.cursor]
			if r.kind == rowHeader {
				g := &m.list.groups[r.groupIdx]
				if g.date.IsZero() {
					sel.HeaderKey = "undated"
				} else {
					sel.HeaderKey = g.date.Format("20060102")
				}
			} else {
				sel.ArticleID = m.list.groups[r.groupIdx].articles[r.artIdx].ID
			}
		}
	}
	_ = m.store.SaveLastSelection(sel)
}

// restoreSelection repositions the list cursor to the previously saved
// selection and, when the reader was in the article view, reopens that
// article. It is called once at startup after the list is loaded; a selection
// that no longer exists (article pruned, header gone) falls back to the clamped
// cursor.
func (m *Model) restoreSelection() {
	sel, err := m.store.LoadLastSelection()
	if err != nil {
		return
	}
	if sel.HeaderKey != "" {
		for gi := range m.list.groups {
			g := &m.list.groups[gi]
			key := "undated"
			if !g.date.IsZero() {
				key = g.date.Format("20060102")
			}
			if key == sel.HeaderKey {
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
		}
	}
}
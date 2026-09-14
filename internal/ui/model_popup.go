package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
	"yerss/internal/ui/articleview"
	"yerss/internal/ui/popup"
)

// openTagPopup opens the tag popup over the current view, seeded from the
// session's tag counts. kind selects the popup variant: the list popup shows
// every tag, the article popup only the open article's own categories.
func (m *Model) openTagPopup(kind popup.Kind) {
	var categories []string
	if kind == popup.KindTagsArticle {
		categories = m.article.Article().Categories
	}
	if status := m.popupData.OpenTags(kind, categories); status != "" {
		m.setStatus(status)
	}
}

// openLinksPopup opens the links popup over the article view, seeded with the
// links harvested from the current article's markdown source.
func (m *Model) openLinksPopup() {
	links := make([]popup.Link, len(m.article.Links()))
	for i, l := range m.article.Links() {
		links[i] = popup.Link{Text: l.Text, URL: l.URL}
	}
	m.popupData.OpenLinks(links)
}

// updatePopup routes popup keys through the popup state and acts on the coarse
// actions it reports: confirming a selection, opening the article URL, closing
// the popup, or quitting.
func (m *Model) updatePopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.popupData.Update(msg, m.keys[config.ViewPopup]) {
	case popup.ActionQuit:
		m.persistSelection()
		return m, tea.Quit
	case popup.ActionBack:
		// The popup state closed itself; nothing more to do.
	case popup.ActionOpenURL:
		if (m.popupData.Kind() == popup.KindLinks || m.popupData.Kind() == popup.KindTagsArticle) && m.article.Article().Link != "" {
			return m, articleview.OpenURLCmd(m.article.Article().Link)
		}
	case popup.ActionConfirm:
		if m.popupData.Kind() == popup.KindLinks {
			return m.confirmLinkSelection()
		}
		return m.confirmTagSelection()
	}
	return m, nil
}

// confirmLinkSelection closes the links popup and opens the selected URL in
// the system browser. An empty popup (an article without links) just closes.
func (m *Model) confirmLinkSelection() (tea.Model, tea.Cmd) {
	url := m.popupData.SelectedURL()
	m.popupData.SetKind(popup.KindNone)
	if url == "" {
		return m, nil
	}
	return m, articleview.OpenURLCmd(url)
}

// confirmTagSelection closes the tag popup and filters the list to the
// selected tag, returning to the list view.
func (m *Model) confirmTagSelection() (tea.Model, tea.Cmd) {
	t, ok := m.popupData.SelectedTag()
	m.popupData.SetKind(popup.KindNone)
	if !ok {
		return m, nil
	}
	m.list.SetFilter(t.Name)
	m.view = viewList
	m.article.Sel = articleview.Selection{}
	if status := m.list.Load(); status != "" {
		m.setStatus(status)
	}
	return m, nil
}
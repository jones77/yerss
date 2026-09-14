package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
	"yerss/internal/ui/listview"
	"yerss/internal/ui/popup"
)

// renderList returns the list view frame.
func (m *Model) renderList() string {
	return m.list.Render()
}

// updateList routes list view keys to the list view and acts on the actions it
// reports back. List-internal movements are handled inside the view; the
// actions that cross view boundaries (open article, refresh, tag popup, help,
// quit) are performed here.
func (m *Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.list.Update(msg, m.keys[config.ViewList]) {
	case listview.ActionQuit:
		m.persistSelection()
		return m, tea.Quit
	case listview.ActionOpenArticle:
		return m, m.openArticle()
	case listview.ActionRefresh:
		return m, m.refreshManual()
	case listview.ActionTagPopup:
		m.openTagPopup(popup.KindTagsList)
	case listview.ActionHelp:
		m.popupData.SetKind(popup.KindHelp)
	}
	return m, nil
}

// updateListMouse routes list view mouse events to the list view and opens the
// article when a click selects one.
func (m *Model) updateListMouse(msg tea.MouseMsg) tea.Cmd {
	if m.list.UpdateMouse(msg) == listview.ActionOpenArticle {
		return m.openArticle()
	}
	return nil
}

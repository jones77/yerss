package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
)

func (m *Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.visibleRows()
	if m.list.cursor >= 0 && m.list.cursor < len(rows) {
		row := rows[m.list.cursor]
		if row.kind == rowHeader {
			switch msg.String() {
			case keyFoldIn:
				m.collapse(row.groupIdx)
				return m, nil
			case keyFoldOut:
				m.expand(row.groupIdx)
				return m, nil
			case keyFoldToggle, keyFoldToggleSpace, keyFoldToggleTab:
				m.toggle(row.groupIdx)
				return m, nil
			}
		} else if msg.String() == keyFoldToggleTab {
			m.toggle(row.groupIdx)
			return m, nil
		}
	}

	act, ok := m.keys[config.ViewList][msg.String()]
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	switch act {
	case config.Quit:
		m.persistSelection()
		return m, tea.Quit
	case config.MoveDown:
		m.moveListCursor(1)
	case config.MoveUp:
		m.moveListCursor(-1)
	case config.OpenArticle:
		cmd = m.openArticle()
	case config.Back:
		if m.list.filter != "" {
			m.clearFilter()
		}
	case config.Refresh:
		cmd = m.refreshManual()
	case config.PageDown:
		m.list.cursor = clampIndex(m.list.cursor+m.pageSize(), len(m.visibleRows()))
	case config.PageUp:
		m.list.cursor = clampIndex(m.list.cursor-m.pageSize(), len(m.visibleRows()))
	case config.HalfPageDown:
		m.list.cursor = clampIndex(m.list.cursor+m.pageSize()/2, len(m.visibleRows()))
	case config.HalfPageUp:
		m.list.cursor = clampIndex(m.list.cursor-m.pageSize()/2, len(m.visibleRows()))
	case config.Top:
		if n := len(m.visibleRows()); n > 0 {
			m.list.cursor = 0
		}
	case config.Bottom:
		if n := len(m.visibleRows()); n > 0 {
			m.list.cursor = n - 1
		}
	case config.TagPopup:
		m.openTagPopup(popupTagsList)
	case config.ExpandToggle:
		m.expandToggle()
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

// updateListMouse handles mouse events in the list view: wheel moves the
// cursor by one row, a left-click on an article row selects and opens it, and
// a left-click on a day-header row toggles that group's fold. Clicks on the
// status bar (the bottom line) are ignored.
func (m *Model) updateListMouse(msg tea.MouseMsg) tea.Cmd {
	switch {
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		m.moveListCursor(-1)
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		m.moveListCursor(1)
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		return m.handleListClick(msg.Y)
	}
	return nil
}

// handleListClick maps a screen row to a visible-row index via the list window
// start offset and dispatches on the row kind found there. It returns the open
// article's image-load command when an article row is clicked.
func (m *Model) handleListClick(y int) tea.Cmd {
	if y >= m.height-1 {
		return nil
	}
	rows := m.visibleRows()
	if len(rows) == 0 {
		return nil
	}
	start, _ := listWindow(len(rows), m.list.cursor, m.pageSize())
	idx := start + y
	if idx < 0 || idx >= len(rows) {
		return nil
	}
	if rows[idx].kind == rowHeader {
		m.toggle(rows[idx].groupIdx)
		return nil
	}
	m.list.cursor = idx
	return m.openArticle()
}

func (m *Model) pageSize() int {
	n := m.height - 1
	if n < 1 {
		n = 1
	}
	return n
}

func (m *Model) moveListCursor(delta int) {
	n := len(m.visibleRows())
	if n == 0 {
		return
	}
	m.list.cursor = wrapIndex(m.list.cursor, delta, n)
}

// articleAtCursor returns the article item under the list cursor, or false when
// the cursor is out of range or on a day-header row.
func (m *Model) articleAtCursor() (*articleItem, bool) {
	rows := m.visibleRows()
	if m.list.cursor < 0 || m.list.cursor >= len(rows) {
		return nil, false
	}
	row := rows[m.list.cursor]
	if row.kind != rowArticle {
		return nil, false
	}
	return &m.list.groups[row.groupIdx].articles[row.artIdx], true
}

func (m *Model) clearFilter() {
	m.list.filter = ""
	m.loadList()
}

// listWindow returns the slice of the list to render around the cursor.
func listWindow(n, cursor, height int) (int, int) {
	if n <= height {
		return 0, n
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > n {
		end = n
		start = end - height
	}
	return start, end
}

package listview

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
	"yerss/internal/timeutil"
	"yerss/internal/ui/render"
)

// Action is a signal the list view returns from its input handlers for the
// Model to act on. List-internal movements (cursor, fold, filter) are handled
// inside the view; actions that reach across view boundaries are reported back.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionOpenArticle
	ActionRefresh
	ActionTagPopup
	ActionHelp
)

// Update handles a key message for the list view and returns the Action the
// Model must perform (open the article, refresh, open the tag popup, etc.).
func (s *State) Update(msg tea.KeyMsg, keys map[string]config.Action) Action {
	rows := s.visibleRows()
	if s.cursor >= 0 && s.cursor < len(rows) {
		row := rows[s.cursor]
		if row.kind == rowHeader {
			switch msg.String() {
			case keyFoldIn:
				s.collapse(row.groupIdx)
				return ActionNone
			case keyFoldOut:
				s.expand(row.groupIdx)
				return ActionNone
			case keyFoldToggle, keyFoldToggleTab:
				s.toggle(row.groupIdx)
				return ActionNone
			}
		} else if msg.String() == keyFoldToggleTab {
			s.toggle(row.groupIdx)
			return ActionNone
		}
	}

	act, ok := keys[msg.String()]
	if !ok {
		return ActionNone
	}
	switch act {
	case config.Quit:
		return ActionQuit
	case config.MoveDown:
		s.moveListCursor(1)
	case config.MoveUp:
		s.moveListCursor(-1)
	case config.OpenArticle:
		return ActionOpenArticle
	case config.Back:
		if s.filter != "" {
			if status := s.ClearFilter(); status != "" {
				s.SetStatus(status)
			}
		}
	case config.Refresh:
		return ActionRefresh
	case config.PageDown:
		s.cursor = render.ClampIndex(s.cursor+s.pageSize(), len(s.visibleRows()))
	case config.PageUp:
		s.cursor = render.ClampIndex(s.cursor-s.pageSize(), len(s.visibleRows()))
	case config.HalfPageDown:
		s.cursor = render.ClampIndex(s.cursor+s.pageSize()/2, len(s.visibleRows()))
	case config.HalfPageUp:
		s.cursor = render.ClampIndex(s.cursor-s.pageSize()/2, len(s.visibleRows()))
	case config.Top:
		if n := len(s.visibleRows()); n > 0 {
			s.cursor = 0
		}
	case config.Bottom:
		if n := len(s.visibleRows()); n > 0 {
			s.cursor = n - 1
		}
	case config.TagPopup:
		return ActionTagPopup
	case config.ExpandToggle:
		s.expandToggle()
	case config.Help:
		return ActionHelp
	}
	return ActionNone
}

// UpdateMouse handles mouse events in the list view: wheel moves the cursor by
// one row, a left-click on an article row selects and opens it, and a
// left-click on a day-header row toggles that group's fold. Clicks on the
// status bar (the bottom line) are ignored.
func (s *State) UpdateMouse(msg tea.MouseMsg) Action {
	switch {
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		s.moveListCursor(-1)
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		s.moveListCursor(1)
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		return s.handleListClick(msg.Y)
	}
	return ActionNone
}

// handleListClick maps a screen row to a visible-row index via the list window
// start offset and dispatches on the row kind found there. It reports
// ActionOpenArticle when an article row is clicked.
func (s *State) handleListClick(y int) Action {
	if y >= s.height-1 {
		return ActionNone
	}
	rows := s.visibleRows()
	if len(rows) == 0 {
		return ActionNone
	}
	start, _ := listWindow(len(rows), s.cursor, s.pageSize())
	idx := start + y
	if idx < 0 || idx >= len(rows) {
		return ActionNone
	}
	if rows[idx].kind == rowHeader {
		s.toggle(rows[idx].groupIdx)
		return ActionNone
	}
	s.cursor = idx
	return ActionOpenArticle
}

func (s *State) pageSize() int {
	n := s.height - 1
	if n < 1 {
		n = 1
	}
	return n
}

func (s *State) moveListCursor(delta int) {
	n := len(s.visibleRows())
	if n == 0 {
		return
	}
	s.cursor = render.WrapIndex(s.cursor, delta, n)
}

// ArticleAtCursor returns the article item under the list cursor, or false when
// the cursor is out of range or on a day-header row.
func (s *State) ArticleAtCursor() (*Item, bool) {
	rows := s.visibleRows()
	if s.cursor < 0 || s.cursor >= len(rows) {
		return nil, false
	}
	row := rows[s.cursor]
	if row.kind != rowArticle {
		return nil, false
	}
	return &s.groups[row.groupIdx].articles[row.artIdx], true
}

// HeaderKeyAtCursor returns the day key of the header row under the cursor, or
// "" when the cursor is not on a day-header row.
func (s *State) HeaderKeyAtCursor() string {
	rows := s.visibleRows()
	if s.cursor < 0 || s.cursor >= len(rows) {
		return ""
	}
	row := rows[s.cursor]
	if row.kind != rowHeader {
		return ""
	}
	return timeutil.DayKey(s.groups[row.groupIdx].date)
}

// SelectionForCursor returns the selection identity of the row under the
// cursor: the article ID when it is an article row, else the day key of the
// header when it is a header row ("" when the cursor is on neither).
func (s *State) SelectionForCursor() (articleID int64, headerKey string) {
	if item, ok := s.ArticleAtCursor(); ok {
		return item.ID, ""
	}
	return 0, s.HeaderKeyAtCursor()
}

// RestoreCursor repositions the cursor to the row matching the saved selection:
// the day header for headerKey, else the article row for articleID. A selection
// that no longer exists leaves the cursor clamped where it is.
func (s *State) RestoreCursor(articleID int64, headerKey string) {
	if headerKey != "" {
		for gi := range s.groups {
			if timeutil.DayKey(s.groups[gi].date) == headerKey {
				rows := s.visibleRows()
				for i, r := range rows {
					if r.kind == rowHeader && r.groupIdx == gi {
						s.cursor = i
						return
					}
				}
			}
		}
		return
	}
	if articleID != 0 {
		rows := s.visibleRows()
		for i, r := range rows {
			if r.kind == rowArticle && s.groups[r.groupIdx].articles[r.artIdx].ID == articleID {
				s.cursor = i
				return
			}
		}
	}
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
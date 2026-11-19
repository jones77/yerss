package listview

// Cursor returns the index of the selected row in the flattened visible-row
// list. The Model exposes it so higher layers can persist and restore the
// selection identity around view switches.
func (s *State) Cursor() int { return s.cursor }

// SetCursor moves the cursor to a raw visible-row index. Callers that need the
// index clamped to the current row list should reload first or drive movement
// through Update.
func (s *State) SetCursor(c int) { s.cursor = c }

// RowCount returns the number of visible rows (day headers plus expanded
// article rows) in the current list.
func (s *State) RowCount() int { return len(s.visibleRows()) }

// ArticleCount returns the number of loaded articles across all day groups
// (the filtered count shown in the status bar).
func (s *State) ArticleCount() int { return s.articleCount() }

// CursorIsHeader reports whether the cursor rests on a day-header row.
func (s *State) CursorIsHeader() bool {
	rows := s.visibleRows()
	return s.cursor >= 0 && s.cursor < len(rows) && rows[s.cursor].kind == rowHeader
}

// Collapsed reports whether the day group at the given index is folded.
func (s *State) Collapsed(gi int) bool {
	return gi >= 0 && gi < len(s.groups) && s.groups[gi].collapsed
}

// ArticleIDAtRow returns the article ID of the visible row at the given index,
// or 0 when that row is a day header or out of range.
func (s *State) ArticleIDAtRow(i int) int64 {
	rows := s.visibleRows()
	if i < 0 || i >= len(rows) {
		return 0
	}
	row := rows[i]
	if row.kind != rowArticle {
		return 0
	}
	return s.groups[row.groupIdx].articles[row.artIdx].ID
}

// ItemAt returns the article item at the given day group and article index, or
// nil when out of range.
func (s *State) ItemAt(gi, ai int) *Item {
	if gi < 0 || gi >= len(s.groups) || ai < 0 || ai >= len(s.groups[gi].articles) {
		return nil
	}
	return &s.groups[gi].articles[ai]
}

// ArticleIDs returns the IDs of every loaded article across all day groups, in
// display order.
func (s *State) ArticleIDs() []int64 {
	var ids []int64
	for gi := range s.groups {
		for ai := range s.groups[gi].articles {
			ids = append(ids, s.groups[gi].articles[ai].ID)
		}
	}
	return ids
}

// SelectArticleID moves the cursor to the article with the given ID and reports
// whether the article is present in the loaded list. A collapsed day group is
// expanded so the article's row becomes visible before the raw cursor is set to
// it.
func (s *State) SelectArticleID(id int64) bool {
	for gi := range s.groups {
		for ai := range s.groups[gi].articles {
			if s.groups[gi].articles[ai].ID != id {
				continue
			}
			s.groups[gi].collapsed = false
			row := 0
			for g := 0; g < gi; g++ {
				row++
				if !s.groups[g].collapsed {
					row += len(s.groups[g].articles)
				}
			}
			row++
			row += ai
			s.cursor = row
			return true
		}
	}
	return false
}

// RenderArticleRow renders a single article row. It is exported for the Model
// layer's tests; the list's own rendering calls the unexported version.
func (s *State) RenderArticleRow(item *Item, selected bool) string {
	return s.renderArticleRow(item, selected)
}

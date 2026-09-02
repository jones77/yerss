package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
)

func newTestModel(t *testing.T) (*Model, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	m := New(app.New(config.Default(), st))
	m.width = 80
	m.height = 24
	return m, st
}

func insertArticle(t *testing.T, st *store.Store, title string, tags []string) int64 {
	t.Helper()
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-" + title,
		Title:   title,
		Content: "<p>" + title + "</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle: %v", err)
	}
	if err := st.SetArticleTags(id, tags); err != nil {
		t.Fatalf("SetArticleTags: %v", err)
	}
	return id
}

func TestTagCountText(t *testing.T) {
	cases := []struct {
		unread, total int
		plain, bold   string
	}{
		{60, 60, "", "60"},
		{0, 60, "60", ""},
		{15, 60, "/60", "15"},
	}
	for _, c := range cases {
		plain, bold := tagCountText(c.unread, c.total)
		if plain != c.plain || bold != c.bold {
			t.Errorf("tagCountText(%d,%d) = (%q,%q), want (%q,%q)",
				c.unread, c.total, plain, bold, c.plain, c.bold)
		}
	}
}

func TestTagFiltering(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "a1", []string{"tech"})
	insertArticle(t, st, "a2", []string{"news"})

	m.loadList()
	if m.articleCount() != 2 {
		t.Fatalf("expected 2 articles, got %d", m.articleCount())
	}

	m.popupData = popupState{
		tags:   []store.TagCount{{Name: "tech", Total: 1, Unread: 1}},
		cursor: 0,
	}
	m.confirmTagSelection()
	if m.list.filter != "tech" {
		t.Errorf("filter = %q, want tech", m.list.filter)
	}
	if m.articleCount() != 1 || !containsID(loadedIDs(m), id1) {
		t.Fatalf("expected only tech article, got %v", loadedIDs(m))
	}

	m.clearFilter()
	if m.list.filter != "" {
		t.Errorf("filter after clear = %q", m.list.filter)
	}
	if m.articleCount() != 2 {
		t.Fatalf("expected all articles after clear, got %d", m.articleCount())
	}
}

func loadedIDs(m *Model) []int64 {
	var ids []int64
	for gi := range m.list.groups {
		for ai := range m.list.groups[gi].articles {
			ids = append(ids, m.list.groups[gi].articles[ai].ID)
		}
	}
	return ids
}

func containsID(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestPopupWrapAround(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{
		tags: []store.TagCount{
			{Name: "a"}, {Name: "b"},
		},
		cursor: 1,
	}
	m.movePopupCursor(1)
	if m.popupData.cursor != 0 {
		t.Errorf("wrap-down: cursor = %d, want 0", m.popupData.cursor)
	}
	m.movePopupCursor(-1)
	if m.popupData.cursor != 1 {
		t.Errorf("wrap-up: cursor = %d, want 1", m.popupData.cursor)
	}
}

func makeTags(n int) []store.TagCount {
	tags := make([]store.TagCount, n)
	for i := range tags {
		tags[i] = store.TagCount{Name: fmt.Sprintf("tag%02d", i+1), Total: 1, Unread: 1}
	}
	return tags
}

func TestTagPopupSizeMargins(t *testing.T) {
	m, _ := newTestModel(t)
	h, w := m.tagPopupSize()
	if h != 20 || w != 56 {
		t.Errorf("tagPopupSize at 80x24 = %dx%d, want 20x56", h, w)
	}
	// Tiny terminals clamp to a minimal usable box, never negative.
	m.width, m.height = 30, 6
	h, w = m.tagPopupSize()
	if h != 3 || w != 12 {
		t.Errorf("tagPopupSize at 30x6 = %dx%d, want 3x12", h, w)
	}
}

func tagNameAt(i int) string { return fmt.Sprintf("tag%02d", i+1) }

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func TestTagPopupWindowSymmetricScroll(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: wideTags(60, 16)} // 4 columns of 23 at 80x24
	rows, cols, _, _ := m.tagGrid()
	if cols < 4 {
		t.Fatalf("want a scrollable grid, got %d columns", cols)
	}
	// Start at the last column; the window advances right to reveal it and the
	// first column is off-screen.
	m.popupData.cursor = (cols - 1) * rows
	if s := ansi.Strip(m.renderTagPopup()); strings.Contains(s, "tag01") {
		t.Fatalf("first column should be off-screen at the last column")
	}
	// Moving left within the window must not scroll it: the first column stays
	// off-screen until the cursor actually reaches it.
	for step := 1; step < cols-1; step++ {
		m.moveTagCursorHorizontal(-1)
		if s := ansi.Strip(m.renderTagPopup()); strings.Contains(s, "tag01") {
			t.Fatalf("first column visible too early (step %d, scrollCol %d)", step, m.popupData.scrollCol)
		}
	}
	// Reaching the first column scrolls the window back to the start.
	for m.tagDispPos(m.popupData.cursor)/rows != 0 {
		m.moveTagCursorHorizontal(-1)
	}
	if s := ansi.Strip(m.renderTagPopup()); !strings.Contains(s, "tag01") {
		t.Fatalf("first column should be visible at the start")
	}
}

func TestTagPopupColumnsFillInterior(t *testing.T) {
	m, _ := newTestModel(t) // 80x24 → box 56 wide, interior 52
	m.popup = popupTagsList
	// Many short tags → 3 columns; the spare space widens them to one equal
	// width that fills the interior (two spaces between columns).
	m.popupData = popupState{tags: makeTags(40)}
	_, cols, colWidths, boxW := m.tagGrid()
	if cols != 3 {
		t.Fatalf("cols = %d, want 3", cols)
	}
	if colWidths[0] != colWidths[1] || colWidths[1] != colWidths[2] {
		t.Errorf("columns should fill to equal widths, got %v", colWidths)
	}
	if boxW != 56 {
		t.Errorf("box width should stay 56 (no remainder), got %d", boxW)
	}
	// grid = 16*3 + 2 gaps*2 = 52 = boxW-4: no trailing space.
	if got := colWidths[0]*cols + (cols-1)*tagColGap; got != boxW-4 {
		t.Errorf("grid width %d should equal the interior %d", got, boxW-4)
	}
	// A single column fills the whole interior, leaving the box width alone.
	m.popupData = popupState{tags: makeTags(5)}
	_, _, colWidths, boxW = m.tagGrid()
	if colWidths[0] != 52 {
		t.Errorf("single column should fill the 52-column interior, got %d", colWidths[0])
	}
	if boxW != 56 {
		t.Errorf("single column should keep the full box width, got %d", boxW)
	}
	// Four columns leave a 2-column remainder that pulls the right border in.
	m.popupData = popupState{tags: makeTags(56)}
	_, cols, colWidths, boxW = m.tagGrid()
	if cols != 4 {
		t.Fatalf("cols = %d, want 4", cols)
	}
	if colWidths[0] != colWidths[1] || colWidths[1] != colWidths[2] || colWidths[2] != colWidths[3] {
		t.Errorf("four columns should fill to equal widths, got %v", colWidths)
	}
	if boxW != 54 {
		t.Errorf("four-column grid should shrink the box to 54, got %d", boxW)
	}
	// Every rendered grid row fills its box: no trailing space on the right.
	m.popupData = popupState{tags: makeTags(40)}
	lines := strings.Split(ansi.Strip(m.renderTagPopup()), "\n")
	for i := 1; i < len(lines)-1; i++ {
		if ansi.StringWidth(lines[i]) != 56 {
			t.Errorf("grid row %d = %d wide, want 56: %q", i, ansi.StringWidth(lines[i]), lines[i])
		}
	}
}

func TestTagPopupScrollWindowFillsInterior(t *testing.T) {
	m, _ := newTestModel(t)
	m.width, m.height = 104, 31
	m.popup = popupTagsList
	// Wide names cap the columns at 23 and enough tags make the grid wider than
	// the interior, so the visible window scrolls; it must still span the
	// interior with no trailing space.
	m.popupData = popupState{tags: wideTags(80, 16)} // entries hit the 23-column cap
	_, cols, _, _ := m.tagGrid()
	if cols < 4 {
		t.Fatalf("want a grid wider than the interior, got %d columns", cols)
	}
	lines := strings.Split(ansi.Strip(m.renderTagPopup()), "\n")
	boxW := ansi.StringWidth(lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if ansi.StringWidth(lines[i]) != boxW {
			t.Errorf("grid row %d = %d wide, want box width %d: %q", i, ansi.StringWidth(lines[i]), boxW, lines[i])
		}
	}
}

func TestTagPopupGridLayout(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(40)}
	rows, _, _, _ := m.tagGrid()
	s := ansi.Strip(m.renderTagPopup())
	for i := 1; i <= 40; i++ {
		if !strings.Contains(s, fmt.Sprintf("tag%02d", i)) {
			t.Errorf("popup missing tag%02d", i)
		}
	}
	lineOf := func(sub string) string {
		for _, l := range strings.Split(s, "\n") {
			if strings.Contains(l, sub) {
				return l
			}
		}
		return ""
	}
	// Row 0 holds the first tag of each column (column-major fill).
	if row := lineOf("tag01"); !containsAll(row, tagNameAt(0), tagNameAt(rows), tagNameAt(2*rows)) || strings.Contains(row, "tag02") {
		t.Errorf("first grid row should hold the column 0/1/2 starts (tag01, tag%02d, tag%02d): %q", rows+1, 2*rows+1, row)
	}
	if row := lineOf("tag02"); !containsAll(row, tagNameAt(1), tagNameAt(rows+1), tagNameAt(2*rows+1)) {
		t.Errorf("second grid row should hold tag02, tag%02d, tag%02d: %q", rows+2, 2*rows+2, row)
	}
	if row := lineOf(tagNameAt(rows - 1)); !strings.Contains(row, tagNameAt(2*rows-1)) || strings.Contains(row, tagNameAt(2*rows)) {
		t.Errorf("last full row should hold tag%02d and tag%02d (final column is shorter): %q", rows, 2*rows, row)
	}
}

func wideTags(n, nameLen int) []store.TagCount {
	tags := make([]store.TagCount, n)
	for i := range tags {
		tags[i] = store.TagCount{
			Name:  fmt.Sprintf("tag%02d", i+1) + strings.Repeat("x", nameLen),
			Total: 1, Unread: 1,
		}
	}
	return tags
}

func TestTagPopupColumnWidthCap(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	// Names short enough to fit the 23-column cap render in full.
	m.popupData = popupState{tags: wideTags(20, 15)} // entry = 5+15+1+1 = 22 <= 23
	s := ansi.Strip(m.renderTagPopup())
	full := "tag01" + strings.Repeat("x", 15)
	if !strings.Contains(s, full) {
		t.Errorf("entry within the 23-column cap should render in full: %q", s)
	}
	if strings.Contains(s, "…") {
		t.Errorf("no elision expected within the cap: %q", s)
	}
	// Names over the cap are middle-elided, keeping both ends.
	m.popupData = popupState{tags: wideTags(20, 30)} // entry = 37 > 23
	s = ansi.Strip(m.renderTagPopup())
	if !strings.Contains(s, "…") {
		t.Errorf("entry over the cap should be middle-elided: %q", s)
	}
	if !strings.Contains(s, "tag01") {
		t.Errorf("elided entry should keep its start: %q", s)
	}
	if strings.Contains(s, "tag01"+strings.Repeat("x", 30)) {
		t.Errorf("over-cap entry should not render in full: %q", s)
	}
}

func TestTagPopupMultipleColumnsVisible(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	tags := makeTags(40)
	tags[0].Name = "tag01" + strings.Repeat("x", 15) // 20 chars → column 0 is 26 wide
	m.popupData = popupState{tags: tags, cursor: 0}
	rows, _, _, _ := m.tagGrid()
	s := ansi.Strip(m.renderTagPopup())
	line := ""
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, "tag01") {
			line = l
			break
		}
	}
	if !strings.Contains(line, tagNameAt(rows)) {
		t.Errorf("second column should be visible next to a long first column: %q", line)
	}
	if !strings.Contains(line, strings.Repeat("x", 15)) {
		t.Errorf("the long entry should render in full: %q", line)
	}
}

func TestTagPopupScrollsRight(t *testing.T) {
	m, _ := newTestModel(t) // 80x24 → 18 rows, 23-cap columns → 2 of 3 columns visible
	m.popup = popupTagsList
	m.popupData = popupState{tags: wideTags(40, 16)} // entries hit the 23-column cap
	rows, _, _, _ := m.tagGrid()
	// Cursor in the first column shows columns 0 and 1; the third is off-screen.
	m.popupData.cursor = 0
	s := ansi.Strip(m.renderTagPopup())
	if !strings.Contains(s, "tag01") {
		t.Errorf("first column should be visible: %q", s)
	}
	if strings.Contains(s, tagNameAt(2*rows)) {
		t.Errorf("third column should be off-screen: %q", s)
	}
	// Moving into the third column scrolls the window right, hiding the first.
	m.popupData.cursor = 2 * rows
	s = ansi.Strip(m.renderTagPopup())
	if !strings.Contains(s, tagNameAt(2*rows)) {
		t.Errorf("third column should be visible after scrolling: %q", s)
	}
	if strings.Contains(s, "tag01") {
		t.Errorf("first column should be scrolled off-screen: %q", s)
	}
}

func TestTagPopupTinyTerminalDoesNotPanic(t *testing.T) {
	m, _ := newTestModel(t)
	m.width, m.height = 30, 6
	m.popup = popupTagsArticle
	m.popupData = popupState{tags: makeTags(40)}
	m.renderTagPopup()
}

func TestTagGridContinuesIntoNextColumn(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{tags: makeTags(40)}
	rows, _, _, _ := m.tagGrid()
	m.popupData.cursor = rows - 1 // bottom of column 0
	m.movePopupCursor(1)
	if m.popupData.cursor != rows {
		t.Errorf("down from bottom of column 0 = %d, want %d (top of column 1)", m.popupData.cursor, rows)
	}
	m.popupData.cursor = 39
	m.movePopupCursor(1)
	if m.popupData.cursor != 0 {
		t.Errorf("down from last tag = %d, want 0", m.popupData.cursor)
	}
	m.popupData.cursor = 0
	m.movePopupCursor(-1)
	if m.popupData.cursor != 39 {
		t.Errorf("up from first tag = %d, want 39", m.popupData.cursor)
	}
}

func TestTagGridHorizontalWrap(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{tags: makeTags(40)}
	rows, cols, _, _ := m.tagGrid()
	m.popupData.cursor = 1 // column 0, row 1
	m.moveTagCursorHorizontal(-1)
	if want := (cols-1)*rows + 1; m.popupData.cursor != want {
		t.Errorf("left from column 0 row 1 = %d, want %d (column %d row 1)", m.popupData.cursor, want, cols-1)
	}
	m.moveTagCursorHorizontal(1)
	if m.popupData.cursor != 1 {
		t.Errorf("right from the last column row 1 = %d, want 1", m.popupData.cursor)
	}
	// Single column is a no-op.
	m.popupData = popupState{tags: makeTags(5), cursor: 2}
	m.moveTagCursorHorizontal(-1)
	if m.popupData.cursor != 2 {
		t.Errorf("left in a single-column grid = %d, want 2 (no-op)", m.popupData.cursor)
	}
}

func TestTagGridHorizontalClamp(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{tags: makeTags(40)}
	rows, cols, _, _ := m.tagGrid()
	// Tag 34 sits on the last row of column 1 (the "Tags" subtitle occupies
	// the top cell of column 0, shifting the tags down by one row).
	m.popupData.cursor = 34
	m.moveTagCursorHorizontal(1)
	if m.popupData.cursor != 39 {
		t.Errorf("right into the shorter final column = %d, want 39 (clamped)", m.popupData.cursor)
	}
	m.popupData.cursor = (cols - 1) * rows // final column, row 0
	m.moveTagCursorHorizontal(1)
	if m.popupData.cursor != 0 {
		t.Errorf("right from final column row 0 = %d, want 0", m.popupData.cursor)
	}
}

func TestTagGridJumps(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{tags: makeTags(40), cursor: 12}
	m.jumpTagCursor(true)
	if m.popupData.cursor != 0 {
		t.Errorf("jump to first = %d, want 0", m.popupData.cursor)
	}
	m.jumpTagCursor(false)
	if m.popupData.cursor != 39 {
		t.Errorf("jump to last = %d, want 39", m.popupData.cursor)
	}
	m.popupData = popupState{tags: nil, cursor: 0}
	m.jumpTagCursor(true)
	m.jumpTagCursor(false)
}

func TestTagPopupNavKeys(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(40)}
	rows, cols, _, _ := m.tagGrid()

	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.popupData.cursor != 0 {
		t.Errorf("1 should select the first tag, got %d", m.popupData.cursor)
	}
	m.popupData.cursor = 20
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.popupData.cursor != 0 {
		t.Errorf("g should select the first tag, got %d", m.popupData.cursor)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m.popupData.cursor != 39 {
		t.Errorf("G should select the last tag, got %d", m.popupData.cursor)
	}

	wantLeft := (cols - 1) * rows
	m.popupData.cursor = 1 // column 0, row 1
	m.updatePopup(tea.KeyMsg{Type: tea.KeyLeft})
	if m.popupData.cursor != wantLeft+1 {
		t.Errorf("left should wrap to the last column row 1, got %d", m.popupData.cursor)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRight})
	if m.popupData.cursor != 1 {
		t.Errorf("right should wrap back to column 0 row 1, got %d", m.popupData.cursor)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.popupData.cursor != wantLeft+1 {
		t.Errorf("h should wrap to the last column row 1, got %d", m.popupData.cursor)
	}
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.popupData.cursor != 1 {
		t.Errorf("l should wrap back to column 0 row 1, got %d", m.popupData.cursor)
	}

	m.popupData.cursor = rows - 1 // bottom of column 0
	m.updatePopup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.popupData.cursor != rows {
		t.Errorf("j should continue into column 1, got %d", m.popupData.cursor)
	}
	m.popupData.cursor = 39
	m.updatePopup(tea.KeyMsg{Type: tea.KeyDown})
	if m.popupData.cursor != 0 {
		t.Errorf("down from the last tag should wrap to the first, got %d", m.popupData.cursor)
	}
}

func TestTagPopupCellLayout(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(40), cursor: 1}
	rows, _, _, _ := m.tagGrid()
	line := ""
	for _, l := range strings.Split(ansi.Strip(m.renderTagPopup()), "\n") {
		if strings.Contains(l, "tag01") {
			line = l
			break
		}
	}
	if !strings.HasPrefix(line, "│ tag01") {
		t.Errorf("first tag should sit one space from the popup edge: %q", line)
	}
	if !strings.Contains(line, "1  "+tagNameAt(rows)) || !strings.Contains(line, "1  "+tagNameAt(2*rows)) {
		t.Errorf("columns should be separated by two spaces: %q", line)
	}
	// The selection highlight starts exactly at the tag, covering the full
	// column width, with no gutter before it.
	raw := m.renderTagPopup()
	bg := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#2c2c2c")).Render("x"))
	i := strings.Index(raw, bg)
	if i < 0 {
		t.Fatal("selected cell should be highlighted")
	}
	if rest := ansi.Strip(raw[i+len(bg):]); !strings.HasPrefix(rest, "tag02") {
		t.Errorf("highlight should start exactly at the tag with no gutter: %q", rest[:20])
	}
}

func TestTagPopupHighlightTracksColumnWidth(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	tags := makeTags(56)
	tags[0].Name = "tag01" + strings.Repeat("x", 28) // widens column 0 only
	m.popupData = popupState{tags: tags}
	rows, _, colWidths, _ := m.tagGrid()
	if colWidths[0] == colWidths[1] {
		t.Fatalf("columns should differ in width for this test, got %v", colWidths)
	}
	// The selected cell's highlight spans its whole column, and the width
	// changes when the cursor moves to a differently-sized column.
	for _, idx := range []int{0, 1, rows} { // long in col0, short in col0, col1
		col := idx / rows
		m.popupData.cursor = idx
		cell := m.tagCell(idx, colWidths[col])
		if got := ansi.StringWidth(ansi.Strip(cell)); got != colWidths[col] {
			t.Errorf("selected cell in column %d renders %d columns, want column width %d", col, got, colWidths[col])
		}
	}
	if got := ansi.StringWidth(ansi.Strip(m.tagCell(1, colWidths[0]))); got != colWidths[0] {
		t.Errorf("highlight should be fixed to the column width, got %d want %d", got, colWidths[0])
	}
	if got := ansi.StringWidth(ansi.Strip(m.tagCell(rows, colWidths[1]))); got != colWidths[1] {
		t.Errorf("highlight should narrow to the new column's width, got %d want %d", got, colWidths[1])
	}
}

func TestTagPopupSelectionUsesBackgroundHighlight(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	// Select a short item in a wide column so the highlight must extend across
	// the column's trailing padding to match the widest entry.
	tags := makeTags(20)
	tags[0].Name = "tag01" + strings.Repeat("x", 26) // widens column 0
	m.popupData = popupState{tags: tags, cursor: 1}
	s := m.renderTagPopup()
	if strings.Contains(s, "> ") {
		t.Errorf("selection should not use a > marker: %q", s)
	}
	bgEscape := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#2c2c2c")).Render("x"))
	if !strings.Contains(s, bgEscape) {
		t.Errorf("selected cell should be highlighted: %q", s)
	}
	// The highlight covers the trailing padding too: a background escape is
	// immediately followed by the run of pad spaces (two or more).
	if !strings.Contains(s, bgEscape+"  ") {
		t.Errorf("highlight should extend across the column padding: %q", s)
	}
}

func TestTagPopupCountsRightAligned(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	// Two tags in one column: the long-named tag sizes the column to the
	// 23-column cap, and the short-named tag's counts right-align to the edge.
	tags := []store.TagCount{
		{Name: "aaaa", Total: 12, Unread: 3},
		{Name: "bbbbbbbbbbbbbbbbbbbbbb", Total: 5, Unread: 5},
	}
	m.popupData = popupState{tags: tags}
	_, _, colWidths, _ := m.tagGrid()
	cell := ansi.Strip(m.tagCell(0, colWidths[0]))
	if ansi.StringWidth(cell) != colWidths[0] {
		t.Fatalf("cell width = %d, want %d", ansi.StringWidth(cell), colWidths[0])
	}
	// The counts sit at the right edge: the cell ends with "3/12".
	if !strings.HasSuffix(cell, "3/12") {
		t.Errorf("counts should right-align to the column edge: %q", cell)
	}
	// The name is left-aligned with padding between it and the counts.
	trimmed := strings.TrimRight(cell[:len(cell)-len("3/12")], " ")
	if trimmed != "aaaa" {
		t.Errorf("name should be left-aligned with padding before the counts: %q", cell)
	}
}

func TestTagPopupTitleInlineInBorder(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(5)}
	s := ansi.Strip(m.renderTagPopup())
	lines := strings.Split(s, "\n")
	if !strings.Contains(lines[0], "Tags & Sources") {
		t.Errorf("title should be inline in the top border: %q", lines[0])
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "Tags & Sources" {
			t.Errorf("title should not appear as a body line: %q", l)
		}
	}
}

func TestTagPopupTitleArticleView(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsArticle
	m.popupData = popupState{tags: makeTags(5)}
	lines := strings.Split(ansi.Strip(m.renderTagPopup()), "\n")
	if !strings.Contains(lines[0], "Tags & Sources") {
		t.Errorf("article-view popup title should be Tags & Sources: %q", lines[0])
	}
}

func TestTagSubtitleCellCentered(t *testing.T) {
	m, _ := newTestModel(t)
	cases := []struct {
		width int
		want  string
	}{
		{12, "    Tags    "}, // 4-wide label, even padding
		{11, "   Tags    "},  // odd padding: left gets the smaller half
		{4, "Tags"},          // exact fit, no padding
	}
	for _, c := range cases {
		if got := ansi.Strip(m.tagSubtitleCell(tagItem{label: "Tags", tagIdx: -1}, c.width)); got != c.want {
			t.Errorf("tagSubtitleCell(Tags, %d) = %q, want %q", c.width, got, c.want)
		}
	}
}

func TestTagPopupSubtitlesGroupTheTags(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: []store.TagCount{
		{Name: "nytimes", Total: 3, Unread: 1, IsSource: true},
		{Name: "theguardian", Total: 2, Unread: 0, IsSource: true},
		{Name: "tech", Total: 5, Unread: 2},
		{Name: "news", Total: 4, Unread: 1},
	}}
	s := ansi.Strip(m.renderTagPopup())
	lines := strings.Split(s, "\n")
	// Content lives in the grid body, so skip the top border's title line.
	rowOf := func(sub string) int {
		for i := 1; i < len(lines); i++ {
			if strings.Contains(lines[i], sub) {
				return i
			}
		}
		return -1
	}
	srcRow := rowOf("Sources")
	if srcRow < 0 {
		t.Fatalf("missing Sources subtitle: %q", s)
	}
	tagsRow := rowOf("Tags")
	if tagsRow < 0 {
		t.Fatalf("missing Tags subtitle: %q", s)
	}
	nyRow := rowOf("nytimes")
	if nyRow < 0 {
		t.Fatalf("missing source tag: %q", s)
	}
	techRow := rowOf("tech")
	if techRow < 0 {
		t.Fatalf("missing category tag: %q", s)
	}
	// Sources come first (column-major): the Sources subtitle sits above the
	// source tags, the Tags subtitle above the category tags.
	if !(srcRow < nyRow && nyRow < tagsRow && tagsRow < techRow) {
		t.Errorf("subtitle/group order wrong (Sources %d, nytimes %d, Tags %d, tech %d): %q",
			srcRow, nyRow, tagsRow, techRow, s)
	}
	// A tags-only popup still labels its section.
	m.popupData = popupState{tags: []store.TagCount{{Name: "tech", Total: 1, Unread: 1}}}
	s = ansi.Strip(m.renderTagPopup())
	if !strings.Contains(strings.Split(s, "\n")[1], "Tags") {
		t.Errorf("category-only popup should show the Tags subtitle: %q", s)
	}
}

func TestTagPopupSubtitlesNotSelectable(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: []store.TagCount{
		{Name: "nytimes", Total: 3, Unread: 1, IsSource: true},
		{Name: "tech", Total: 5, Unread: 2},
	}}
	// The cursor starts on the first tag (index 0); cycling down and up wraps
	// within the tag list, never landing on a subtitle.
	for i := 0; i < 40; i++ {
		m.movePopupCursor(1)
		if m.popupData.cursor < 0 || m.popupData.cursor >= len(m.popupData.tags) {
			t.Fatalf("cursor %d escaped the tag list", m.popupData.cursor)
		}
	}
	// Horizontal moves skip subtitles too.
	m.popupData.cursor = len(m.popupData.tags) - 1
	m.moveTagCursorHorizontal(-1)
	if m.popupData.cursor < 0 || m.popupData.cursor >= len(m.popupData.tags) {
		t.Fatalf("horizontal cursor %d escaped the tag list", m.popupData.cursor)
	}
	// Subtitles render in the status-bar blue role without the selection
	// background and without bold.
	forceTrueColor(t)
	m.popupData.cursor = 0
	s := m.renderTagPopup()
	bgEscape := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#2c2c2c")).Render("x"))
	blueEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render("x"))
	subLine := ""
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, "Sources") {
			subLine = l
		}
	}
	if subLine == "" {
		t.Fatalf("missing Sources subtitle: %q", s)
	}
	if strings.Contains(subLine, bgEscape) {
		t.Errorf("subtitle should not carry the selection highlight: %q", subLine)
	}
	if !strings.Contains(subLine, blueEscape) {
		t.Errorf("subtitle should render in the status-bar blue role: %q", subLine)
	}
}

func TestTagPopupCountUsesListViewColors(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{
		tags: []store.TagCount{
			{Name: "news", Total: 60, Unread: 15},
			{Name: "sports", Total: 10, Unread: 0},
		},
		cursor: 1, // keep the counts cell unselected so colors render plainly
	}
	textEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.Text).Render("x"))
	dimEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.Dim).Render("x"))
	newsLine := ""
	for _, l := range strings.Split(m.renderTagPopup(), "\n") {
		if strings.Contains(l, "news") {
			newsLine = l
			break
		}
	}
	if !strings.Contains(newsLine, textEscape) {
		t.Errorf("count total should render in the list-view text color: %q", newsLine)
	}
	if strings.Contains(newsLine, dimEscape) {
		t.Errorf("count total should not render in the dim role: %q", newsLine)
	}
}

func TestTagPopupBottomBorderMatchesChrome(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(5), cursor: 0}
	lines := strings.Split(m.renderTagPopup(), "\n")
	top := lines[0]
	bottom := lines[len(lines)-1]
	chromeEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render("x"))
	if !strings.HasPrefix(top, chromeEscape) {
		t.Fatalf("top border should use the chrome role: %q", top)
	}
	if !strings.HasPrefix(bottom, chromeEscape) {
		t.Errorf("bottom border should match the rest of the popup border (chrome role): %q", bottom)
	}
}

func TestTagPopupBottomBorderPosition(t *testing.T) {
	m, _ := newTestModel(t)
	m.popup = popupTagsList
	m.popupData = popupState{tags: makeTags(40), cursor: 19} // 20th of 40 → 50%
	lines := strings.Split(ansi.Strip(m.renderTagPopup()), "\n")
	last := lines[len(lines)-1]
	if !strings.Contains(last, "50% · 20/40") {
		t.Errorf("bottom border should show the position indicator: %q", last)
	}
	// The dashes run continuously from the corner; no gap that would align
	// with the first column's first character.
	if !strings.HasPrefix(last, "└─") || strings.HasPrefix(last, "└─ ") {
		t.Errorf("bottom border should have no leading gap after the corner: %q", last)
	}
	// The first grid row sits directly under the top border (no top padding),
	// holding the centered "Tags" section subtitle (the tags are all
	// categories).
	if !strings.Contains(lines[1], "Tags") {
		t.Errorf("grid should start right after the top border with the Tags subtitle: %q", lines[1])
	}
}

func TestTagPopupGroupsSourcesFirst(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "a1", []string{"tech"})
	id2 := insertArticle(t, st, "a2", []string{"news"})
	_ = st.SetArticleTagsWithSource(id1, []string{"tech", "nytimes"}, "nytimes")
	_ = st.SetArticleTagsWithSource(id2, []string{"news", "theguardian"}, "theguardian")
	m.openTagPopup(popupTagsList)
	var names []string
	for _, tg := range m.popupData.tags {
		names = append(names, tg.Name)
	}
	want := []string{"nytimes", "theguardian", "news", "tech"}
	if len(names) != len(want) {
		t.Fatalf("popup tags = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("popup tag order = %v, want %v (sources first)", names, want)
			break
		}
	}
}

func TestOpenArticleStaleCursorNoPanic(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()

	m.list.cursor = len(m.visibleRows()) + 5
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with stale-high cursor changed view to %d", m.view)
	}

	m.list.cursor = -1
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with negative cursor changed view to %d", m.view)
	}

	m.list.groups = nil
	m.list.cursor = 0
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle on empty list changed view to %d", m.view)
	}
}

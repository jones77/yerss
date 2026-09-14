package popup

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui/render"
)

func newTestState(t *testing.T) (*State, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(app.New(config.Default(), st), render.DarkPalette(), false, 80, 24), st
}

func popupKeys(t *testing.T) map[string]config.Action {
	t.Helper()
	km, err := config.ParseKeybindings(config.DefaultKeybindings())
	if err != nil {
		t.Fatalf("ParseKeybindings: %v", err)
	}
	eff, err := km.EffectiveKeys(config.ViewPopup)
	if err != nil {
		t.Fatalf("EffectiveKeys: %v", err)
	}
	return eff
}

func forceTrueColor(t *testing.T) {
	t.Helper()
	orig := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(orig) })
}

func escapePrefix(rendered string) string {
	return strings.Split(rendered, "x")[0]
}

func makeTags(n int) []store.TagCount {
	tags := make([]store.TagCount, n)
	for i := range tags {
		tags[i] = store.TagCount{Name: fmt.Sprintf("tag%02d", i+1), Total: 1, Unread: 1}
	}
	return tags
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

func TestActionLabel(t *testing.T) {
	cases := map[config.Action]string{
		config.OpenArticle:     "Open Article",
		config.HalfPageDown:    "Half Page Down",
		config.CopyArticleText: "Copy Article Text",
	}
	for a, want := range cases {
		if got := actionLabel(a); got != want {
			t.Errorf("actionLabel(%s) = %q, want %q", a, got, want)
		}
	}
}

func TestPopupWrapAround(t *testing.T) {
	s, _ := newTestState(t)
	s.SetTags([]store.TagCount{
		{Name: "a"}, {Name: "b"},
	})
	s.SetCursor(1)
	s.movePopupCursor(1)
	if s.cursor != 0 {
		t.Errorf("wrap-down: cursor = %d, want 0", s.cursor)
	}
	s.movePopupCursor(-1)
	if s.cursor != 1 {
		t.Errorf("wrap-up: cursor = %d, want 1", s.cursor)
	}
}

func TestTagPopupSizeMargins(t *testing.T) {
	s, _ := newTestState(t)
	h, w := s.tagPopupSize()
	if h != 20 || w != 56 {
		t.Errorf("tagPopupSize at 80x24 = %dx%d, want 20x56", h, w)
	}
	// Tiny terminals clamp to a minimal usable box, never negative.
	s.SetSize(30, 6)
	h, w = s.tagPopupSize()
	if h != 3 || w != 12 {
		t.Errorf("tagPopupSize at 30x6 = %dx%d, want 3x12", h, w)
	}
}

func TestTagPopupWindowSymmetricScroll(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(wideTags(60, 16)) // 4 columns of 23 at 80x24
	rows, cols, _, _ := s.tagGrid()
	if cols < 4 {
		t.Fatalf("want a scrollable grid, got %d columns", cols)
	}
	// Start at the last column; the window advances right to reveal it and the
	// first column is off-screen.
	s.SetCursor((cols - 1) * rows)
	if out := ansi.Strip(s.renderTagPopup()); strings.Contains(out, "tag01") {
		t.Fatalf("first column should be off-screen at the last column")
	}
	// Moving left within the window must not scroll it: the first column stays
	// off-screen until the cursor actually reaches it.
	for step := 1; step < cols-1; step++ {
		s.moveTagCursorHorizontal(-1)
		if out := ansi.Strip(s.renderTagPopup()); strings.Contains(out, "tag01") {
			t.Fatalf("first column visible too early (step %d, scrollCol %d)", step, s.scrollCol)
		}
	}
	// Reaching the first column scrolls the window back to the start.
	for s.tagDispPos(s.cursor)/rows != 0 {
		s.moveTagCursorHorizontal(-1)
	}
	if out := ansi.Strip(s.renderTagPopup()); !strings.Contains(out, "tag01") {
		t.Fatalf("first column should be visible at the start")
	}
}

func TestTagPopupColumnsFillInterior(t *testing.T) {
	s, _ := newTestState(t) // 80x24 → box 56 wide, interior 52
	s.SetKind(KindTagsList)
	// Many short tags → 3 columns; the spare space widens them to one equal
	// width that fills the interior (two spaces between columns).
	s.SetTags(makeTags(40))
	_, cols, colWidths, boxW := s.tagGrid()
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
	s.SetTags(makeTags(5))
	_, _, colWidths, boxW = s.tagGrid()
	if colWidths[0] != 52 {
		t.Errorf("single column should fill the 52-column interior, got %d", colWidths[0])
	}
	if boxW != 56 {
		t.Errorf("single column should keep the full box width, got %d", boxW)
	}
	// Four columns leave a 2-column remainder that pulls the right border in.
	s.SetTags(makeTags(56))
	_, cols, colWidths, boxW = s.tagGrid()
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
	s.SetTags(makeTags(40))
	lines := strings.Split(ansi.Strip(s.renderTagPopup()), "\n")
	for i := 1; i < len(lines)-1; i++ {
		if ansi.StringWidth(lines[i]) != 56 {
			t.Errorf("grid row %d = %d wide, want 56: %q", i, ansi.StringWidth(lines[i]), lines[i])
		}
	}
}

func TestTagPopupScrollWindowFillsInterior(t *testing.T) {
	s, _ := newTestState(t)
	s.SetSize(104, 31)
	s.SetKind(KindTagsList)
	// Wide names cap the columns at 23 and enough tags make the grid wider than
	// the interior, so the visible window scrolls; it must still span the
	// interior with no trailing space.
	s.SetTags(wideTags(80, 16)) // entries hit the 23-column cap
	_, cols, _, _ := s.tagGrid()
	if cols < 4 {
		t.Fatalf("want a grid wider than the interior, got %d columns", cols)
	}
	lines := strings.Split(ansi.Strip(s.renderTagPopup()), "\n")
	boxW := ansi.StringWidth(lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if ansi.StringWidth(lines[i]) != boxW {
			t.Errorf("grid row %d = %d wide, want box width %d: %q", i, ansi.StringWidth(lines[i]), boxW, lines[i])
		}
	}
}

func TestTagPopupGridLayout(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(40))
	rows, _, _, _ := s.tagGrid()
	out := ansi.Strip(s.renderTagPopup())
	for i := 1; i <= 40; i++ {
		if !strings.Contains(out, fmt.Sprintf("tag%02d", i)) {
			t.Errorf("popup missing tag%02d", i)
		}
	}
	lineOf := func(sub string) string {
		for _, l := range strings.Split(out, "\n") {
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

func TestTagPopupColumnWidthCap(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	// Names short enough to fit the 23-column cap render in full.
	s.SetTags(wideTags(20, 15)) // entry = 5+15+1+1 = 22 <= 23
	out := ansi.Strip(s.renderTagPopup())
	full := "tag01" + strings.Repeat("x", 15)
	if !strings.Contains(out, full) {
		t.Errorf("entry within the 23-column cap should render in full: %q", out)
	}
	if strings.Contains(out, "…") {
		t.Errorf("no elision expected within the cap: %q", out)
	}
	// Names over the cap are middle-elided, keeping both ends.
	s.SetTags(wideTags(20, 30)) // entry = 37 > 23
	out = ansi.Strip(s.renderTagPopup())
	if !strings.Contains(out, "…") {
		t.Errorf("entry over the cap should be middle-elided: %q", out)
	}
	if !strings.Contains(out, "tag01") {
		t.Errorf("elided entry should keep its start: %q", out)
	}
	if strings.Contains(out, "tag01"+strings.Repeat("x", 30)) {
		t.Errorf("over-cap entry should not render in full: %q", out)
	}
}

func TestTagPopupMultipleColumnsVisible(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	tags := makeTags(40)
	tags[0].Name = "tag01" + strings.Repeat("x", 15) // 20 chars → column 0 is 26 wide
	s.SetTags(tags)
	s.SetCursor(0)
	rows, _, _, _ := s.tagGrid()
	out := ansi.Strip(s.renderTagPopup())
	line := ""
	for _, l := range strings.Split(out, "\n") {
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
	s, _ := newTestState(t) // 80x24 → 18 rows, 23-cap columns → 2 of 3 columns visible
	s.SetKind(KindTagsList)
	s.SetTags(wideTags(40, 16)) // entries hit the 23-column cap
	rows, _, _, _ := s.tagGrid()
	// Cursor in the first column shows columns 0 and 1; the third is off-screen.
	s.SetCursor(0)
	out := ansi.Strip(s.renderTagPopup())
	if !strings.Contains(out, "tag01") {
		t.Errorf("first column should be visible: %q", out)
	}
	if strings.Contains(out, tagNameAt(2*rows)) {
		t.Errorf("third column should be off-screen: %q", out)
	}
	// Moving into the third column scrolls the window right, hiding the first.
	s.SetCursor(2 * rows)
	out = ansi.Strip(s.renderTagPopup())
	if !strings.Contains(out, tagNameAt(2*rows)) {
		t.Errorf("third column should be visible after scrolling: %q", out)
	}
	if strings.Contains(out, "tag01") {
		t.Errorf("first column should be scrolled off-screen: %q", out)
	}
}

func TestTagPopupTinyTerminalDoesNotPanic(t *testing.T) {
	s, _ := newTestState(t)
	s.SetSize(30, 6)
	s.SetKind(KindTagsArticle)
	s.SetTags(makeTags(40))
	s.renderTagPopup()
}

func TestTagGridContinuesIntoNextColumn(t *testing.T) {
	s, _ := newTestState(t)
	s.SetTags(makeTags(40))
	rows, _, _, _ := s.tagGrid()
	s.SetCursor(rows - 1) // bottom of column 0
	s.movePopupCursor(1)
	if s.cursor != rows {
		t.Errorf("down from bottom of column 0 = %d, want %d (top of column 1)", s.cursor, rows)
	}
	s.SetCursor(39)
	s.movePopupCursor(1)
	if s.cursor != 0 {
		t.Errorf("down from last tag = %d, want 0", s.cursor)
	}
	s.SetCursor(0)
	s.movePopupCursor(-1)
	if s.cursor != 39 {
		t.Errorf("up from first tag = %d, want 39", s.cursor)
	}
}

func TestTagGridHorizontalWrap(t *testing.T) {
	s, _ := newTestState(t)
	s.SetTags(makeTags(40))
	rows, cols, _, _ := s.tagGrid()
	s.SetCursor(1) // column 0, row 1
	s.moveTagCursorHorizontal(-1)
	if want := (cols-1)*rows + 1; s.cursor != want {
		t.Errorf("left from column 0 row 1 = %d, want %d (column %d row 1)", s.cursor, want, cols-1)
	}
	s.moveTagCursorHorizontal(1)
	if s.cursor != 1 {
		t.Errorf("right from the last column row 1 = %d, want 1", s.cursor)
	}
	// Single column is a no-op.
	s.SetTags(makeTags(5))
	s.SetCursor(2)
	s.moveTagCursorHorizontal(-1)
	if s.cursor != 2 {
		t.Errorf("left in a single-column grid = %d, want 2 (no-op)", s.cursor)
	}
}

func TestTagGridHorizontalClamp(t *testing.T) {
	s, _ := newTestState(t)
	s.SetTags(makeTags(40))
	rows, cols, _, _ := s.tagGrid()
	// Tag 34 sits on the last row of column 1 (the "Tags" subtitle occupies
	// the top cell of column 0, shifting the tags down by one row).
	s.SetCursor(34)
	s.moveTagCursorHorizontal(1)
	if s.cursor != 39 {
		t.Errorf("right into the shorter final column = %d, want 39 (clamped)", s.cursor)
	}
	s.SetCursor((cols - 1) * rows) // final column, row 0
	s.moveTagCursorHorizontal(1)
	if s.cursor != 0 {
		t.Errorf("right from final column row 0 = %d, want 0", s.cursor)
	}
}

func TestTagGridJumps(t *testing.T) {
	s, _ := newTestState(t)
	s.SetTags(makeTags(40))
	s.SetCursor(12)
	s.jumpTagCursor(true)
	if s.cursor != 0 {
		t.Errorf("jump to first = %d, want 0", s.cursor)
	}
	s.jumpTagCursor(false)
	if s.cursor != 39 {
		t.Errorf("jump to last = %d, want 39", s.cursor)
	}
	s.SetTags(nil)
	s.SetCursor(0)
	s.jumpTagCursor(true)
	s.jumpTagCursor(false)
}

func TestTagPopupNavKeys(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(40))
	rows, cols, _, _ := s.tagGrid()
	keys := popupKeys(t)

	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}, keys)
	if s.cursor != 0 {
		t.Errorf("1 should select the first tag, got %d", s.cursor)
	}
	s.SetCursor(20)
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}, keys)
	if s.cursor != 0 {
		t.Errorf("g should select the first tag, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}, keys)
	if s.cursor != 39 {
		t.Errorf("G should select the last tag, got %d", s.cursor)
	}

	wantLeft := (cols - 1) * rows
	s.SetCursor(1) // column 0, row 1
	s.Update(tea.KeyMsg{Type: tea.KeyLeft}, keys)
	if s.cursor != wantLeft+1 {
		t.Errorf("left should wrap to the last column row 1, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyRight}, keys)
	if s.cursor != 1 {
		t.Errorf("right should wrap back to column 0 row 1, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, keys)
	if s.cursor != wantLeft+1 {
		t.Errorf("h should wrap to the last column row 1, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, keys)
	if s.cursor != 1 {
		t.Errorf("l should wrap back to column 0 row 1, got %d", s.cursor)
	}

	s.SetCursor(rows - 1) // bottom of column 0
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, keys)
	if s.cursor != rows {
		t.Errorf("j should continue into column 1, got %d", s.cursor)
	}
	s.SetCursor(39)
	s.Update(tea.KeyMsg{Type: tea.KeyDown}, keys)
	if s.cursor != 0 {
		t.Errorf("down from the last tag should wrap to the first, got %d", s.cursor)
	}
}

func TestTagPopupCellLayout(t *testing.T) {
	forceTrueColor(t)
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(40))
	s.SetCursor(1)
	rows, _, _, _ := s.tagGrid()
	line := ""
	for _, l := range strings.Split(ansi.Strip(s.renderTagPopup()), "\n") {
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
	raw := s.renderTagPopup()
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
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	tags := makeTags(56)
	tags[0].Name = "tag01" + strings.Repeat("x", 28) // widens column 0 only
	s.SetTags(tags)
	rows, _, colWidths, _ := s.tagGrid()
	if colWidths[0] == colWidths[1] {
		t.Fatalf("columns should differ in width for this test, got %v", colWidths)
	}
	// The selected cell's highlight spans its whole column, and the width
	// changes when the cursor moves to a differently-sized column.
	for _, idx := range []int{0, 1, rows} { // long in col0, short in col0, col1
		col := idx / rows
		s.SetCursor(idx)
		cell := s.tagCell(idx, colWidths[col])
		if got := ansi.StringWidth(ansi.Strip(cell)); got != colWidths[col] {
			t.Errorf("selected cell in column %d renders %d columns, want column width %d", col, got, colWidths[col])
		}
	}
	if got := ansi.StringWidth(ansi.Strip(s.tagCell(1, colWidths[0]))); got != colWidths[0] {
		t.Errorf("highlight should be fixed to the column width, got %d want %d", got, colWidths[0])
	}
	if got := ansi.StringWidth(ansi.Strip(s.tagCell(rows, colWidths[1]))); got != colWidths[1] {
		t.Errorf("highlight should narrow to the new column's width, got %d want %d", got, colWidths[1])
	}
}

func TestTagPopupSelectionUsesBackgroundHighlight(t *testing.T) {
	forceTrueColor(t)
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	// Select a short item in a wide column so the highlight must extend across
	// the column's trailing padding to match the widest entry.
	tags := makeTags(20)
	tags[0].Name = "tag01" + strings.Repeat("x", 26) // widens column 0
	s.SetTags(tags)
	s.SetCursor(1)
	out := s.renderTagPopup()
	if strings.Contains(out, "> ") {
		t.Errorf("selection should not use a > marker: %q", out)
	}
	bgEscape := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#2c2c2c")).Render("x"))
	if !strings.Contains(out, bgEscape) {
		t.Errorf("selected cell should be highlighted: %q", out)
	}
	// The highlight covers the trailing padding too: a background escape is
	// immediately followed by the run of pad spaces (two or more).
	if !strings.Contains(out, bgEscape+"  ") {
		t.Errorf("highlight should extend across the column padding: %q", out)
	}
}

func TestTagPopupCountsRightAligned(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	// Two tags in one column: the long-named tag sizes the column to the
	// 23-column cap, and the short-named tag's counts right-align to the edge.
	tags := []store.TagCount{
		{Name: "aaaa", Total: 12, Unread: 3},
		{Name: "bbbbbbbbbbbbbbbbbbbbbb", Total: 5, Unread: 5},
	}
	s.SetTags(tags)
	_, _, colWidths, _ := s.tagGrid()
	cell := ansi.Strip(s.tagCell(0, colWidths[0]))
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
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(5))
	out := ansi.Strip(s.renderTagPopup())
	lines := strings.Split(out, "\n")
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
	s, _ := newTestState(t)
	s.SetKind(KindTagsArticle)
	s.SetTags(makeTags(5))
	lines := strings.Split(ansi.Strip(s.renderTagPopup()), "\n")
	if !strings.Contains(lines[0], "Tags & Sources") {
		t.Errorf("article-view popup title should be Tags & Sources: %q", lines[0])
	}
}

func TestTagSubtitleCellCentered(t *testing.T) {
	s, _ := newTestState(t)
	cases := []struct {
		width int
		want  string
	}{
		{12, "    Tags    "}, // 4-wide label, even padding
		{11, "   Tags    "},  // odd padding: left gets the smaller half
		{4, "Tags"},          // exact fit, no padding
	}
	for _, c := range cases {
		if got := ansi.Strip(s.tagSubtitleCell(tagItem{label: "Tags", tagIdx: -1}, c.width)); got != c.want {
			t.Errorf("tagSubtitleCell(Tags, %d) = %q, want %q", c.width, got, c.want)
		}
	}
}

func TestTagPopupSubtitlesGroupTheTags(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags([]store.TagCount{
		{Name: "nytimes", Total: 3, Unread: 1, IsSource: true},
		{Name: "theguardian", Total: 2, Unread: 0, IsSource: true},
		{Name: "tech", Total: 5, Unread: 2},
		{Name: "news", Total: 4, Unread: 1},
	})
	out := ansi.Strip(s.renderTagPopup())
	lines := strings.Split(out, "\n")
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
		t.Fatalf("missing Sources subtitle: %q", out)
	}
	tagsRow := rowOf("Tags")
	if tagsRow < 0 {
		t.Fatalf("missing Tags subtitle: %q", out)
	}
	nyRow := rowOf("nytimes")
	if nyRow < 0 {
		t.Fatalf("missing source tag: %q", out)
	}
	techRow := rowOf("tech")
	if techRow < 0 {
		t.Fatalf("missing category tag: %q", out)
	}
	// Sources come first (column-major): the Sources subtitle sits above the
	// source tags, the Tags subtitle above the category tags.
	if !(srcRow < nyRow && nyRow < tagsRow && tagsRow < techRow) {
		t.Errorf("subtitle/group order wrong (Sources %d, nytimes %d, Tags %d, tech %d): %q",
			srcRow, nyRow, tagsRow, techRow, out)
	}
	// A tags-only popup still labels its section.
	s.SetTags([]store.TagCount{{Name: "tech", Total: 1, Unread: 1}})
	out = ansi.Strip(s.renderTagPopup())
	if !strings.Contains(strings.Split(out, "\n")[1], "Tags") {
		t.Errorf("category-only popup should show the Tags subtitle: %q", out)
	}
}

func TestTagPopupSubtitlesNotSelectable(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags([]store.TagCount{
		{Name: "nytimes", Total: 3, Unread: 1, IsSource: true},
		{Name: "tech", Total: 5, Unread: 2},
	})
	// The cursor starts on the first tag (index 0); cycling down and up wraps
	// within the tag list, never landing on a subtitle.
	for i := 0; i < 40; i++ {
		s.movePopupCursor(1)
		if s.cursor < 0 || s.cursor >= len(s.tags) {
			t.Fatalf("cursor %d escaped the tag list", s.cursor)
		}
	}
	// Horizontal moves skip subtitles too.
	s.SetCursor(len(s.tags) - 1)
	s.moveTagCursorHorizontal(-1)
	if s.cursor < 0 || s.cursor >= len(s.tags) {
		t.Fatalf("horizontal cursor %d escaped the tag list", s.cursor)
	}
	// Subtitles render in the status-bar blue role without the selection
	// background and without bold.
	forceTrueColor(t)
	s.SetCursor(0)
	out := s.renderTagPopup()
	bgEscape := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#2c2c2c")).Render("x"))
	blueEscape := escapePrefix(lipgloss.NewStyle().Foreground(s.palette.StatusBar).Render("x"))
	subLine := ""
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "Sources") {
			subLine = l
		}
	}
	if subLine == "" {
		t.Fatalf("missing Sources subtitle: %q", out)
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
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags([]store.TagCount{
		{Name: "news", Total: 60, Unread: 15},
		{Name: "sports", Total: 10, Unread: 0},
	})
	s.SetCursor(1) // keep the counts cell unselected so colors render plainly
	textEscape := escapePrefix(lipgloss.NewStyle().Foreground(s.palette.Text).Render("x"))
	dimEscape := escapePrefix(lipgloss.NewStyle().Foreground(s.palette.Dim).Render("x"))
	newsLine := ""
	for _, l := range strings.Split(s.renderTagPopup(), "\n") {
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
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(5))
	s.SetCursor(0)
	lines := strings.Split(s.renderTagPopup(), "\n")
	top := lines[0]
	bottom := lines[len(lines)-1]
	chromeEscape := escapePrefix(lipgloss.NewStyle().Foreground(s.palette.StatusBar).Render("x"))
	if !strings.HasPrefix(top, chromeEscape) {
		t.Fatalf("top border should use the chrome role: %q", top)
	}
	if !strings.HasPrefix(bottom, chromeEscape) {
		t.Errorf("bottom border should match the rest of the popup border (chrome role): %q", bottom)
	}
}

func TestTagPopupBottomBorderPosition(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindTagsList)
	s.SetTags(makeTags(40))
	s.SetCursor(19) // 20th of 40 → 50%
	lines := strings.Split(ansi.Strip(s.renderTagPopup()), "\n")
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

func TestLinkPopupRowsShowTextAndDimURL(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindLinks)
	s.SetLinks([]Link{
		{Text: "one", URL: "https://example.com/one"},
		{Text: "two", URL: "https://example.com/two"},
	})
	s.SetCursor(1)

	out := s.renderLinksPopup()
	if !strings.Contains(out, "Links") {
		t.Errorf("popup missing title: %q", out)
	}
	stripped := ansi.Strip(out)
	for _, want := range []string{"one", "two", "https://example.com/one", "https://example.com/two"} {
		if !strings.Contains(stripped, want) {
			t.Errorf("popup missing %q: %q", want, stripped)
		}
	}
	if strings.Contains(stripped, "post") {
		t.Errorf("popup should not list the article's own URL: %q", stripped)
	}
	if !strings.Contains(out, "> ") {
		t.Errorf("popup missing the cursor marker: %q", out)
	}
}

func TestLinkPopupNavigationWraps(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindLinks)
	s.SetLinks([]Link{
		{Text: "one", URL: "https://example.com/one"},
		{Text: "two", URL: "https://example.com/two"},
		{Text: "three", URL: "https://example.com/three"},
	})
	n := len(s.links)
	keys := popupKeys(t)

	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, keys)
	if s.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", s.cursor)
	}
	for i := 0; i < n-1; i++ {
		s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, keys)
	}
	if s.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyUp}, keys)
	if s.cursor != n-1 {
		t.Errorf("cursor should wrap up to %d, got %d", n-1, s.cursor)
	}
}

func TestLinkPopupIgnoresGridKeys(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindLinks)
	s.SetLinks([]Link{
		{Text: "one", URL: "https://example.com/one"},
		{Text: "two", URL: "https://example.com/two"},
		{Text: "three", URL: "https://example.com/three"},
	})
	s.SetCursor(1)
	keys := popupKeys(t)
	gridKeys := []tea.KeyMsg{
		{Type: tea.KeyLeft},
		{Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune("h")},
		{Type: tea.KeyRunes, Runes: []rune("l")},
		{Type: tea.KeyRunes, Runes: []rune("1")},
		{Type: tea.KeyRunes, Runes: []rune("g")},
		{Type: tea.KeyRunes, Runes: []rune("G")},
	}
	for _, key := range gridKeys {
		s.Update(key, keys)
		if s.cursor != 1 {
			t.Fatalf("grid key %v changed the links cursor to %d", key, s.cursor)
		}
	}
	// Up/down navigation still works.
	s.Update(tea.KeyMsg{Type: tea.KeyDown}, keys)
	if s.cursor != 2 {
		t.Errorf("down should move the links cursor to 2, got %d", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyUp}, keys)
	if s.cursor != 1 {
		t.Errorf("up should move the links cursor back to 1, got %d", s.cursor)
	}
}

func TestHelpListsLinkPopup(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindHelp)
	out := ansi.Strip(s.renderHelp())
	if !strings.Contains(out, "Link Popup") {
		t.Errorf("help should list the link_popup action: %q", out)
	}
	if !strings.Contains(out, "l, right") {
		t.Errorf("help should show the link_popup keys: %q", out)
	}
}

func TestHelpListsEveryActionGrouped(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindHelp)
	help := ansi.Strip(s.renderHelp())
	for _, a := range config.AllActions() {
		keys := strings.Join(displayKeys(s.sess.Config().Keybindings[a]), ", ")
		found := false
		for _, l := range strings.Split(help, "\n") {
			if strings.Contains(l, s.helpLabel(a)) && strings.Contains(l, keys) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("help missing row for %q (%s %s)", a, s.helpLabel(a), keys)
		}
	}
	for _, want := range []string{"Global", "List view", "Article view"} {
		if !strings.Contains(help, want) {
			t.Errorf("help missing section %q", want)
		}
	}
	gi := strings.Index(help, "Global")
	li := strings.Index(help, "List view")
	ai := strings.Index(help, "Article view")
	if gi >= li || li >= ai {
		t.Errorf("help sections out of order: Global=%d List view=%d Article view=%d", gi, li, ai)
	}
}

func TestHelpKeysStandOutInBright(t *testing.T) {
	forceTrueColor(t)
	s, _ := newTestState(t)
	s.SetKind(KindHelp)
	help := s.renderHelp()
	if !strings.Contains(help, "\x1b[97mctrl+c\x1b[0m") {
		t.Errorf("keys should render in the bright role (ANSI 15): %q", help)
	}
	if !strings.Contains(help, "\x1b[1;94m Key Bindings \x1b[0m") {
		t.Errorf("Key Bindings title should render bold in the chrome role (ANSI 12): %q", help)
	}
	// Section headings stay in the chrome role but are no longer bold.
	if !strings.Contains(help, "\x1b[94mGlobal\x1b[0m") {
		t.Errorf("section headings should render plain in the chrome role: %q", help)
	}
	for _, l := range strings.Split(help, "\n") {
		if strings.Contains(l, "ctrl+c") {
			if !strings.Contains(l, "\x1b[0m, \x1b[97m") {
				t.Errorf("commas between keys should stay uncolored: %q", l)
			}
			break
		}
	}
}

func TestHelpShowsSpaceKeyAsWord(t *testing.T) {
	// Load normalizes "space" to a literal " " for runtime matching; the help
	// popup must still display it as the word "space".
	km, err := config.ParseKeybindings(config.DefaultKeybindings())
	if err != nil {
		t.Fatalf("ParseKeybindings: %v", err)
	}
	cfg := config.Default()
	cfg.Keybindings = km
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	s := New(app.New(cfg, st), render.DarkPalette(), false, 80, 24)
	s.SetKind(KindHelp)
	help := ansi.Strip(s.renderHelp())
	if !strings.Contains(help, "pgdn, ctrl+f, space") {
		t.Errorf("help should render the space key as the word space: %q", help)
	}
}

func TestHelpTwoColumnWithinSeventyColumns(t *testing.T) {
	s, _ := newTestState(t)
	s.SetKind(KindHelp)
	lines := strings.Split(s.renderHelp(), "\n")
	if len(lines) < 2 {
		t.Fatalf("help should have content lines, got %d", len(lines))
	}
	// List view must sit in the second column, on the same row as Global.
	row := ""
	for _, l := range lines {
		if strings.Contains(l, "Global") {
			row = l
			break
		}
	}
	if row == "" {
		t.Fatalf("help should render a Global row")
	}
	if i := strings.Index(row, "Global"); i < 0 || strings.Index(row, "List view") <= i {
		t.Errorf("List view should be to the right of Global: %q", row)
	}
	// The popup including its border must not exceed 70 columns.
	for i, l := range lines {
		if w := ansi.StringWidth(l); w > 70 {
			t.Errorf("help line %d is %d columns, want <= 70: %q", i, w, l)
		}
	}
}
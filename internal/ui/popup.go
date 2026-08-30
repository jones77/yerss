package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"yerss/internal/config"
	"yerss/internal/store"
)

type popupState struct {
	tags      []store.TagCount
	links     []articleLink
	cursor    int
	scrollCol int
}

func (m *Model) openTagPopup(mode popupMode) {
	tags, err := m.store.ListTags()
	if err != nil {
		m.setStatus("tags error: " + err.Error())
		return
	}
	if mode == popupTagsArticle {
		want := map[string]bool{}
		for _, c := range m.article.article.Categories {
			want[c] = true
		}
		filtered := make([]store.TagCount, 0, len(want))
		for _, t := range tags {
			if want[t.Name] {
				filtered = append(filtered, t)
			}
		}
		tags = filtered
	}
	m.popup = mode
	m.popupData = popupState{tags: tags}
}

func (m *Model) updatePopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.popup == popupHelp {
		switch msg.String() {
		case "q", "esc", "enter", " ":
			m.popup = noPopup
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}
	if msg.String() == "enter" {
		if m.popup == popupLinks {
			return m.confirmLinkSelection()
		}
		return m.confirmTagSelection()
	}
	act, ok := m.keys[config.ViewPopup][msg.String()]
	if !ok {
		return m, nil
	}
	switch act {
	case config.Quit:
		m.persistSelection()
		return m, tea.Quit
	case config.Back:
		m.popup = noPopup
	case config.MoveDown:
		m.movePopupCursor(1)
	case config.MoveUp:
		m.movePopupCursor(-1)
	case config.MoveLeft:
		if m.isTagPopup() {
			m.moveTagCursorHorizontal(-1)
		}
	case config.MoveRight:
		if m.isTagPopup() {
			m.moveTagCursorHorizontal(1)
		}
	case config.Top:
		if m.isTagPopup() {
			m.jumpTagCursor(true)
		}
	case config.Bottom:
		if m.isTagPopup() {
			m.jumpTagCursor(false)
		}
	case config.OpenURL:
		if (m.popup == popupLinks || m.popup == popupTagsArticle) && m.article.article.Link != "" {
			return m, openURLCmd(m.article.article.Link)
		}
	}
	return m, nil
}

// openLinksPopup opens the links popup over the article view, seeded with the
// links harvested from the current article's markdown source.
func (m *Model) openLinksPopup() {
	m.popup = popupLinks
	m.popupData = popupState{links: m.article.links}
}

// confirmLinkSelection closes the links popup and opens the selected URL in
// the system browser. An empty popup (an article without links) just closes.
func (m *Model) confirmLinkSelection() (tea.Model, tea.Cmd) {
	links := m.popupData.links
	m.popup = noPopup
	if m.popupData.cursor < 0 || m.popupData.cursor >= len(links) {
		return m, nil
	}
	return m, openURLCmd(links[m.popupData.cursor].url)
}

func (m *Model) confirmTagSelection() (tea.Model, tea.Cmd) {
	if len(m.popupData.tags) == 0 {
		m.popup = noPopup
		return m, nil
	}
	t := m.popupData.tags[m.popupData.cursor]
	m.popup = noPopup
	m.list.filter = t.Name
	m.view = viewList
	m.article.sel = textSelection{}
	m.loadList()
	return m, nil
}

func (m *Model) movePopupCursor(delta int) {
	n := len(m.popupData.tags)
	if m.popup == popupLinks {
		n = len(m.popupData.links)
	}
	if n == 0 {
		return
	}
	m.popupData.cursor = wrapIndex(m.popupData.cursor, delta, n)
}

// isTagPopup reports whether the current popup is one of the tag popups, which
// use the column-major grid navigation. The links popup stays single-column
// and linear.
func (m *Model) isTagPopup() bool {
	return m.popup == popupTagsList || m.popup == popupTagsArticle
}

// moveTagCursorHorizontal moves the tag popup cursor one column left or right,
// wrapping around the grid's horizontal edges and clamping into a final column
// that holds fewer tags than the full columns. Vertical movement stays on
// movePopupCursor: in column-major order the next index is exactly "one row
// down, or the top of the next column", so wrapIndex already wraps both ends.
func (m *Model) moveTagCursorHorizontal(delta int) {
	rows, cols, _, _ := m.tagGrid()
	n := len(m.popupData.tags)
	if n == 0 || cols <= 1 {
		return
	}
	items := m.tagItems()
	disp := m.tagDispPos(m.popupData.cursor)
	col := disp / rows
	row := disp % rows
	col = wrapIndex(col, delta, cols)
	if idx := m.tagAtColumnRow(col, row, items, rows); idx >= 0 {
		m.popupData.cursor = idx
	}
}

// tagAtColumnRow returns the tag index at the item-grid cell in column c,
// row r, landing on the nearest tag in that column when the cell holds a
// section subtitle or lies past the column's last item (a shorter final
// column), preferring the tag just below. Returns -1 when the column holds no
// tags.
func (m *Model) tagAtColumnRow(c, r int, items []tagItem, rows int) int {
	d := len(items)
	pos := c*rows + r
	if pos >= 0 && pos < d && items[pos].tagIdx >= 0 {
		return items[pos].tagIdx
	}
	start := c * rows
	end := min(d, (c+1)*rows)
	for p := max(pos+1, start); p < end; p++ {
		if items[p].tagIdx >= 0 {
			return items[p].tagIdx
		}
	}
	for p := min(pos-1, end-1); p >= start; p-- {
		if items[p].tagIdx >= 0 {
			return items[p].tagIdx
		}
	}
	return -1
}

// jumpTagCursor moves the tag popup cursor to the first (most popular) or last
// (least popular) tag.
func (m *Model) jumpTagCursor(first bool) {
	n := len(m.popupData.tags)
	if n == 0 {
		return
	}
	if first {
		m.popupData.cursor = 0
	} else {
		m.popupData.cursor = n - 1
	}
}

// tagPopupSize returns the tag popup box dimensions: the screen minus a
// 12-column margin on the left and right and a 2-row margin at the top and
// bottom, clamped to a minimal usable box on tiny terminals.
func (m *Model) tagPopupSize() (h, w int) {
	return max(3, m.height-4), max(12, m.width-24)
}

// maxTagColWidth is the widest a tag popup column may be; longer entries have
// their middle elided with the ellipsis.
const maxTagColWidth = 23

// tagColGap is the number of spaces between adjacent tag popup columns.
const tagColGap = 2

// tagItem is one cell in the tag popup grid: either a non-selectable section
// subtitle ("Sources" above the news-organization tags, "Tags" above the
// category tags) or a tag. Subtitles render in the status-bar blue role and
// never carry the selection cursor.
type tagItem struct {
	label  string
	tagIdx int
}

// tagItems lays out the popup's tags in display order, inserting a "Sources"
// subtitle before the source-domain tags and a "Tags" subtitle before the
// category tags, each only when its group is non-empty. Tags keep their
// column-major order in the returned list, so vertical navigation over the
// tags (wrapIndex on the tag count) is unaffected by the subtitles.
func (m *Model) tagItems() []tagItem {
	tags := m.popupData.tags
	items := make([]tagItem, 0, len(tags)+2)
	sourceHeader := false
	tagsHeader := false
	for i, t := range tags {
		if t.IsSource && !sourceHeader {
			items = append(items, tagItem{label: "Sources", tagIdx: -1})
			sourceHeader = true
		} else if !t.IsSource && !tagsHeader {
			items = append(items, tagItem{label: "Tags", tagIdx: -1})
			tagsHeader = true
		}
		items = append(items, tagItem{tagIdx: i})
	}
	return items
}

// tagDispPos returns the item-grid position of the tag with the given index:
// its tag index plus the subtitles laid out before it.
func (m *Model) tagDispPos(tagIdx int) int {
	for pos, it := range m.tagItems() {
		if it.tagIdx == tagIdx {
			return pos
		}
	}
	return tagIdx
}

// tagItemWidth returns the display width of a grid cell: the count-bearing
// entry for a tag, or the subtitle label on its own.
func (m *Model) tagItemWidth(it tagItem) int {
	if it.tagIdx < 0 {
		return ansi.StringWidth(it.label)
	}
	return m.tagEntryWidth(it.tagIdx)
}

// tagGrid computes the tag popup grid geometry: the number of rows (the box
// interior between the borders, with no vertical padding), the number of
// columns needed to hold every tag, each column's width, and the effective box
// width. Rows are
// column-major: tags fill the first
// column top to bottom and continue into the next column. Each column is sized
// to fit its own widest entry (capped at maxTagColWidth) so a single long tag
// name does not widen every column; longer entries are middle-elided and the
// grid may still be wider than the popup interior, so the renderer scrolls a
// horizontal window over it. When the natural grid is narrower than the
// interior, the spare space on the right is used to widen the columns so the
// grid fills the interior (see fillColumns), and the box width follows suit.
func (m *Model) tagGrid() (rows, cols int, colWidths []int, boxW int) {
	h, boxW := m.tagPopupSize()
	textW := max(1, boxW-4)
	items := m.tagItems()
	d := len(items)
	rows = max(1, h-2)
	if d == 0 {
		return rows, 1, nil, boxW
	}
	cols = (d + rows - 1) / rows
	colWidths = make([]int, cols)
	for c := range colWidths {
		colWidths[c] = 1
		for r := 0; r < rows; r++ {
			pos := c*rows + r
			if pos >= d {
				break
			}
			colWidths[c] = max(colWidths[c], m.tagItemWidth(items[pos]))
		}
		colWidths[c] = min(colWidths[c], maxTagColWidth)
	}

	// Fill the whole grid when it fits the interior; a grid wider than the
	// interior keeps its natural widths and the visible window is filled at
	// render time.
	all := make([]int, cols)
	for c := range all {
		all[c] = c
	}
	colWidths, textW = fillColumns(colWidths, all, textW)
	return rows, cols, colWidths, textW + 4
}

// fillColumns widens the given columns so they span textW, keeping tagColGap
// spaces between them: all the columns grow to one equal width when that
// width fits every column's current width, otherwise each column grows by an
// equal share of the spare. Any leftover space that cannot be split into whole
// columns pulls textW in, so the grid never leaves space to spare. The widened
// widths and the reduced textW are returned.
func fillColumns(widths, cols []int, textW int) ([]int, int) {
	out := append([]int(nil), widths...)
	n := len(cols)
	if n == 0 {
		return out, textW
	}
	gaps := (n - 1) * tagColGap
	total := gaps
	needed := 0
	for _, c := range cols {
		total += widths[c]
		needed = max(needed, widths[c])
	}
	if total >= textW {
		return out, textW
	}
	if equalW := (textW - gaps) / n; equalW >= needed {
		for _, c := range cols {
			out[c] = equalW
		}
	} else {
		add := (textW - total) / n
		for _, c := range cols {
			out[c] += add
		}
	}
	used := gaps
	for _, c := range cols {
		used += out[c]
	}
	if rem := textW - used; rem > 0 {
		textW -= rem
	}
	return out, textW
}

// tagEntryWidth returns the display width of the grid cell for tag index idx:
// the tag name, a space, and the count display. Cells have no cursor gutter;
// the box's one-space padding and the single space between columns keep the
// tag one column from the edges.
func (m *Model) tagEntryWidth(idx int) int {
	t := m.popupData.tags[idx]
	plain, bold := tagCountText(t.Unread, t.Total)
	return ansi.StringWidth(t.Name) + 1 + ansi.StringWidth(plain+bold)
}

// tagCell renders one grid cell for tag index idx: the tag name on the left
// and the count display right-aligned to the column edge, with any leftover
// space padding between them. A name too long for the column has its middle
// elided with the ellipsis. The selected cell is highlighted with the list
// view's selection background across the full column width instead of a ">"
// marker, and the count display uses the list view's read/unread colors: the
// unread count bold bright and the total in the text (white) role. The
// background is applied per segment (like the list view) so inner style
// resets cannot drop the highlight from the padding, keeping the highlight
// exactly the column's width for every item in the column.
func (m *Model) tagCell(idx, width int) string {
	t := m.popupData.tags[idx]
	plain, bold := tagCountText(t.Unread, t.Total)
	counts := plain + bold
	countsW := ansi.StringWidth(counts)
	nameW := width - 1 - countsW
	name := t.Name
	if w := ansi.StringWidth(name); w > nameW {
		name = elideMiddle(name, max(1, nameW), m.glyphs().ellipsis)
	}

	selected := idx == m.popupData.cursor
	bg := lipgloss.Color("#707070")
	nameStyle := lipgloss.NewStyle()
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bright)
	plainStyle := lipgloss.NewStyle().Foreground(m.palette.Text)
	padStyle := lipgloss.NewStyle()
	if selected {
		nameStyle = nameStyle.Background(bg)
		boldStyle = boldStyle.Background(bg)
		plainStyle = plainStyle.Background(bg)
		padStyle = padStyle.Background(bg)
	}

	// The name sits on the left; the counts right-align to the column edge,
	// leaving at least one space between them.
	cell := nameStyle.Render(name)
	if gap := width - ansi.StringWidth(name) - countsW; gap > 0 {
		cell += padStyle.Render(strings.Repeat(" ", gap))
	}
	cell += boldStyle.Render(bold) + plainStyle.Render(plain)
	return cell
}

// tagSubtitleCell renders a non-selectable section subtitle ("Sources" or
// "Tags") in the status-bar blue role, non-bold, centered within the column
// width. It never carries the selection highlight and navigation never lands
// on it.
func (m *Model) tagSubtitleCell(it tagItem, width int) string {
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	label := style.Render(it.label)
	if pad := width - ansi.StringWidth(it.label); pad > 0 {
		left := pad / 2
		return strings.Repeat(" ", left) + label + strings.Repeat(" ", pad-left)
	}
	return label
}

// tagPopupWindow returns the indices of the columns visible for the current
// cursor and stored scroll position. The window is anchored by the stored
// scroll column so horizontal movement is symmetric: it advances right only
// when the cursor passes the right edge of the window and left only when it
// passes the left edge, staying put otherwise — the previous cursor-derived
// window scrolled left eagerly on every left move. Wraps, jumps, and resizes
// still snap the window to show the cursor, and the stored scroll column is
// kept in sync here so it reflects the last render.
func (m *Model) tagPopupWindow(colWidths []int, rows, textW int) []int {
	blockStart := make([]int, len(colWidths))
	off := 0
	for c, cw := range colWidths {
		blockStart[c] = off
		off += cw + tagColGap
	}
	cursorCol := m.tagDispPos(m.popupData.cursor) / rows

	// visibleFrom returns how many columns fit starting at column start.
	visibleFrom := func(start int) int {
		if start < 0 || start >= len(colWidths) {
			return 0
		}
		edge := blockStart[start] + textW
		n := 0
		for c := start; c < len(colWidths); c++ {
			if blockStart[c]+colWidths[c] > edge {
				break
			}
			n++
		}
		return n
	}

	scrollCol := m.popupData.scrollCol
	vis := visibleFrom(scrollCol)
	if vis == 0 {
		scrollCol = cursorCol
		vis = visibleFrom(scrollCol)
	}
	if cursorCol < scrollCol {
		scrollCol = cursorCol
	} else if vis > 0 && cursorCol >= scrollCol+vis {
		scrollCol = max(0, cursorCol-vis+1)
	}
	// Clamp the window to the grid's right edge (e.g. after a resize).
	for scrollCol > 0 && scrollCol+visibleFrom(scrollCol) > len(colWidths) {
		scrollCol--
	}
	m.popupData.scrollCol = scrollCol

	edge := blockStart[scrollCol] + textW
	var visible []int
	for c := scrollCol; c < len(colWidths); c++ {
		if blockStart[c]+colWidths[c] > edge {
			break
		}
		visible = append(visible, c)
	}
	return visible
}

func (m *Model) renderTagPopup() string {
	h, _ := m.tagPopupSize()
	title := "Tags & Sources"
	g := m.glyphs()
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	v := style.Render(g.v)
	pad := " "

	// The title is inlined into the top border and the position indicator into
	// the bottom border, so the grid fills the interior between the borders
	// (one space of left/right padding only). Section subtitles ("Sources" and
	// "Tags") are non-selectable cells laid out in the same column-major grid.
	// The box width comes from the grid so it can shrink to avoid a trailing
	// gap once the columns fill the interior.
	rows, _, colWidths, boxW := m.tagGrid()
	textW := max(1, boxW-4)
	items := m.tagItems()
	d := len(items)
	gridLines := min(rows, max(0, h-2))
	var body []string
	var renderWidths []int
	if d > 0 {
		visible := m.tagPopupWindow(colWidths, rows, textW)
		// Fill the visible window so it spans the interior: a no-op when the
		// whole grid already fits, and the way a grid wider than the interior
		// avoids a trailing gap in its horizontal window. The box width follows
		// any right-edge pull-in.
		renderWidths, textW = fillColumns(colWidths, visible, textW)
		boxW = textW + 4
		for r := 0; r < gridLines; r++ {
			var rowCells []string
			for _, c := range visible {
				pos := c*rows + r
				if pos >= d {
					break
				}
				it := items[pos]
				if it.tagIdx < 0 {
					rowCells = append(rowCells, m.tagSubtitleCell(it, renderWidths[c]))
				} else {
					rowCells = append(rowCells, m.tagCell(it.tagIdx, renderWidths[c]))
				}
			}
			body = append(body, strings.Join(rowCells, strings.Repeat(" ", tagColGap)))
		}
	}
	for len(body) < gridLines {
		body = append(body, "")
	}

	out := []string{inlineTitleBorder(title, boxW, g, m.palette)}
	for _, l := range body {
		out = append(out, v+pad+padRight(truncate(l, textW), textW)+pad+v)
	}
	out = append(out, m.tagBottomBorder(boxW, g))
	return strings.Join(out, "\n")
}

// tagBottomBorder renders the tag popup's bottom edge, right-aligning the
// position indicator `<percent>% · <nth>/<total>` — the selected tag's place
// among all tags — mirroring the article view's bottom border and the list
// view's status bar, with dash fill to the left. The line and the indicator
// render in the same chrome role as the rest of the popup border.
func (m *Model) tagBottomBorder(w int, g borderGlyphs) string {
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	textStyle := lipgloss.NewStyle().Foreground(m.palette.StatusBar)

	n := len(m.popupData.tags)
	pos := m.popupData.cursor + 1
	if pos < 0 || n == 0 {
		pos = 0
	}
	if pos > n {
		pos = n
	}
	pct := 0
	if n > 0 {
		pct = int(float64(pos)/float64(n)*100 + 0.5)
	}
	percentStr := fmt.Sprintf("%d%%", pct)
	ratioStr := fmt.Sprintf("%d/%d", pos, n)
	iw := ansi.StringWidth(percentStr) + 1 + ansi.StringWidth(g.bullet) + 1 + ansi.StringWidth(ratioStr)
	fill := max(0, w-6-iw)
	dash := style.Render(g.h)
	return style.Render(g.bl) + dash + style.Render(strings.Repeat(g.h, fill)) + " " +
		textStyle.Render(percentStr) + " " + style.Render(g.bullet) + " " + textStyle.Render(ratioStr) +
		" " + dash + style.Render(g.br)
}

// renderLinksPopup renders the article links popup: each row shows the link
// text followed by the URL in the dim style, truncated to the box width.
func (m *Model) renderLinksPopup() string {
	h := m.height * 6 / 10
	w := m.width * 8 / 10
	if h < 3 {
		h = 3
	}
	if w < 12 {
		w = 12
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(m.palette.StatusBar).Render("Links"))
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
	for i, l := range m.popupData.links {
		cursor := "  "
		if i == m.popupData.cursor {
			cursor = "> "
		}
		// text + separator + dim URL, truncated to the interior width.
		innerW := w - 6
		url := ansi.Truncate(l.url, max(1, innerW-ansi.StringWidth(l.text)-3), m.glyphs().ellipsis)
		line := l.text + dim.Render(" · "+url)
		lines = append(lines, cursor+ansi.Truncate(line, innerW, ""))
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.StatusBar).
		Width(w).
		Height(h).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}

// tagCountText splits the tag count display into a plain and a bold portion.
// All-unread and all-read tags collapse to a single total; mixed tags show
// unread/total with the unread count bold. No brackets are shown around the
// counts.
func tagCountText(unread, total int) (plain, bold string) {
	if unread == 0 {
		return fmt.Sprintf("%d", total), ""
	}
	if total == unread {
		return "", fmt.Sprintf("%d", total)
	}
	return fmt.Sprintf("/%d", total), fmt.Sprintf("%d", unread)
}

// actionLabel renders an action identifier as a display label: underscores
// become spaces and each word is capitalized (e.g. open_article -> Open
// Article).
func actionLabel(a config.Action) string {
	return cases.Title(language.Und).String(strings.ReplaceAll(string(a), "_", " "))
}

// helpLabel renders an action label for the help popup, replacing the "Half"
// prefix of the half-page actions with the half glyph (½, or 1/2 in ASCII
// mode).
func (m *Model) helpLabel(a config.Action) string {
	label := actionLabel(a)
	if strings.HasPrefix(label, "Half") {
		return m.glyphs().half + strings.TrimPrefix(label, "Half")
	}
	return label
}

// displayKeys renders normalized key strings for the help popup: a literal
// space key (as stored by normalizeKey for runtime matching) is shown as the
// word "space", and the normalized "pgdown" spelling is shown as "pgdn".
func displayKeys(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		switch k {
		case " ":
			out[i] = "space"
		case "pgdown":
			out[i] = "pgdn"
		default:
			out[i] = k
		}
	}
	return out
}

func (m *Model) renderHelp() string {
	heading := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	keyStyle := lipgloss.NewStyle().Foreground(m.palette.Bright)

	// actionRow renders one binding with its label padded to labelW+2 columns
	// so the keys align across the section, leaving at least two spaces between
	// the label and its bound keys. Each bound key renders in the bright role
	// so it stands out; the commas between keys stay in the default text role.
	actionRow := func(a config.Action, labelW int) string {
		keys := displayKeys(m.cfg.Keybindings[a])
		if len(keys) == 0 {
			return ""
		}
		styled := make([]string, len(keys))
		for i, k := range keys {
			styled[i] = keyStyle.Render(k)
		}
		return fmt.Sprintf("%-*s%s", labelW+2, m.helpLabel(a), strings.Join(styled, ", "))
	}

	// section renders a heading followed by its action rows, aligning the keys
	// under the widest label in the section.
	section := func(label string, actions []config.Action) []string {
		labelW := 0
		for _, a := range actions {
			labelW = max(labelW, ansi.StringWidth(m.helpLabel(a)))
		}
		lines := []string{heading.Render(label)}
		for _, a := range actions {
			if r := actionRow(a, labelW); r != "" {
				lines = append(lines, r)
			}
		}
		return lines
	}

	// Global fills the left column; List view and Article view stack in the
	// right column, separated by a blank line.
	left := section("Global", config.GlobalActions())
	right := section("List view", config.ListActions())
	right = append(right, "")
	right = append(right, section("Article view", config.ArticleActions())...)

	n := max(len(left), len(right))
	for len(left) < n {
		left = append(left, "")
	}
	for len(right) < n {
		right = append(right, "")
	}

	leftW := 0
	for _, l := range left {
		leftW = max(leftW, ansi.StringWidth(l))
	}
	rightW := 0
	for _, l := range right {
		rightW = max(rightW, ansi.StringWidth(l))
	}

	// The popup including its border must not exceed 70 columns. The box keeps
	// one space of padding on each side.
	gap := 2
	boxW := min(70, m.width, leftW+gap+rightW+4)
	textW := max(1, boxW-4)

	g := m.glyphs()
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	v := style.Render(g.v)
	pad := " "

	var content []string
	for i := 0; i < n; i++ {
		line := padRight(left[i], leftW) + strings.Repeat(" ", gap) + padRight(right[i], rightW)
		content = append(content, padRight(truncate(line, textW), textW))
	}

	// The top edge inlines the "Key Bindings" title; one blank row pads the
	// top and bottom of the content.
	blank := v + pad + strings.Repeat(" ", textW) + pad + v
	rows := []string{inlineTitleBorder("Key Bindings", boxW, g, m.palette), blank}
	for _, l := range content {
		rows = append(rows, v+pad+l+pad+v)
	}
	rows = append(rows, blank)
	rows = append(rows, style.Render(g.bl+strings.Repeat(g.h, boxW-2)+g.br))
	return strings.Join(rows, "\n")
}

// inlineTitleBorder renders a popup's top edge as `┌─ Title ───┐`,
// centering the title in the bold chrome role and keeping the rails in the
// chrome role.
func inlineTitleBorder(title string, w int, g borderGlyphs, p Palette) string {
	style := lipgloss.NewStyle().Foreground(p.StatusBar)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.StatusBar)
	leftRail := g.tl + g.h
	rightRail := g.h + g.tr
	core := " " + title + " "
	fill := max(0, w-ansi.StringWidth(leftRail)-ansi.StringWidth(core)-ansi.StringWidth(rightRail))
	left := fill / 2
	right := fill - left
	return style.Render(leftRail+strings.Repeat(g.h, left)) +
		titleStyle.Render(core) +
		style.Render(strings.Repeat(g.h, right)+rightRail)
}

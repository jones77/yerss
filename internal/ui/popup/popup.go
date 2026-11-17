// Package popup renders and handles the modal overlays that float over the
// list and article views: the tag grid, the article links list, and the key
// binding help. It is a pure view component: it knows nothing of the Model,
// the list, or the article, and reports coarse actions for the Model to act
// on.
package popup

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal"
	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/ui/render"
)

// Kind identifies which popup is open. KindNone means no popup is showing.
type Kind int

const (
	KindNone Kind = iota
	// KindTagsList is the tag grid opened over the list view.
	KindTagsList
	// KindTagsArticle is the tag grid opened over the article view, limited
	// to the open article's own categories.
	KindTagsArticle
	// KindHelp is the key binding help popup.
	KindHelp
	// KindLinks is the current article's links popup.
	KindLinks
)

// Action reports how the Model should react to a popup key message.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionBack
	ActionOpenURL
	ActionConfirm
)

// Link is one hyperlink shown in the links popup.
type Link struct {
	Text string
	URL  string
}

// State holds the popup's data and rendering/input state.
type State struct {
	kind Kind

	sess    *app.Session
	styles  styles
	palette render.Palette
	ascii   bool

	width  int
	height int

	tags      []internal.TagCount
	links     []Link
	cursor    int
	scrollCol int
}

// New builds a popup State bound to a session.
func New(sess *app.Session, palette render.Palette, ascii bool, width, height int) *State {
	return &State{
		sess:    sess,
		styles:  newStyles(palette),
		palette: palette,
		ascii:   ascii,
		width:   width,
		height:  height,
	}
}

// SetSize updates the terminal dimensions the popup lays out against.
func (s *State) SetSize(width, height int) { s.width, s.height = width, height }

// Kind returns the currently open popup.
func (s *State) Kind() Kind { return s.kind }

// SetKind switches the open popup (or closes it with KindNone).
func (s *State) SetKind(k Kind) { s.kind = k }

// Cursor returns the selected item's index.
func (s *State) Cursor() int { return s.cursor }

// SetCursor positions the selection cursor.
func (s *State) SetCursor(c int) { s.cursor = c }

// Tags returns the popup's tag counts.
func (s *State) Tags() []internal.TagCount { return s.tags }

// SetTags replaces the popup's tag counts.
func (s *State) SetTags(tags []internal.TagCount) { s.tags = tags }

// Links returns the popup's links.
func (s *State) Links() []Link { return s.links }

// SetLinks replaces the popup's links.
func (s *State) SetLinks(links []Link) { s.links = links }

// OpenTags loads the session's tag counts into the popup and switches to the
// given tag popup kind. categories limits the tags to the named categories
// when non-empty (the article-view variant); an empty slice keeps all tags.
// It returns a status message on error.
func (s *State) OpenTags(kind Kind, categories []string) string {
	tags, err := s.sess.ListTags()
	if err != nil {
		return "tags error: " + err.Error()
	}
	if kind == KindTagsArticle {
		want := map[string]bool{}
		for _, c := range categories {
			want[c] = true
		}
		filtered := make([]internal.TagCount, 0, len(want))
		for _, t := range tags {
			if want[t.Name] {
				filtered = append(filtered, t)
			}
		}
		tags = filtered
	}
	s.kind = kind
	s.tags = tags
	return ""
}

// OpenLinks switches to the links popup seeded with the given links.
func (s *State) OpenLinks(links []Link) {
	s.kind = KindLinks
	s.links = links
}

// SelectedTag returns the tag at the cursor, if any.
func (s *State) SelectedTag() (internal.TagCount, bool) {
	if len(s.tags) == 0 || s.cursor < 0 || s.cursor >= len(s.tags) {
		return internal.TagCount{}, false
	}
	return s.tags[s.cursor], true
}

// SelectedURL returns the URL of the link at the cursor, if any.
func (s *State) SelectedURL() string {
	if s.cursor < 0 || s.cursor >= len(s.links) {
		return ""
	}
	return s.links[s.cursor].URL
}

// Update processes a key message against the popup and returns the action the
// Model should take. Keys bound in the popup view map (keys) drive navigation
// and selection.
func (s *State) Update(msg tea.KeyMsg, keys map[string]config.Action) Action {
	if s.kind == KindHelp {
		switch msg.String() {
		case "q", "esc", "enter", " ":
			s.kind = KindNone
			return ActionBack
		case "ctrl+c":
			return ActionQuit
		}
		return ActionNone
	}
	if msg.String() == "enter" {
		return ActionConfirm
	}
	act, ok := keys[msg.String()]
	if !ok {
		return ActionNone
	}
	switch act {
	case config.Quit:
		return ActionQuit
	case config.Back:
		s.kind = KindNone
		return ActionBack
	case config.MoveDown:
		s.movePopupCursor(1)
	case config.MoveUp:
		s.movePopupCursor(-1)
	case config.MoveLeft:
		if s.isTagPopup() {
			s.moveTagCursorHorizontal(-1)
		}
	case config.MoveRight:
		if s.isTagPopup() {
			s.moveTagCursorHorizontal(1)
		}
	case config.Top:
		if s.isTagPopup() {
			s.jumpTagCursor(true)
		}
	case config.Bottom:
		if s.isTagPopup() {
			s.jumpTagCursor(false)
		}
	case config.OpenURL:
		if s.kind == KindLinks || s.kind == KindTagsArticle {
			return ActionOpenURL
		}
	}
	return ActionNone
}

// Render returns the current popup's frame, or "" when no popup is open.
func (s *State) Render() string {
	switch s.kind {
	case KindTagsList, KindTagsArticle:
		return s.renderTagPopup()
	case KindLinks:
		return s.renderLinksPopup()
	case KindHelp:
		return s.renderHelp()
	}
	return ""
}

func (s *State) movePopupCursor(delta int) {
	n := len(s.tags)
	if s.kind == KindLinks {
		n = len(s.links)
	}
	if n == 0 {
		return
	}
	s.cursor = render.WrapIndex(s.cursor, delta, n)
}

// isTagPopup reports whether the current popup is one of the tag popups, which
// use the column-major grid navigation. The links popup stays single-column
// and linear.
func (s *State) isTagPopup() bool {
	return s.kind == KindTagsList || s.kind == KindTagsArticle
}

// moveTagCursorHorizontal moves the tag popup cursor one column left or right,
// wrapping around the grid's horizontal edges and clamping into a final column
// that holds fewer tags than the full columns. Vertical movement stays on
// movePopupCursor: in column-major order the next index is exactly "one row
// down, or the top of the next column", so wrapIndex already wraps both ends.
func (s *State) moveTagCursorHorizontal(delta int) {
	rows, cols, _, _ := s.tagGrid()
	n := len(s.tags)
	if n == 0 || cols <= 1 {
		return
	}
	items := s.tagItems()
	disp := s.tagDispPos(s.cursor)
	col := disp / rows
	row := disp % rows
	col = render.WrapIndex(col, delta, cols)
	if idx := s.tagAtColumnRow(col, row, items, rows); idx >= 0 {
		s.cursor = idx
	}
}

// tagAtColumnRow returns the tag index at the item-grid cell in column c,
// row r, landing on the nearest tag in that column when the cell holds a
// section subtitle or lies past the column's last item (a shorter final
// column), preferring the tag just below. Returns -1 when the column holds no
// tags.
func (s *State) tagAtColumnRow(c, r int, items []tagItem, rows int) int {
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
func (s *State) jumpTagCursor(first bool) {
	n := len(s.tags)
	if n == 0 {
		return
	}
	if first {
		s.cursor = 0
	} else {
		s.cursor = n - 1
	}
}

// tagPopupSize returns the tag popup box dimensions: the screen minus a
// 12-column margin on the left and right and a 2-row margin at the top and
// bottom, clamped to a minimal usable box on tiny terminals.
func (s *State) tagPopupSize() (h, w int) {
	return max(3, s.height-4), max(12, s.width-24)
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
func (s *State) tagItems() []tagItem {
	tags := s.tags
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
func (s *State) tagDispPos(tagIdx int) int {
	for pos, it := range s.tagItems() {
		if it.tagIdx == tagIdx {
			return pos
		}
	}
	return tagIdx
}

// tagItemWidth returns the display width of a grid cell: the count-bearing
// entry for a tag, or the subtitle label on its own.
func (s *State) tagItemWidth(it tagItem) int {
	if it.tagIdx < 0 {
		return ansi.StringWidth(it.label)
	}
	return s.tagEntryWidth(it.tagIdx)
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
func (s *State) tagGrid() (rows, cols int, colWidths []int, boxW int) {
	h, boxW := s.tagPopupSize()
	textW := max(1, boxW-4)
	items := s.tagItems()
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
			colWidths[c] = max(colWidths[c], s.tagItemWidth(items[pos]))
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
func (s *State) tagEntryWidth(idx int) int {
	t := s.tags[idx]
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
func (s *State) tagCell(idx, width int) string {
	t := s.tags[idx]
	plain, bold := tagCountText(t.Unread, t.Total)
	counts := plain + bold
	countsW := ansi.StringWidth(counts)
	nameW := width - 1 - countsW
	name := t.Name
	if w := ansi.StringWidth(name); w > nameW {
		name = render.ElideMiddle(name, max(1, nameW), s.glyphs().Ellipsis)
	}

	selected := idx == s.cursor
	nameStyle := s.styles.plain
	boldStyle := s.styles.bright
	plainStyle := s.styles.text
	padStyle := s.styles.plain
	if selected {
		nameStyle = s.styles.selRow
		boldStyle = s.styles.selBright
		plainStyle = s.styles.selText
		padStyle = s.styles.selRow
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
func (s *State) tagSubtitleCell(it tagItem, width int) string {
	style := s.styles.status
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
func (s *State) tagPopupWindow(colWidths []int, rows, textW int) []int {
	blockStart := make([]int, len(colWidths))
	off := 0
	for c, cw := range colWidths {
		blockStart[c] = off
		off += cw + tagColGap
	}
	cursorCol := s.tagDispPos(s.cursor) / rows

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

	scrollCol := s.scrollCol
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
	s.scrollCol = scrollCol

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

func (s *State) renderTagPopup() string {
	h, _ := s.tagPopupSize()
	title := "Tags & Sources"
	g := s.glyphs()
	style := s.styles.status
	v := style.Render(g.V)
	pad := " "

	// The title is inlined into the top border and the position indicator into
	// the bottom border, so the grid fills the interior between the borders
	// (one space of left/right padding only). Section subtitles ("Sources" and
	// "Tags") are non-selectable cells laid out in the same column-major grid.
	// The box width comes from the grid so it can shrink to avoid a trailing
	// gap once the columns fill the interior.
	rows, _, colWidths, boxW := s.tagGrid()
	textW := max(1, boxW-4)
	items := s.tagItems()
	d := len(items)
	gridLines := min(rows, max(0, h-2))
	var body []string
	var renderWidths []int
	if d > 0 {
		visible := s.tagPopupWindow(colWidths, rows, textW)
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
					rowCells = append(rowCells, s.tagSubtitleCell(it, renderWidths[c]))
				} else {
					rowCells = append(rowCells, s.tagCell(it.tagIdx, renderWidths[c]))
				}
			}
			body = append(body, strings.Join(rowCells, strings.Repeat(" ", tagColGap)))
		}
	}
	for len(body) < gridLines {
		body = append(body, "")
	}

	out := []string{inlineTitleBorder(title, boxW, g, s.palette)}
	for _, l := range body {
		out = append(out, v+pad+render.PadRight(render.Truncate(l, textW), textW)+pad+v)
	}
	out = append(out, s.tagBottomBorder(boxW, g))
	return strings.Join(out, "\n")
}

// tagBottomBorder renders the tag popup's bottom edge, right-aligning the
// position indicator `<percent>% • <nth>/<total>` — the selected tag's place
// among all tags — mirroring the article view's bottom border and the list
// view's status bar, with dash fill to the left. The line and the indicator
// render in the same chrome role as the rest of the popup border.
func (s *State) tagBottomBorder(w int, g render.BorderGlyphs) string {
	style := s.styles.status
	textStyle := s.styles.status

	n := len(s.tags)
	pos := s.cursor + 1
	if pos < 0 || n == 0 {
		pos = 0
	}
	if pos > n {
		pos = n
	}
	pct := 0
	if n > 0 {
		pct = (pos*100 + n/2) / n
	}
	percentStr := fmt.Sprintf("%d%%", pct)
	ratioStr := fmt.Sprintf("%d/%d", pos, n)
	iw := ansi.StringWidth(percentStr) + 1 + ansi.StringWidth(g.Bullet) + 1 + ansi.StringWidth(ratioStr)
	fill := max(0, w-6-iw)
	dash := style.Render(g.H)
	return style.Render(g.BL) + dash + style.Render(strings.Repeat(g.H, fill)) + " " +
		textStyle.Render(percentStr) + " " + style.Render(g.Bullet) + " " + textStyle.Render(ratioStr) +
		" " + dash + style.Render(g.BR)
}

// renderLinksPopup renders the article links popup: each row shows the link
// text followed by the URL in the dim style, truncated to the box width.
func (s *State) renderLinksPopup() string {
	h := s.height * 6 / 10
	w := s.width * 8 / 10
	if h < 3 {
		h = 3
	}
	if w < 12 {
		w = 12
	}

	var lines []string
	lines = append(lines, s.styles.statusBold.Render("Links"))
	dim := s.styles.dim
	for i, l := range s.links {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}
		// text + separator + dim URL, truncated to the interior width.
		innerW := w - 6
		url := ansi.Truncate(l.URL, max(1, innerW-ansi.StringWidth(l.Text)-3), s.glyphs().Ellipsis)
		line := l.Text + dim.Render(" · "+url)
		lines = append(lines, cursor+ansi.Truncate(line, innerW, ""))
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.palette.StatusBar).
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

func (s *State) glyphs() render.BorderGlyphs {
	g := render.GlyphsFor(s.ascii)
	if s.sess.Config().Display.Scrollbar == "double" && !s.ascii {
		g.Fill = "║"
	}
	return g
}

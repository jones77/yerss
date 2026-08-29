package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"yerss/internal/config"
	"yerss/internal/store"
)

type listState struct {
	groups []dayGroup
	total  int
	cursor int
	filter string
}

type articleItem struct {
	ID          int64
	Title       string
	Read        bool
	PublishedAt time.Time
}

// dayGroup is a collapsible bucket of articles sharing one local calendar day.
// date is local midnight of that day; label is the rendered header text.
type dayGroup struct {
	date      time.Time
	label     string
	collapsed bool
	articles  []articleItem
}

// rowKind distinguishes a day-header row from an article row in the flattened
// visible-row list.
type rowKind int

const (
	rowHeader rowKind = iota
	rowArticle
)

// visibleRow identifies a rendered row in the flattened visible-row list.
type visibleRow struct {
	kind     rowKind
	groupIdx int
	artIdx   int // valid when kind is rowArticle
}

func (m *Model) loadList() {
	arts, err := m.store.ListArticles(m.list.filter)
	if err != nil {
		m.setStatus("load error: " + err.Error())
		return
	}
	m.list.groups = bucketDayGroups(arts)
	if total, err := m.store.ArticleCount(); err == nil {
		m.list.total = total
	}
	if size, err := m.store.DBSize(); err == nil {
		m.dbSize = size
	}
	m.clampCursor()
}

// bucketDayGroups buckets a reverse-chronological article list into day groups,
// most recent day first. Articles with no publication time land in a trailing
// "Undated" group.
func bucketDayGroups(arts []store.Article) []dayGroup {
	var groups []dayGroup
	idx := map[string]int{}
	for _, a := range arts {
		title := a.Title
		if title == "" {
			title = "(untitled)"
		}
		item := articleItem{ID: a.ID, Title: title, Read: a.Read, PublishedAt: a.PublishedAt}

		key := "undated"
		t := a.PublishedAt.Local()
		if !a.PublishedAt.IsZero() {
			day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			key = day.Format("20060102")
			t = day
		}

		i, ok := idx[key]
		if !ok {
			label := "Undated"
			if key != "undated" {
				label = longDate(t)
			}
			i = len(groups)
			groups = append(groups, dayGroup{date: t, label: label})
			idx[key] = i
		}
		groups[i].articles = append(groups[i].articles, item)
	}
	return groups
}

// visibleRows flattens the day groups into the ordered rows that are rendered
// and navigated: each group emits its header, then its articles unless the
// group is collapsed.
func (m *Model) visibleRows() []visibleRow {
	var rows []visibleRow
	for gi := range m.list.groups {
		rows = append(rows, visibleRow{kind: rowHeader, groupIdx: gi})
		if !m.list.groups[gi].collapsed {
			for ai := range m.list.groups[gi].articles {
				rows = append(rows, visibleRow{kind: rowArticle, groupIdx: gi, artIdx: ai})
			}
		}
	}
	return rows
}

// articleCount returns the number of articles currently loaded across all day
// groups (the filtered count for the status bar).
func (m *Model) articleCount() int {
	n := 0
	for i := range m.list.groups {
		n += len(m.list.groups[i].articles)
	}
	return n
}

// clampCursor keeps the cursor within the flattened visible-row list.
func (m *Model) clampCursor() {
	n := len(m.visibleRows())
	if n == 0 {
		m.list.cursor = 0
	} else if m.list.cursor >= n {
		m.list.cursor = n - 1
	} else if m.list.cursor < 0 {
		m.list.cursor = 0
	}
}

func (m *Model) collapse(i int) {
	if i >= 0 && i < len(m.list.groups) {
		m.list.groups[i].collapsed = true
	}
	m.clampCursor()
}

func (m *Model) expand(i int) {
	if i >= 0 && i < len(m.list.groups) {
		m.list.groups[i].collapsed = false
	}
	m.clampCursor()
}

func (m *Model) toggle(i int) {
	if i >= 0 && i < len(m.list.groups) {
		m.list.groups[i].collapsed = !m.list.groups[i].collapsed
	}
	m.clampCursor()
}

// longDate renders a local calendar day as `Weekday Day-ordinal Month, Year`,
// for example "Saturday 28th August, 2026".
func longDate(t time.Time) string {
	d := t.Day()
	suf := "th"
	switch d % 10 {
	case 1:
		suf = "st"
	case 2:
		suf = "nd"
	case 3:
		suf = "rd"
	}
	if d/10 == 1 {
		suf = "th"
	}
	return t.Format("Monday ") + fmt.Sprintf("%d%s ", d, suf) + t.Format("January, 2006")
}

func (m *Model) renderList() string {
	var b strings.Builder

	if m.articleCount() == 0 {
		dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
		b.WriteString(dim.Render("No articles."))
		b.WriteString("\n")
		if m.list.filter != "" {
			b.WriteString(dim.Render("Filter \"" + m.list.filter + "\" has no articles. Press Esc to clear."))
		} else {
			b.WriteString(dim.Render("Add feeds to " + m.cfg.FeedsFile() + " and press R to refresh."))
		}
	} else {
		rows := m.visibleRows()
		visible := m.height - 1
		if visible < 1 {
			visible = 1
		}
		start, end := listWindow(len(rows), m.list.cursor, visible)
		for i := start; i < end; i++ {
			row := rows[i]
			if row.kind == rowHeader {
				b.WriteString(m.renderDayHeader(&m.list.groups[row.groupIdx], i == m.list.cursor))
			} else {
				g := &m.list.groups[row.groupIdx]
				b.WriteString(m.renderArticleRow(&g.articles[row.artIdx], i == m.list.cursor))
			}
			b.WriteString("\n")
		}
	}

	status := m.renderStatusBar()
	content := b.String()
	lines := strings.Split(content, "\n")
	h := m.height
	if h < 1 {
		h = 1
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	if len(lines) > 0 {
		lines[len(lines)-1] = status
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderDayHeader(g *dayGroup, selected bool) string {
	fold := glyphsFor(m.ascii).expand
	if g.collapsed {
		fold = glyphsFor(m.ascii).collapse
	}
	style := lipgloss.NewStyle().Foreground(m.palette.Dim)
	if selected {
		style = style.Background(lipgloss.Color("#333333"))
	}
	return style.Render(fold + " " + g.label)
}

func (m *Model) renderArticleRow(item *articleItem, selected bool) string {
	const (
		timeW = 8
		gap   = 1
	)
	cursor := "  "
	if selected {
		cursor = "> "
	}
	titleW := m.width - 2 - timeW - gap
	if titleW < 1 {
		titleW = 1
	}
	title := truncate(item.Title, titleW)

	ts := "--:--:--"
	if !item.PublishedAt.IsZero() {
		ts = item.PublishedAt.Local().Format("15:04:05")
	}

	pad := titleW - runewidth.StringWidth(title)
	if pad < 0 {
		pad = 0
	}
	line := cursor + title + strings.Repeat(" ", pad+gap) + ts
	line = truncate(line, m.width)

	var style lipgloss.Style
	if item.Read {
		style = lipgloss.NewStyle().Foreground(m.palette.Dim)
	} else {
		style = lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bold)
	}
	if selected {
		style = style.Background(lipgloss.Color("#333333"))
	}
	return style.Render(line)
}

func (m *Model) renderStatusBar() string {
	left := ""
	if m.statusMsg != "" && time.Now().Before(m.statusExpires) {
		left = m.statusMsg
	} else if m.refreshing {
		left = "refreshing..."
	} else {
		left = fmt.Sprintf("%d/%d articles", m.articleCount(), m.list.total)
	}
	if m.list.filter != "" && (m.statusMsg == "" || !time.Now().Before(m.statusExpires)) {
		left += " · filter " + m.list.filter
	}

	right := ""
	if m.lastRefreshedAt.IsZero() {
		right = "never refreshed"
	} else {
		right = "last refresh " + m.lastRefreshedAt.Format("2006-01-02 15:04")
	}
	if m.dbSize > 0 {
		right = formatSize(m.dbSize) + " · " + right
	}

	pad := m.width - runewidth.StringWidth(left) - runewidth.StringWidth(right)
	if pad < 1 {
		pad = 1
	}
	line := left + strings.Repeat(" ", pad) + right
	line = truncate(line, m.width)
	return lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render(line)
}

func (m *Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.visibleRows()
	if m.list.cursor >= 0 && m.list.cursor < len(rows) {
		row := rows[m.list.cursor]
		if row.kind == rowHeader {
			switch msg.String() {
			case "h":
				m.collapse(row.groupIdx)
				return m, nil
			case "l":
				m.expand(row.groupIdx)
				return m, nil
			case "enter", " ", "tab":
				m.toggle(row.groupIdx)
				return m, nil
			}
		} else if msg.String() == "tab" {
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
		return m, tea.Quit
	case config.MoveDown:
		m.moveListCursor(1)
	case config.MoveUp:
		m.moveListCursor(-1)
	case config.OpenArticle:
		m.openArticle()
	case config.Back:
		if m.list.filter != "" {
			m.clearFilter()
		}
	case config.Refresh:
		cmd = m.refreshManual()
	case config.PageDown:
		m.moveListCursor(m.pageSize())
	case config.PageUp:
		m.moveListCursor(-m.pageSize())
	case config.HalfPageDown:
		m.moveListCursor(m.pageSize() / 2)
	case config.HalfPageUp:
		m.moveListCursor(-m.pageSize() / 2)
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
	case config.ToggleRead:
		m.toggleReadAtCursor()
	case config.MarkAllRead:
		m.markAllVisibleRead()
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

// updateListMouse handles mouse events in the list view: wheel moves the
// cursor by one row, a left-click on an article row selects and opens it, and
// a left-click on a day-header row toggles that group's fold. Clicks on the
// status bar (the bottom line) are ignored.
func (m *Model) updateListMouse(msg tea.MouseMsg) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		m.moveListCursor(-1)
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		m.moveListCursor(1)
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		m.handleListClick(msg.Y)
	}
}

// handleListClick maps a screen row to a visible-row index via the list window
// start offset and dispatches on the row kind found there.
func (m *Model) handleListClick(y int) {
	if y >= m.height-1 {
		return
	}
	rows := m.visibleRows()
	if len(rows) == 0 {
		return
	}
	visible := m.height - 1
	if visible < 1 {
		visible = 1
	}
	start, _ := listWindow(len(rows), m.list.cursor, visible)
	idx := start + y
	if idx < 0 || idx >= len(rows) {
		return
	}
	row := rows[idx]
	if row.kind == rowHeader {
		m.toggle(row.groupIdx)
		return
	}
	m.list.cursor = idx
	m.openArticle()
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
	m.list.cursor = (m.list.cursor + delta) % n
	if m.list.cursor < 0 {
		m.list.cursor += n
	}
}

func (m *Model) toggleReadAtCursor() {
	rows := m.visibleRows()
	if m.list.cursor < 0 || m.list.cursor >= len(rows) {
		return
	}
	row := rows[m.list.cursor]
	if row.kind != rowArticle {
		return
	}
	item := &m.list.groups[row.groupIdx].articles[row.artIdx]
	item.Read = !item.Read
	_ = m.store.SetRead(item.ID, item.Read)
}

func (m *Model) markAllVisibleRead() {
	var ids []int64
	for gi := range m.list.groups {
		for ai := range m.list.groups[gi].articles {
			item := &m.list.groups[gi].articles[ai]
			ids = append(ids, item.ID)
			item.Read = true
		}
	}
	if len(ids) > 0 {
		_ = m.store.MarkAllRead(ids)
	}
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
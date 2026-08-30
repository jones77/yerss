package ui

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"golang.org/x/net/publicsuffix"

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
	Link        string
	FeedURL     string
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
	m.list.groups = bucketDayGroups(arts, time.Now())
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
// "Undated" group. now is the reference time for the today/yesterday label
// prefixes; callers pass time.Now() and tests pass a fixed time.
func bucketDayGroups(arts []store.Article, now time.Time) []dayGroup {
	var groups []dayGroup
	idx := map[string]int{}
	for _, a := range arts {
		title := a.Title
		if title == "" {
			title = "(untitled)"
		}
		item := articleItem{ID: a.ID, Title: title, Read: a.Read, PublishedAt: a.PublishedAt, Link: a.Link, FeedURL: a.FeedURL}

		key := dayKey(a.PublishedAt)
		i, ok := idx[key]
		if !ok {
			label := "Undated"
			var date time.Time
			if key != "undated" {
				date = dayStart(a.PublishedAt)
				label = dayLabel(date, now)
			}
			i = len(groups)
			groups = append(groups, dayGroup{date: date, label: label})
			idx[key] = i
		}
		groups[i].articles = append(groups[i].articles, item)
	}
	return groups
}

// dayLabel renders a day group's header label: the long local date, prefixed
// with "today, " or "yesterday, " when the day is the reference day or the one
// before it.
func dayLabel(date, now time.Time) string {
	today := dayStart(now)
	switch date {
	case today:
		return "today, " + longDate(date)
	case today.AddDate(0, 0, -1):
		return "yesterday, " + longDate(date)
	}
	return longDate(date)
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

// selectedArticlePosition returns the 1-based position of the article under the
// cursor among the loaded article rows (day headers are not counted) and the
// total number of articles in the store. A cursor on a day header reports the
// position of the article just above it (0 when no articles precede it).
func (m *Model) selectedArticlePosition() (n, total int) {
	rows := m.visibleRows()
	for i := 0; i <= m.list.cursor && i < len(rows); i++ {
		if rows[i].kind == rowArticle {
			n++
		}
	}
	return n, m.list.total
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
		start, end := listWindow(len(rows), m.list.cursor, m.pageSize())
		for i := start; i < end; i++ {
			row := rows[i]
			if row.kind == rowHeader {
				corner := m.railGlyph(len(m.list.groups), row.groupIdx)
				b.WriteString(m.renderDayHeader(&m.list.groups[row.groupIdx], corner, i == m.list.cursor))
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

// railGlyph returns the tree-rail corner for day group i of n: `┌` on the
// first day group, `└` on the last, `├` on every interior group. Corners are
// computed against the full group list, not the scroll window, so they do not
// move as the window scrolls. Only day headers render a rail; article rows
// carry no tree glyph.
func (m *Model) railGlyph(n, i int) string {
	g := glyphsFor(m.ascii)
	if i == 0 {
		return g.tl
	}
	if i == n-1 {
		return g.bl
	}
	return g.tee
}

func (m *Model) renderDayHeader(g *dayGroup, corner string, selected bool) string {
	style := lipgloss.NewStyle().Foreground(m.palette.Dim)
	if selected {
		style = style.Background(lipgloss.Color("#333333"))
	}
	return style.Render(corner + " " + g.label)
}

func (m *Model) renderArticleRow(item *articleItem, selected bool) string {
	g := glyphsFor(m.ascii)
	bg := lipgloss.Color("#333333")
	var titleStyle lipgloss.Style
	if item.Read {
		titleStyle = lipgloss.NewStyle().Foreground(m.palette.Dim)
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bright)
	}
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
	bar := lipgloss.NewStyle()
	if selected {
		titleStyle = titleStyle.Background(bg)
		dim = dim.Background(bg)
		bar = bar.Background(bg)
	}

	src := sourceID(item.Link, item.FeedURL)
	ts := "--:--"
	if !item.PublishedAt.IsZero() {
		ts = item.PublishedAt.Local().Format("15:04")
	}

	// The row starts with the local publication time, the title filling the
	// middle, and the source identifier right-aligned as the row's only
	// right-hand field, with at least one space between the two even when the
	// title is truncated (it then ends with an ellipsis). Article rows carry
	// no tree glyph; only day headers branch.
	railW := ansi.StringWidth(ts) + 1
	titleW := m.width - railW
	if src != "" {
		titleW -= ansi.StringWidth(src) + 1
	}
	if titleW < 1 {
		titleW = 1
	}
	title := item.Title
	if w := ansi.StringWidth(title); w > titleW {
		if titleW < ansi.StringWidth(g.ellipsis) {
			title = truncate(title, titleW)
		} else {
			title = ansi.Truncate(title, titleW, g.ellipsis)
		}
	}
	pad := titleW - ansi.StringWidth(title)
	if pad < 0 {
		pad = 0
	}

	var b strings.Builder
	b.WriteString(dim.Render(ts))
	b.WriteString(bar.Render(" "))
	b.WriteString(titleStyle.Render(title))
	b.WriteString(bar.Render(strings.Repeat(" ", pad)))
	if src != "" {
		b.WriteString(bar.Render(" "))
		b.WriteString(dim.Render(src))
	}
	return truncate(b.String(), m.width)
}

// sourceID derives a short publication identifier from a URL host: strip any
// trailing dot and a leading "www.", take the organization label of the
// registrable domain (the public suffix plus one, so multi-part suffixes like
// "co.uk" are removed whole), and truncate to at most 12 characters. The
// article link host is used when present; otherwise the feed URL host is
// used. Hosts with no registrable domain (single-label hosts, IP literals)
// fall back to dropping everything from the final dot.
func sourceID(rawURL, feedURL string) string {
	host := ""
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		host = u.Host
	} else if u, err := url.Parse(feedURL); err == nil && u.Host != "" {
		host = u.Host
	}
	if host == "" {
		return ""
	}
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	h = strings.TrimPrefix(h, "www.")
	if etld1, err := publicsuffix.EffectiveTLDPlusOne(h); err == nil {
		if i := strings.Index(etld1, "."); i > 0 {
			h = etld1[:i]
		} else {
			h = etld1
		}
	} else if i := strings.LastIndex(h, "."); i > 0 {
		h = h[:i]
	}
	if len(h) > 12 {
		h = h[:12]
	}
	return h
}

func (m *Model) renderStatusBar() string {
	left := ""
	if m.statusMsg != "" && time.Now().Before(m.statusExpires) {
		left = m.statusMsg
	} else if m.refreshing {
		left = "refreshing..."
	} else {
		n, total := m.selectedArticlePosition()
		pct := 0
		if total > 0 {
			pct = int(float64(n)/float64(total)*100 + 0.5)
		}
		left = fmt.Sprintf("%d%% · %d/%d", pct, n, total)
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
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/timeutil"
	"yerss/internal/ui/render"
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

// Fold keys toggled while the cursor rests on a day-header row. h/l fold and
// unfold the group; enter/space/tab toggle it. These are context-sensitive
// (only active on a header row), so they are centralized here rather than
// registered in the keybinding catalog.
const (
	keyFoldIn          = "h"
	keyFoldOut         = "l"
	keyFoldToggle      = "enter"
	keyFoldToggleSpace = " "
	keyFoldToggleTab   = "tab"
)

// visibleRow identifies a rendered row in the flattened visible-row list.
type visibleRow struct {
	kind     rowKind
	groupIdx int
	artIdx   int // valid when kind is rowArticle
}

func (m *Model) loadList() {
	arts, err := m.sess.ListArticles(m.list.filter)
	if err != nil {
		m.setStatus("load error: " + err.Error())
		return
	}
	m.list.groups = bucketDayGroups(arts, time.Now())
	if total, err := m.sess.ArticleCount(); err == nil {
		m.list.total = total
	}
	if size, err := m.sess.DBSize(); err == nil {
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

		key := timeutil.DayKey(a.PublishedAt)
		i, ok := idx[key]
		if !ok {
			label := "Undated"
			var date time.Time
			if key != "undated" {
				date = timeutil.DayStart(a.PublishedAt)
				label = timeutil.DayLabel(date, now)
			}
			i = len(groups)
			groups = append(groups, dayGroup{date: date, label: label})
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

// expandToggle expands every day group when any is collapsed and collapses
// them all when every group is already expanded, so a mixed state expands.
// The new state is applied uniformly and the cursor is clamped because the
// visible-row list may shrink or grow.
func (m *Model) expandToggle() {
	anyCollapsed := false
	for i := range m.list.groups {
		if m.list.groups[i].collapsed {
			anyCollapsed = true
			break
		}
	}
	for i := range m.list.groups {
		m.list.groups[i].collapsed = !anyCollapsed
	}
	m.clampCursor()
}

func (m *Model) renderList() string {
	var b strings.Builder

	if m.articleCount() == 0 {
		dim := m.styles.dim
		b.WriteString(dim.Render("No articles."))
		b.WriteString("\n")
		if m.list.filter != "" {
			b.WriteString(dim.Render("Filter \"" + m.list.filter + "\" has no articles. Press Esc to clear."))
		} else {
			b.WriteString(dim.Render("Add feeds to " + m.sess.Config().FeedsFile() + " and press R to refresh."))
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
	g := m.glyphs()
	if i == 0 {
		return g.TL
	}
	if i == n-1 {
		return g.BL
	}
	return g.Tee
}

func (m *Model) renderDayHeader(g *dayGroup, corner string, selected bool) string {
	style := m.styles.status
	if selected {
		style = m.styles.selStatus
	}
	connector := strings.Repeat(m.glyphs().H, 4)
	return style.Render(corner + connector + " " + g.label)
}

func (m *Model) renderArticleRow(item *articleItem, selected bool) string {
	g := m.glyphs()
	titleStyle := m.styles.text
	if !item.Read {
		titleStyle = m.styles.bright
	}
	text := m.styles.text
	bar := m.styles.plain
	if selected {
		if !item.Read {
			titleStyle = m.styles.selBright
		} else {
			titleStyle = m.styles.selText
		}
		text = m.styles.selText
		bar = m.styles.selRow
	}

	src := store.SourceLabel(item.Link, item.FeedURL)
	if src != "" {
		src = render.ElideMiddle(src, 15, g.Ellipsis)
	}
	ts := "--:--"
	if !item.PublishedAt.IsZero() {
		ts = item.PublishedAt.Local().Format(timeutil.LayoutTime)
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
		if titleW < ansi.StringWidth(g.Ellipsis) {
			title = render.Truncate(title, titleW)
		} else {
			title = ansi.Truncate(title, titleW, g.Ellipsis)
		}
	}
	pad := titleW - ansi.StringWidth(title)
	if pad < 0 {
		pad = 0
	}

	var b strings.Builder
	b.WriteString(text.Render(ts))
	b.WriteString(bar.Render(" "))
	b.WriteString(titleStyle.Render(title))
	b.WriteString(bar.Render(strings.Repeat(" ", pad)))
	if src != "" {
		b.WriteString(bar.Render(" "))
		b.WriteString(text.Render(src))
	}
	return render.Truncate(b.String(), m.width)
}

// sourceID derives a short publication identifier from a URL host: the
func (m *Model) renderStatusBar() string {
	base := m.styles.status
	dim := m.styles.dim
	bullet := m.glyphs().Bullet

	left := ""
	if m.statusMsg != "" && time.Now().Before(m.statusExpires) {
		left = base.Render(m.statusMsg)
	} else if m.refreshing {
		left = base.Render("refreshing...")
	} else if m.lastRefreshedAt.IsZero() {
		left = base.Render("never refreshed")
	} else {
		t := m.lastRefreshedAt
		left = base.Render(t.Format(timeutil.LayoutTime)+" "+timeutil.LongDate(t)) + " " + dim.Render("last refresh")
	}
	if m.list.filter != "" && (m.statusMsg == "" || !time.Now().Before(m.statusExpires)) {
		left += " " + dim.Render(bullet) + " " + base.Render("filter "+m.list.filter)
	}

	n, total := m.selectedArticlePosition()
	pct := 0
	if total > 0 {
		pct = (n*100 + total/2) / total
	}
	// The right side runs `?: help · DB <n>MB · <percent>% ·
	// <n>/<total>`: the help hint first, then the on-disk database size
	// (DB) as a whole number of megabytes, then the scroll percentage and
	// position. Each element is prepended so the first right-aligned
	// element is `?: help`.
	right := base.Render(fmt.Sprintf("%d/%d", n, total))
	right = base.Render(fmt.Sprintf("%d%%", pct)) + " " + dim.Render(bullet) + " " + right
	right = base.Render("DB " + formatMB(m.dbSize)) + " " + dim.Render(bullet) + " " + right
	right = base.Render("?: help") + " " + dim.Render(bullet) + " " + right

	pad := m.width - ansi.StringWidth(left) - ansi.StringWidth(right)
	if pad < 1 {
		pad = 1
	}
	line := left + strings.Repeat(" ", pad) + right
	line = render.Truncate(line, m.width)
	return base.Render(line)
}

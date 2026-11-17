package listview

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/app"
	"yerss/internal"
	"yerss/internal/ui/render"
)

// Status carries the transient status-bar inputs owned by the Model: the
// session's latest status message (with its expiry), the in-flight refresh
// flag, and the last refresh time. The list view renders them in its status
// bar; the Model writes them via SetStatus and direct field assignment.
type Status struct {
	Msg             string
	Expires         time.Time
	Refreshing      bool
	LastRefreshedAt time.Time
}

// State is the list view: the day-grouped article list, the cursor, the
// active filter, and the rendering and input handling for that state.
type State struct {
	sess   *app.Session
	styles styles
	ascii  bool
	width  int
	height int

	groups []dayGroup
	total  int
	cursor int
	filter string
	dbSize int64

	Status Status
}

// New builds a list view bound to an application session.
func New(sess *app.Session, palette render.Palette, ascii bool, width, height int) *State {
	return &State{
		sess:   sess,
		styles: newStyles(palette),
		ascii:  ascii,
		width:  width,
		height: height,
	}
}

// styles holds the list view's palette-derived lipgloss styles.
type styles struct {
	text      lipgloss.Style
	bright    lipgloss.Style
	dim       lipgloss.Style
	status    lipgloss.Style
	plain     lipgloss.Style
	selText   lipgloss.Style
	selBright lipgloss.Style
	selStatus lipgloss.Style
	selRow    lipgloss.Style
}

func newStyles(p render.Palette) styles {
	sel := lipgloss.Color(string(p.Selection))
	return styles{
		text:      lipgloss.NewStyle().Foreground(p.Text),
		bright:    lipgloss.NewStyle().Bold(true).Foreground(p.Bright),
		dim:       lipgloss.NewStyle().Foreground(p.Dim),
		status:    lipgloss.NewStyle().Foreground(p.StatusBar),
		plain:     lipgloss.NewStyle(),
		selText:   lipgloss.NewStyle().Foreground(p.Text).Background(sel),
		selBright: lipgloss.NewStyle().Bold(true).Foreground(p.Bright).Background(sel),
		selStatus: lipgloss.NewStyle().Foreground(p.StatusBar).Background(sel),
		selRow:    lipgloss.NewStyle().Background(sel),
	}
}

// SetSize updates the view's geometry for rendering and mouse hit-testing.
func (s *State) SetSize(width, height int) {
	s.width, s.height = width, height
}

// SetStatus records a transient status message with the standard 3s expiry.
func (s *State) SetStatus(msg string) {
	s.Status.Msg = msg
	s.Status.Expires = time.Now().Add(3 * time.Second)
}

// Filter returns the active tag filter.
func (s *State) Filter() string { return s.filter }

// SetFilter sets the active tag filter without reloading the list.
func (s *State) SetFilter(filter string) { s.filter = filter }

// Load reloads the list from the store under the current filter. It returns a
// status message describing a load failure, or "" on success.
func (s *State) Load() string {
	arts, err := s.sess.ListArticlesLite(s.filter)
	if err != nil {
		return "load error: " + err.Error()
	}
	s.groups = bucketDayGroups(arts, time.Now())
	if total, err := s.sess.ArticleCount(); err == nil {
		s.total = total
	}
	if size, err := s.sess.DBSize(); err == nil {
		s.dbSize = size
	}
	s.clampCursor()
	return ""
}

// ClearFilter clears the active filter and reloads the list. It returns a
// status message describing a load failure, or "" on success.
func (s *State) ClearFilter() string {
	s.filter = ""
	return s.Load()
}

// Item is a list row's article summary.
type Item struct {
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
	articles  []Item
}

// rowKind distinguishes a day-header row from an article row in the flattened
// visible-row list.
type rowKind int

const (
	rowHeader rowKind = iota
	rowArticle
)

// Fold keys toggled while the cursor rests on a day-header row. h/l fold and
// unfold the group; enter/tab toggle it. These are context-sensitive
// (only active on a header row), so they are centralized here rather than
// registered in the keybinding catalog.
const (
	keyFoldIn        = "h"
	keyFoldOut       = "l"
	keyFoldToggle    = "enter"
	keyFoldToggleTab = "tab"
)

// visibleRow identifies a rendered row in the flattened visible-row list.
type visibleRow struct {
	kind     rowKind
	groupIdx int
	artIdx   int // valid when kind is rowArticle
}

// bucketDayGroups buckets a reverse-chronological article list into day groups,
// most recent day first. Articles with no publication time land in a trailing
// "Undated" group. now is the reference time for the today/yesterday label
// prefixes; callers pass time.Now() and tests pass a fixed time.
func bucketDayGroups(arts []internal.Article, now time.Time) []dayGroup {
	var groups []dayGroup
	idx := map[string]int{}
	for _, a := range arts {
		title := a.Title
		if title == "" {
			title = "(untitled)"
		}
		item := Item{ID: a.ID, Title: title, Read: a.Read, PublishedAt: a.PublishedAt, Link: a.Link, FeedURL: a.FeedURL}

		key := internal.DayKey(a.PublishedAt)
		i, ok := idx[key]
		if !ok {
			label := "Undated"
			var date time.Time
			if key != "undated" {
				date = internal.DayStart(a.PublishedAt)
				label = internal.DayLabel(date, now)
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
func (s *State) visibleRows() []visibleRow {
	var rows []visibleRow
	for gi := range s.groups {
		rows = append(rows, visibleRow{kind: rowHeader, groupIdx: gi})
		if !s.groups[gi].collapsed {
			for ai := range s.groups[gi].articles {
				rows = append(rows, visibleRow{kind: rowArticle, groupIdx: gi, artIdx: ai})
			}
		}
	}
	return rows
}

// articleCount returns the number of articles currently loaded across all day
// groups (the filtered count for the status bar).
func (s *State) articleCount() int {
	n := 0
	for i := range s.groups {
		n += len(s.groups[i].articles)
	}
	return n
}

// selectedArticlePosition returns the 1-based position of the article under the
// cursor among the loaded article rows (day headers are not counted) and the
// total number of articles in the store. A cursor on a day header reports the
// position of the article just above it (0 when no articles precede it).
func (s *State) selectedArticlePosition() (n, total int) {
	rows := s.visibleRows()
	for i := 0; i <= s.cursor && i < len(rows); i++ {
		if rows[i].kind == rowArticle {
			n++
		}
	}
	return n, s.total
}

// clampCursor keeps the cursor within the flattened visible-row list.
func (s *State) clampCursor() {
	n := len(s.visibleRows())
	if n == 0 {
		s.cursor = 0
	} else if s.cursor >= n {
		s.cursor = n - 1
	} else if s.cursor < 0 {
		s.cursor = 0
	}
}

func (s *State) collapse(i int) {
	if i >= 0 && i < len(s.groups) {
		s.groups[i].collapsed = true
	}
	s.clampCursor()
}

func (s *State) expand(i int) {
	if i >= 0 && i < len(s.groups) {
		s.groups[i].collapsed = false
	}
	s.clampCursor()
}

func (s *State) toggle(i int) {
	if i >= 0 && i < len(s.groups) {
		s.groups[i].collapsed = !s.groups[i].collapsed
	}
	s.clampCursor()
}

// expandToggle expands every day group when any is collapsed and collapses
// them all when every group is already expanded, so a mixed state expands.
// The new state is applied uniformly and the cursor is clamped because the
// visible-row list may shrink or grow.
func (s *State) expandToggle() {
	anyCollapsed := false
	for i := range s.groups {
		if s.groups[i].collapsed {
			anyCollapsed = true
			break
		}
	}
	for i := range s.groups {
		s.groups[i].collapsed = !anyCollapsed
	}
	s.clampCursor()
}

// glyphs returns the glyph set for the current mode, applying the configured
// scrollbar thumb style: single line by default, or a double line when the
// scrollbar option is set to "double" (Unicode only; ASCII fallback has no
// double-line glyph).
func (s *State) glyphs() render.BorderGlyphs {
	g := render.GlyphsFor(s.ascii)
	if s.sess.Config().Display.Scrollbar == "double" && !s.ascii {
		g.Fill = "║"
	}
	return g
}

// Render returns the full list frame for the current state.
func (s *State) Render() string {
	var b strings.Builder

	if s.articleCount() == 0 {
		dim := s.styles.dim
		b.WriteString(dim.Render("No articles."))
		b.WriteString("\n")
		if s.filter != "" {
			b.WriteString(dim.Render("Filter \"" + s.filter + "\" has no articles. Press Esc to clear."))
		} else {
			b.WriteString(dim.Render("Add feeds to " + s.sess.Config().FeedsFile() + " and press R to refresh."))
		}
	} else {
		rows := s.visibleRows()
		start, end := listWindow(len(rows), s.cursor, s.pageSize())
		for i := start; i < end; i++ {
			row := rows[i]
			if row.kind == rowHeader {
				corner := s.railGlyph(len(s.groups), row.groupIdx)
				b.WriteString(s.renderDayHeader(&s.groups[row.groupIdx], corner, i == s.cursor))
			} else {
				g := &s.groups[row.groupIdx]
				b.WriteString(s.renderArticleRow(&g.articles[row.artIdx], i == s.cursor))
			}
			b.WriteString("\n")
		}
	}

	status := s.renderStatusBar()
	content := b.String()
	lines := strings.Split(content, "\n")
	h := s.height
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

// railGlyph returns the tree-rail corner for day group i of n, delegating to
// the pure render helper.
func (s *State) railGlyph(n, i int) string {
	return render.RailGlyph(s.glyphs(), s.groups[i].collapsed, n, i)
}

func (s *State) renderDayHeader(g *dayGroup, corner string, selected bool) string {
	style := s.styles.status
	if selected {
		style = s.styles.selStatus
	}
	return render.DayHeader(style, s.glyphs(), corner, g.label)
}

func (s *State) renderArticleRow(item *Item, selected bool) string {
	g := s.glyphs()
	titleStyle := s.styles.text
	if !item.Read {
		titleStyle = s.styles.bright
	}
	text := s.styles.text
	bar := s.styles.plain
	if selected {
		if !item.Read {
			titleStyle = s.styles.selBright
		} else {
			titleStyle = s.styles.selText
		}
		text = s.styles.selText
		bar = s.styles.selRow
	}

	src := internal.SourceLabel(item.Link, item.FeedURL)
	if src != "" {
		src = render.ElideMiddle(src, 15, g.Ellipsis)
	}
	ts := "--:--"
	if !item.PublishedAt.IsZero() {
		ts = item.PublishedAt.Local().Format(internal.TimeLayout(s.sess.Config().Display.ShowSeconds))
	}

	// The row starts with the local publication time, the title filling the
	// middle, and the source identifier right-aligned as the row's only
	// right-hand field, with at least one space between the two even when the
	// title is truncated (it then ends with an ellipsis). Article rows carry
	// no tree glyph; only day headers branch.
	railW := ansi.StringWidth(ts) + 1
	titleW := s.width - railW
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
	return render.Truncate(b.String(), s.width)
}

func (s *State) renderStatusBar() string {
	base := s.styles.status
	dim := s.styles.dim
	g := s.glyphs()
	left := render.StatusBarLeft(base, dim, g, s.Status.Msg, time.Now().Before(s.Status.Expires), s.Status.Refreshing, s.Status.LastRefreshedAt, s.sess.Config().Display.ShowSeconds, s.filter)
	n, total := s.selectedArticlePosition()
	right := render.StatusBarRight(base, dim, g.Bullet, s.dbSize, n, total)
	return render.StatusBarLine(s.width, base, left, right)
}
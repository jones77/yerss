package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/timeutil"
	"yerss/internal/ui/render"
)

func withLocalZone(t *testing.T, loc *time.Location) {
	t.Helper()
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

func insertTimedArticle(t *testing.T, st *store.Store, title string, pub time.Time) int64 {
	t.Helper()
	a := store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "guid-" + title,
		Title:       title,
		PublishedAt: pub,
		Content:     "<p>" + title + "</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle: %v", err)
	}
	return id
}

func TestLongDateOrdinals(t *testing.T) {
	withLocalZone(t, time.UTC)
	cases := []struct {
		day  int
		want string
	}{
		{1, "1st"}, {2, "2nd"}, {3, "3rd"},
		{11, "11th"}, {12, "12th"}, {13, "13th"},
		{21, "21st"}, {22, "22nd"}, {23, "23rd"},
		{30, "30th"}, {31, "31st"},
	}
	for _, c := range cases {
		tm := time.Date(2026, 8, c.day, 0, 0, 0, 0, time.UTC)
		got := timeutil.LongDate(tm)
		if !strings.Contains(got, c.want) {
			t.Errorf("longDate(%d) = %q, missing ordinal %q", c.day, got, c.want)
		}
		if !strings.Contains(got, "August, 2026") {
			t.Errorf("longDate(%d) = %q, missing month/year", c.day, got)
		}
		if !strings.Contains(got, " ") {
			t.Errorf("longDate(%d) = %q, expected a weekday prefix", c.day, got)
		}
	}
}

func TestBucketDayGroupsOrdering(t *testing.T) {
	withLocalZone(t, time.UTC)
	// Input mirrors ListArticles order: reverse chronological, undated last.
	// The reference time makes day1 "today" and day2 "yesterday".
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	arts := []store.Article{
		{Title: "day1-1", PublishedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)},
		{Title: "day2-1", PublishedAt: time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)},
		{Title: "day2-2", PublishedAt: time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)},
		{Title: "day3-1", PublishedAt: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)},
		{Title: "undated"},
	}
	groups := bucketDayGroups(arts, now)
	if len(groups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(groups))
	}
	day := func(d int) time.Time { return time.Date(2026, 8, d, 0, 0, 0, 0, time.UTC) }
	wantLabels := []string{
		"today, " + timeutil.LongDate(day(28)),
		"yesterday, " + timeutil.LongDate(day(27)),
		timeutil.LongDate(day(26)),
		"Undated",
	}
	for i, want := range wantLabels {
		if groups[i].label != want {
			t.Errorf("groups[%d].label = %q, want %q", i, groups[i].label, want)
		}
	}
	if len(groups[1].articles) != 2 || groups[1].articles[0].Title != "day2-1" || groups[1].articles[1].Title != "day2-2" {
		t.Errorf("within-day articles not reverse chronological: %+v", groups[1].articles)
	}
}

func TestVisibleRowsCollapse(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "c", time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC))
	m.loadList()

	rows := m.visibleRows()
	if len(rows) != 5 {
		t.Fatalf("expanded visible rows = %d, want 5", len(rows))
	}
	if rows[0].kind != rowHeader || rows[1].kind != rowArticle || rows[2].kind != rowArticle ||
		rows[3].kind != rowHeader || rows[4].kind != rowArticle {
		t.Errorf("unexpected expanded row kinds: %+v", rows)
	}

	m.collapse(0)
	rows = m.visibleRows()
	if len(rows) != 3 {
		t.Fatalf("collapsed visible rows = %d, want 3", len(rows))
	}
	if rows[0].kind != rowHeader || rows[1].kind != rowHeader {
		t.Errorf("collapsed group should emit only its header: %+v", rows)
	}
}

func TestCursorWrapsAcrossVisibleRows(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m.loadList()

	// rows: [header, a, b]; cursor 0.
	m.list.cursor = 0
	m.moveListCursor(-1)
	if m.list.cursor != 2 {
		t.Errorf("wrap up: cursor = %d, want 2", m.list.cursor)
	}
	m.moveListCursor(1)
	if m.list.cursor != 0 {
		t.Errorf("wrap down: cursor = %d, want 0", m.list.cursor)
	}
}

func TestPageKeysClampAtBoundaries(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m.loadList()

	// rows: [header, a, b]; pageSize is large (> 3), so one page jump should
	// land directly on the last row, and page up from there back on the first.
	m.list.cursor = 0
	m.updateList(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.list.cursor != 2 {
		t.Errorf("page down from top: cursor = %d, want 2", m.list.cursor)
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyPgUp})
	if m.list.cursor != 0 {
		t.Errorf("page up from bottom: cursor = %d, want 0", m.list.cursor)
	}

	// Half-page and page jumps must not wrap: on the first row a page up and a
	// half-page up stay put, and on the last row a page down stays put.
	m.updateList(tea.KeyMsg{Type: tea.KeyPgUp})
	if m.list.cursor != 0 {
		t.Errorf("page up at top should clamp: cursor = %d, want 0", m.list.cursor)
	}
	m.list.cursor = 2
	m.updateList(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.list.cursor != 2 {
		t.Errorf("page down at bottom should clamp: cursor = %d, want 2", m.list.cursor)
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyCtrlD})
	if m.list.cursor != 2 {
		t.Errorf("half page down at bottom should clamp: cursor = %d, want 2", m.list.cursor)
	}
	m.list.cursor = 0
	m.updateList(tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.list.cursor != 0 {
		t.Errorf("half page up at top should clamp: cursor = %d, want 0", m.list.cursor)
	}
}

func TestOpenAndFoldKeysOnHeader(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	m.loadList()

	m.list.cursor = 0 // header row
	m.updateList(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != viewList {
		t.Errorf("Enter on a header should not open an article, view = %d", m.view)
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.view != viewList {
		t.Errorf("l on a header should not open an article, view = %d", m.view)
	}
	if m.list.groups[0].collapsed {
		t.Error("enter/l on a header should not toggle the fold")
	}
}

func TestFoldKeys(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m.loadList()

	m.list.cursor = 0 // header
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if !m.list.groups[0].collapsed {
		t.Error("h on a header should collapse the day")
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.list.groups[0].collapsed {
		t.Error("l on a header should expand the day")
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.list.groups[0].collapsed {
		t.Error("enter on a header should toggle the fold")
	}
	m.updateList(tea.KeyMsg{Type: tea.KeyTab})
	if m.list.groups[0].collapsed {
		t.Error("tab on a header should toggle the fold")
	}
	m.updateList(tea.KeyMsg{Type: tea.KeySpace})
	if !m.list.groups[0].collapsed {
		t.Error("space on a header should toggle the fold")
	}

	m.expand(0)
	m.list.cursor = 1 // first article row
	m.updateList(tea.KeyMsg{Type: tea.KeyTab})
	if !m.list.groups[0].collapsed {
		t.Error("tab on an article row should toggle its day group")
	}
}

func TestStatusBarShowsNT(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", []string{"tech"})
	insertArticle(t, st, "two", nil)
	m.loadList()

	// rows: [header, one, two]; select the second article (position 2 of 2).
	m.list.cursor = 2
	if got := m.renderStatusBar(); !strings.Contains(got, "100% "+render.GlyphsFor(m.ascii).Bullet+" 2/2") {
		t.Errorf("status bar = %q, want 100%% %s 2/2", got, render.GlyphsFor(m.ascii).Bullet)
	}

	m.popupData = popupState{
		tags:   []store.TagCount{{Name: "tech", Total: 1, Unread: 1}},
		cursor: 0,
	}
	m.confirmTagSelection()
	// filtered to [header, one]; select article one (position 1 of 2 total).
	m.list.cursor = 1
	got := m.renderStatusBar()
	if !strings.Contains(got, "50% "+render.GlyphsFor(m.ascii).Bullet+" 1/2") {
		t.Errorf("filtered status bar = %q, want 50%% %s 1/2", got, render.GlyphsFor(m.ascii).Bullet)
	}
	if !strings.Contains(got, "filter tech") {
		t.Errorf("filtered status bar missing filter name: %q", got)
	}
}

func TestStatusBarShowsDBSize(t *testing.T) {
	m, st := newTestModel(t)
	m.width = 120 // the RAM/DB/percent/position right side is long
	insertArticle(t, st, "one", nil)
	m.loadList()

	got := m.renderStatusBar()
	if m.dbSize <= 0 {
		t.Fatalf("expected dbSize set by loadList, got %d", m.dbSize)
	}
	if !strings.Contains(got, "DB "+formatMB(m.dbSize)) {
		t.Errorf("status bar = %q, want DB %s", got, formatMB(m.dbSize))
	}
	if !strings.Contains(got, "never refreshed") {
		t.Errorf("status bar = %q, want size alongside refresh info", got)
	}

	m.lastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")
	m.list.cursor = 1 // the single article
	got = m.renderStatusBar()
	if !strings.Contains(got, "15:04 Friday 28th August, 2026 last refresh") {
		t.Errorf("status bar = %q, want last refresh time with long date", got)
	}
	if !strings.Contains(got, "DB "+formatMB(m.dbSize)) {
		t.Errorf("status bar = %q, want DB %s", got, formatMB(m.dbSize))
	}
	if !strings.Contains(got, "DB "+formatMB(m.dbSize)+" "+render.GlyphsFor(m.ascii).Bullet+" 100% "+render.GlyphsFor(m.ascii).Bullet+" 1/1") {
		t.Errorf("status bar = %q, want DB %s percentage %s position", got, render.GlyphsFor(m.ascii).Bullet, render.GlyphsFor(m.ascii).Bullet)
	}
}

func TestStatusBarShowsRAM(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()

	got := m.renderStatusBar()
	if !strings.Contains(got, "RAM 0MB") {
		t.Errorf("status bar with empty cache = %q, want RAM 0MB", got)
	}

	m.sess.ImgCache.Set("u", image.NewRGBA(image.Rect(0, 0, 800, 600)))
	got = m.renderStatusBar()
	want := formatMB(int64(800 * 600 * 4))
	if !strings.Contains(got, "RAM "+want) {
		t.Errorf("status bar = %q, want RAM %s", got, want)
	}
}

func TestStatusBarBulletAndRefreshDim(t *testing.T) {
	forceTrueColor(t)
	m, st := newTestModel(t)
	m.SetAscii(false)
	insertArticle(t, st, "one", nil)
	m.loadList()
	m.list.cursor = 1
	m.lastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")

	got := m.renderStatusBar()
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
	base := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	bullet := render.GlyphsFor(m.ascii).Bullet
	if !strings.Contains(got, dim.Render(bullet)) {
		t.Errorf("bullets should render in the dim role (ANSI 8): %q", got)
	}
	if !strings.Contains(got, dim.Render("last refresh")) {
		t.Errorf("last refresh label should render in the dim role (ANSI 8): %q", got)
	}
	if !strings.Contains(got, base.Render("100%")) {
		t.Errorf("elements between bullets should stay in the chrome role: %q", got)
	}
	if !strings.Contains(got, base.Render("DB "+formatMB(m.dbSize))) {
		t.Errorf("database size should stay in the chrome role: %q", got)
	}
}

func TestStatusBarHelpHintTrailing(t *testing.T) {
	m, _ := newTestModel(t)
	m.dbSize = 0
	m.list.total = 1
	m.list.cursor = 1
	m.list.groups = []dayGroup{{articles: []articleItem{{ID: 1, Title: "one"}}}}
	got := ansi.Strip(m.renderStatusBar())
	// left = "never refreshed" (15); the right side leads with the `?: help`
	// hint, then the RAM/DB footprints, then the percentage and position.
	b := render.GlyphsFor(m.ascii).Bullet
	right := "?: help " + b + " RAM 0MB " + b + " DB 0MB " + b + " 100% " + b + " 1/1"
	left := "never refreshed"
	pad := m.width - ansi.StringWidth(left) - ansi.StringWidth(right)
	want := left + strings.Repeat(" ", pad) + right
	if got != want {
		t.Errorf("status bar = %q, want %q", got, want)
	}
	if ansi.StringWidth(got) != m.width {
		t.Errorf("status bar width = %d, want %d: %q", ansi.StringWidth(got), m.width, got)
	}
}

func TestExpandToggleFold(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "c", time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC))
	m.loadList()

	pressX := func() {
		m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	}
	allCollapsed := func() bool {
		for i := range m.list.groups {
			if !m.list.groups[i].collapsed {
				return false
			}
		}
		return true
	}
	allExpanded := func() bool {
		for i := range m.list.groups {
			if m.list.groups[i].collapsed {
				return false
			}
		}
		return true
	}

	// All expanded → x collapses all.
	pressX()
	if !allCollapsed() {
		t.Error("x on an all-expanded list should collapse every group")
	}

	// All collapsed → x expands all.
	pressX()
	if !allExpanded() {
		t.Error("x on an all-collapsed list should expand every group")
	}

	// Mixed → x expands all.
	m.collapse(0)
	pressX()
	if !allExpanded() {
		t.Error("x on a mixed list should expand every group")
	}

	// Works from an article row and clamps the cursor when the list shrinks.
	m.list.cursor = 5 // last article row of the 6 visible rows
	pressX()
	if !allCollapsed() {
		t.Error("x from an article row should collapse every group")
	}
	if n := len(m.visibleRows()); m.list.cursor >= n {
		t.Errorf("cursor = %d out of range for %d visible rows after collapse", m.list.cursor, n)
	}
}

func TestArticleRowShowsLocalTime(t *testing.T) {
	withLocalZone(t, time.FixedZone("UTC-5", -5*60*60))
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	m.loadList()

	item := &m.list.groups[0].articles[0]
	line := ansi.Strip(m.renderArticleRow(item, false))
	if !strings.Contains(line, "10:04") {
		t.Errorf("row = %q, want local time 10:04 (15:04 UTC - 5h)", line)
	}
	if !strings.HasPrefix(line, "10:04 ") {
		t.Errorf("row should start with the publication time: %q", line)
	}
	if !strings.HasSuffix(line, "example") {
		t.Errorf("source should be the only right-aligned field: %q", line)
	}
	if !strings.Contains(line, "timed") {
		t.Errorf("row should contain the title: %q", line)
	}
	if strings.Contains(line, render.GlyphsFor(m.ascii).Bullet) {
		t.Errorf("row should not contain bullets: %q", line)
	}
}

func TestTreeRailBookendGlyphs(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	m.loadList()

	// rows: [header today, a, header yesterday, b]
	lines := strings.Split(ansi.Strip(m.renderList()), "\n")
	if !strings.HasPrefix(lines[0], "┌ ") {
		t.Errorf("first day header should be prefixed with the top corner: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "└ ") {
		t.Errorf("last day header should be prefixed with the bottom corner: %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[3])
	}
}

func TestTreeRailStableWhileScrolling(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	// Two days of 12 articles each = 26 rows; the page window (23) covers
	// neither the first nor the last row.
	for i := 0; i < 12; i++ {
		insertTimedArticle(t, st, fmt.Sprintf("a%02d", i), now.Add(-time.Duration(i)*time.Minute))
		insertTimedArticle(t, st, fmt.Sprintf("b%02d", i), now.Add(-24*time.Hour-time.Duration(i)*time.Minute))
	}
	m.loadList()
	m.list.cursor = 12

	lines := strings.Split(ansi.Strip(m.renderList()), "\n")
	rows := m.visibleRows()
	start, end := listWindow(len(rows), m.list.cursor, m.pageSize())
	for i := start; i < end; i++ {
		line := lines[i-start]
		if rows[i].kind == rowHeader {
			want := m.railGlyph(len(m.list.groups), rows[i].groupIdx)
			if !strings.HasPrefix(line, want+" ") {
				t.Errorf("scrolled window day header %d should use %q, got %q", i, want, line)
			}
		} else if strings.HasPrefix(line, "├") {
			t.Errorf("scrolled window article row %d should carry no tree glyph, got %q", i, line)
		}
	}
}

func TestTreeRailRecomputesWhenCollapsed(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	m.loadList()
	m.collapse(0)

	// rows: [header today, header yesterday, b]
	lines := strings.Split(ansi.Strip(m.renderList()), "\n")
	if !strings.HasPrefix(lines[0], "┌ ") {
		t.Errorf("first day header should stay the top corner: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "└ ") {
		t.Errorf("last day header should take the bottom corner: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[2])
	}
}

func TestTreeRailASCIIFallback(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	m.SetAscii(true)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	m.loadList()

	lines := strings.Split(ansi.Strip(m.renderList()), "\n")
	for i, want := range []string{"+ ", "12:00 ", "+ ", "12:00 "} {
		if !strings.HasPrefix(lines[i], want) {
			t.Errorf("ascii rail row %d = %q, want prefix %q", i, lines[i], want)
		}
	}
}

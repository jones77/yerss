package listview

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal"
	"yerss/internal/ui/render"
)

func withLocalZone(t *testing.T, loc *time.Location) {
	t.Helper()
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

func newTestState(t *testing.T) (*State, *internal.Store) {
	t.Helper()
	st, err := internal.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("internal.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	s := New(app.New(config.Default(), st), render.DarkPalette(), false, 80, 24)
	return s, st
}

func listKeys(t *testing.T) map[string]config.Action {
	t.Helper()
	km, err := config.ParseKeybindings(config.DefaultKeybindings())
	if err != nil {
		t.Fatalf("ParseKeybindings: %v", err)
	}
	keys, err := km.EffectiveKeys(config.ViewList)
	if err != nil {
		t.Fatalf("EffectiveKeys: %v", err)
	}
	return keys
}

func insertTimedArticle(t *testing.T, st *internal.Store, title string, pub time.Time) int64 {
	t.Helper()
	a := internal.Article{
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

func insertArticle(t *testing.T, st *internal.Store, title string, tags []string) int64 {
	t.Helper()
	id := insertTimedArticle(t, st, title, time.Time{})
	if err := st.SetArticleTags(id, tags); err != nil {
		t.Fatalf("SetArticleTags: %v", err)
	}
	return id
}

func mustParseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func forceTrueColor(t *testing.T) {
	t.Helper()
	orig := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(orig) })
}

func TestBucketDayGroupsOrdering(t *testing.T) {
	withLocalZone(t, time.UTC)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	arts := []internal.Article{
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
		internal.LongDate(day(28)) + ", today",
		internal.LongDate(day(27)) + ", yesterday",
		internal.LongDate(day(26)),
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
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "c", time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC))
	s.Load()

	rows := s.visibleRows()
	if len(rows) != 5 {
		t.Fatalf("expanded visible rows = %d, want 5", len(rows))
	}
	if rows[0].kind != rowHeader || rows[1].kind != rowArticle || rows[2].kind != rowArticle ||
		rows[3].kind != rowHeader || rows[4].kind != rowArticle {
		t.Errorf("unexpected expanded row kinds: %+v", rows)
	}

	s.collapse(0)
	rows = s.visibleRows()
	if len(rows) != 3 {
		t.Fatalf("collapsed visible rows = %d, want 3", len(rows))
	}
	if rows[0].kind != rowHeader || rows[1].kind != rowHeader {
		t.Errorf("collapsed group should emit only its header: %+v", rows)
	}
}

func TestCursorWrapsAcrossVisibleRows(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	s.Load()

	s.cursor = 0
	s.moveListCursor(-1)
	if s.cursor != 2 {
		t.Errorf("wrap up: cursor = %d, want 2", s.cursor)
	}
	s.moveListCursor(1)
	if s.cursor != 0 {
		t.Errorf("wrap down: cursor = %d, want 0", s.cursor)
	}
}

func TestPageKeysClampAtBoundaries(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	s.Load()

	keys := listKeys(t)
	s.cursor = 0
	s.Update(tea.KeyMsg{Type: tea.KeyPgDown}, keys)
	if s.cursor != 2 {
		t.Errorf("page down from top: cursor = %d, want 2", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyPgUp}, keys)
	if s.cursor != 0 {
		t.Errorf("page up from bottom: cursor = %d, want 0", s.cursor)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyPgUp}, keys)
	if s.cursor != 0 {
		t.Errorf("page up at top should clamp: cursor = %d, want 0", s.cursor)
	}
	s.cursor = 2
	s.Update(tea.KeyMsg{Type: tea.KeyPgDown}, keys)
	if s.cursor != 2 {
		t.Errorf("page down at bottom should clamp: cursor = %d, want 2", s.cursor)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyCtrlD}, keys)
	if s.cursor != 2 {
		t.Errorf("half page down at bottom should clamp: cursor = %d, want 2", s.cursor)
	}
	s.cursor = 0
	s.Update(tea.KeyMsg{Type: tea.KeyCtrlU}, keys)
	if s.cursor != 0 {
		t.Errorf("half page up at top should clamp: cursor = %d, want 0", s.cursor)
	}
}

func TestOpenAndFoldKeysOnHeader(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	s.Load()

	keys := listKeys(t)
	s.cursor = 0 // header row
	if act := s.Update(tea.KeyMsg{Type: tea.KeyEnter}, keys); act != ActionNone {
		t.Errorf("Enter on a header should fold in place, action = %d", act)
	}
	if act := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, keys); act != ActionNone {
		t.Errorf("l on a header should fold in place, action = %d", act)
	}
	if s.Collapsed(0) {
		t.Error("enter/l on a header should not leave the day folded")
	}
}

func TestFoldKeys(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	s.Load()

	keys := listKeys(t)
	s.cursor = 0 // header
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, keys)
	if !s.Collapsed(0) {
		t.Error("h on a header should collapse the day")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, keys)
	if s.Collapsed(0) {
		t.Error("l on a header should expand the day")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEnter}, keys)
	if !s.Collapsed(0) {
		t.Error("enter on a header should toggle the fold")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyTab}, keys)
	if s.Collapsed(0) {
		t.Error("tab on a header should toggle the fold")
	}
	s.expand(0)
	s.cursor = 1 // first article row
	s.Update(tea.KeyMsg{Type: tea.KeyTab}, keys)
	if !s.Collapsed(0) {
		t.Error("tab on an article row should toggle its day group")
	}
}

func TestStatusBarShowsNT(t *testing.T) {
	s, st := newTestState(t)
	insertArticle(t, st, "one", []string{"tech"})
	insertArticle(t, st, "two", nil)
	s.Load()

	s.cursor = 2
	if got := s.renderStatusBar(); !strings.Contains(got, "100% "+render.GlyphsFor(s.ascii).Bullet+" 2/2") {
		t.Errorf("status bar = %q, want 100%% %s 2/2", got, render.GlyphsFor(s.ascii).Bullet)
	}

	s.SetFilter("tech")
	s.Load()
	s.cursor = 1
	got := s.renderStatusBar()
	if !strings.Contains(got, "100% "+render.GlyphsFor(s.ascii).Bullet+" 1/1") {
		t.Errorf("filtered status bar = %q, want 100%% %s 1/1", got, render.GlyphsFor(s.ascii).Bullet)
	}
	if !strings.Contains(got, "filter tech") {
		t.Errorf("filtered status bar missing filter name: %q", got)
	}
}

func TestStatusBarShowsDBSize(t *testing.T) {
	s, st := newTestState(t)
	s.width = 120 // the DB/percent/position right side is long
	insertArticle(t, st, "one", nil)
	s.Load()

	got := s.renderStatusBar()
	if s.dbSize <= 0 {
		t.Fatalf("expected dbSize set by load, got %d", s.dbSize)
	}
	if !strings.Contains(got, "DB "+render.FormatMB(s.dbSize)) {
		t.Errorf("status bar = %q, want DB %s", got, render.FormatMB(s.dbSize))
	}
	if !strings.Contains(got, "never refreshed") {
		t.Errorf("status bar = %q, want size alongside refresh info", got)
	}

	s.Status.LastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")
	s.cursor = 1 // the single article
	got = s.renderStatusBar()
	if !strings.Contains(got, "15:04 Friday 28th August 2026 last refresh") {
		t.Errorf("status bar = %q, want last refresh time with long date", got)
	}
	if !strings.Contains(got, "DB "+render.FormatMB(s.dbSize)) {
		t.Errorf("status bar = %q, want DB %s", got, render.FormatMB(s.dbSize))
	}
	if !strings.Contains(got, "DB "+render.FormatMB(s.dbSize)+" "+render.GlyphsFor(s.ascii).Bullet+" 100% "+render.GlyphsFor(s.ascii).Bullet+" 1/1") {
		t.Errorf("status bar = %q, want DB %s percentage %s position", got, render.GlyphsFor(s.ascii).Bullet, render.GlyphsFor(s.ascii).Bullet)
	}
}

func TestStatusBarBulletAndRefreshDim(t *testing.T) {
	forceTrueColor(t)
	s, st := newTestState(t)
	insertArticle(t, st, "one", nil)
	s.Load()
	s.cursor = 1
	s.Status.LastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")

	got := s.renderStatusBar()
	dim := lipgloss.NewStyle().Foreground(render.DarkPalette().Dim)
	base := lipgloss.NewStyle().Foreground(render.DarkPalette().StatusBar)
	bullet := render.GlyphsFor(s.ascii).Bullet
	if !strings.Contains(got, dim.Render(bullet)) {
		t.Errorf("bullets should render in the dim role (ANSI 8): %q", got)
	}
	if !strings.Contains(got, dim.Render("last refresh")) {
		t.Errorf("last refresh label should render in the dim role (ANSI 8): %q", got)
	}
	if !strings.Contains(got, base.Render("100%")) {
		t.Errorf("elements between bullets should stay in the chrome role: %q", got)
	}
	if !strings.Contains(got, base.Render("DB "+render.FormatMB(s.dbSize))) {
		t.Errorf("database size should stay in the chrome role: %q", got)
	}
}

func TestStatusBarHelpHintTrailing(t *testing.T) {
	s, _ := newTestState(t)
	s.dbSize = 0
	s.cursor = 1
	s.groups = []dayGroup{{articles: []Item{{ID: 1, Title: "one"}}}}
	got := ansi.Strip(s.renderStatusBar())
	b := render.GlyphsFor(s.ascii).Bullet
	right := "?: help " + b + " DB 0MB " + b + " 100% " + b + " 1/1"
	left := "never refreshed"
	pad := s.width - ansi.StringWidth(left) - ansi.StringWidth(right)
	want := left + strings.Repeat(" ", pad) + right
	if got != want {
		t.Errorf("status bar = %q, want %q", got, want)
	}
	if ansi.StringWidth(got) != s.width {
		t.Errorf("status bar width = %d, want %d: %q", ansi.StringWidth(got), s.width, got)
	}
}

func TestExpandToggleFold(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "c", time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC))
	s.Load()

	keys := listKeys(t)
	pressX := func() {
		s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}, keys)
	}
	allCollapsed := func() bool {
		for i := range s.groups {
			if !s.groups[i].collapsed {
				return false
			}
		}
		return true
	}
	allExpanded := func() bool {
		for i := range s.groups {
			if s.groups[i].collapsed {
				return false
			}
		}
		return true
	}

	pressX()
	if !allCollapsed() {
		t.Error("x on an all-expanded list should collapse every group")
	}

	pressX()
	if !allExpanded() {
		t.Error("x on an all-collapsed list should expand every group")
	}

	s.collapse(0)
	pressX()
	if !allExpanded() {
		t.Error("x on a mixed list should expand every group")
	}

	s.cursor = 5 // last article row of the 6 visible rows
	pressX()
	if !allCollapsed() {
		t.Error("x from an article row should collapse every group")
	}
	if n := len(s.visibleRows()); s.cursor >= n {
		t.Errorf("cursor = %d out of range for %d visible rows after collapse", s.cursor, n)
	}
}

func TestArticleRowShowsLocalTime(t *testing.T) {
	withLocalZone(t, time.FixedZone("UTC-5", -5*60*60))
	s, st := newTestState(t)
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	s.Load()

	item := &s.groups[0].articles[0]
	line := ansi.Strip(s.renderArticleRow(item, false))
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
	if strings.Contains(line, render.GlyphsFor(s.ascii).Bullet) {
		t.Errorf("row should not contain bullets: %q", line)
	}
}

func TestArticleRowShowsSecondsWhenEnabled(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	s.sess.Config().Display.ShowSeconds = true
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	s.Load()

	item := &s.groups[0].articles[0]
	line := ansi.Strip(s.renderArticleRow(item, false))
	if !strings.HasPrefix(line, "15:04:05 ") {
		t.Errorf("row should start with the seconds publication time: %q", line)
	}
	if strings.HasPrefix(line, "15:04 ") {
		t.Errorf("row should not use minute precision when seconds are enabled: %q", line)
	}
}

func TestStatusBarShowsSecondsWhenEnabled(t *testing.T) {
	s, st := newTestState(t)
	s.sess.Config().Display.ShowSeconds = true
	insertArticle(t, st, "one", nil)
	s.Load()
	s.Status.LastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")
	s.cursor = 1

	got := s.renderStatusBar()
	if !strings.Contains(got, "15:04:05 Friday 28th August 2026 last refresh") {
		t.Errorf("status bar = %q, want seconds in the last refresh time", got)
	}
}

func TestTreeRailBookendGlyphs(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	s.Load()

	lines := strings.Split(ansi.Strip(s.Render()), "\n")
	if !strings.HasPrefix(lines[0], "┌──── ") {
		t.Errorf("first day header should be prefixed with the top corner: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "├──── ") {
		t.Errorf("last day header should be a T-junction while expanded: %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[3])
	}
}

func TestTreeRailStableWhileScrolling(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 12; i++ {
		insertTimedArticle(t, st, fmt.Sprintf("a%02d", i), now.Add(-time.Duration(i)*time.Minute))
		insertTimedArticle(t, st, fmt.Sprintf("b%02d", i), now.Add(-24*time.Hour-time.Duration(i)*time.Minute))
	}
	s.Load()
	s.cursor = 12

	lines := strings.Split(ansi.Strip(s.Render()), "\n")
	rows := s.visibleRows()
	start, end := listWindow(len(rows), s.cursor, s.pageSize())
	for i := start; i < end; i++ {
		line := lines[i-start]
		if rows[i].kind == rowHeader {
			want := s.railGlyph(len(s.groups), rows[i].groupIdx)
			if !strings.HasPrefix(line, want+"──── ") {
				t.Errorf("scrolled window day header %d should use %q, got %q", i, want, line)
			}
		} else if strings.HasPrefix(line, "├") {
			t.Errorf("scrolled window article row %d should carry no tree glyph, got %q", i, line)
		}
	}
}

func TestTreeRailRecomputesWhenCollapsed(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	s.Load()
	s.collapse(0)

	lines := strings.Split(ansi.Strip(s.Render()), "\n")
	if !strings.HasPrefix(lines[0], "┌──── ") {
		t.Errorf("first day header should stay the top corner: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "├──── ") {
		t.Errorf("last expanded day header should take the T-junction: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "12:00 ") {
		t.Errorf("article row should carry no tree glyph and start with the time: %q", lines[2])
	}
}

func TestTreeRailASCIIFallback(t *testing.T) {
	withLocalZone(t, time.UTC)
	s, st := newTestState(t)
	s.ascii = true
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	insertTimedArticle(t, st, "a", now)
	insertTimedArticle(t, st, "b", now.Add(-24*time.Hour))
	s.Load()

	lines := strings.Split(ansi.Strip(s.Render()), "\n")
	for i, want := range []string{"+---- ", "12:00 ", "+---- ", "12:00 "} {
		if !strings.HasPrefix(lines[i], want) {
			t.Errorf("ascii rail row %d = %q, want prefix %q", i, lines[i], want)
		}
	}
}

func TestListWindow(t *testing.T) {
	cases := []struct {
		name               string
		n, cursor, height  int
		wantStart, wantEnd int
	}{
		{"all fit", 5, 2, 10, 0, 5},
		{"empty list", 0, 0, 10, 0, 0},
		{"exact fit", 10, 5, 10, 0, 10},
		{"cursor at top", 20, 0, 10, 0, 10},
		{"cursor middle", 20, 10, 10, 5, 15},
		{"cursor at bottom", 20, 19, 10, 10, 20},
		{"height one at top", 20, 0, 1, 0, 1},
		{"height one at bottom", 20, 19, 1, 19, 20},
	}
	for _, c := range cases {
		start, end := listWindow(c.n, c.cursor, c.height)
		if start != c.wantStart || end != c.wantEnd {
			t.Errorf("%s: listWindow(%d,%d,%d) = (%d,%d), want (%d,%d)",
				c.name, c.n, c.cursor, c.height, start, end, c.wantStart, c.wantEnd)
		}
	}
}

func TestRenderListZeroHeight(t *testing.T) {
	s, _ := newTestState(t)
	s.SetSize(0, 0)
	if got := len(strings.Split(s.Render(), "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s.Render())
	}
}

func TestRenderListEmptyHeightOne(t *testing.T) {
	s, _ := newTestState(t)
	s.SetSize(0, 1)
	if got := len(strings.Split(s.Render(), "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s.Render())
	}
}

func TestRenderListNonEmptyHeightOne(t *testing.T) {
	s, st := newTestState(t)
	insertArticle(t, st, "one", nil)
	s.Load()
	s.SetSize(0, 1)
	if got := len(strings.Split(s.Render(), "\n")); got != 1 {
		t.Errorf("expected 1 line, got %d: %q", got, s.Render())
	}
}

func TestRenderListZeroWidth(t *testing.T) {
	s, st := newTestState(t)
	insertArticle(t, st, "one", nil)
	s.Load()
	s.SetSize(0, 24)
	if got := s.Render(); got == "" {
		t.Error("expected non-empty render")
	}
}
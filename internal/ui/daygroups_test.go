package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
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
		got := longDate(tm)
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
	arts := []store.Article{
		{Title: "day1-1", PublishedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)},
		{Title: "day2-1", PublishedAt: time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)},
		{Title: "day2-2", PublishedAt: time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)},
		{Title: "day3-1", PublishedAt: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)},
		{Title: "undated"},
	}
	groups := bucketDayGroups(arts)
	if len(groups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(groups))
	}
	day := func(d int) time.Time { return time.Date(2026, 8, d, 0, 0, 0, 0, time.UTC) }
	if groups[0].label != longDate(day(28)) || groups[1].label != longDate(day(27)) || groups[2].label != longDate(day(26)) {
		t.Errorf("groups out of descending order: %q, %q, %q", groups[0].label, groups[1].label, groups[2].label)
	}
	if groups[3].label != "Undated" {
		t.Errorf("last group should be Undated, got %q", groups[3].label)
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

	if got := m.renderStatusBar(); !strings.Contains(got, "2/2 articles") {
		t.Errorf("status bar = %q, want 2/2 articles", got)
	}

	m.popupData = popupState{
		tags:   []store.TagCount{{Name: "tech", Total: 1, Unread: 1}},
		cursor: 0,
	}
	m.confirmTagSelection()
	got := m.renderStatusBar()
	if !strings.Contains(got, "1/2 articles") {
		t.Errorf("filtered status bar = %q, want 1/2 articles", got)
	}
	if !strings.Contains(got, "filter tech") {
		t.Errorf("filtered status bar missing filter name: %q", got)
	}
}

func TestStatusBarShowsDBSize(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()

	got := m.renderStatusBar()
	if m.dbSize <= 0 {
		t.Fatalf("expected dbSize set by loadList, got %d", m.dbSize)
	}
	if !strings.Contains(got, formatSize(m.dbSize)) {
		t.Errorf("status bar = %q, want size %s", got, formatSize(m.dbSize))
	}
	if !strings.Contains(got, "never refreshed") {
		t.Errorf("status bar = %q, want size alongside refresh info", got)
	}

	m.lastRefreshedAt = mustParseTime(t, "2026-08-28T15:04:05Z")
	got = m.renderStatusBar()
	if !strings.Contains(got, "last refresh") {
		t.Errorf("status bar = %q, want last refresh after setting time", got)
	}
	if !strings.Contains(got, formatSize(m.dbSize)) {
		t.Errorf("status bar = %q, want size with refresh date", got)
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
	if !strings.HasSuffix(line, "·example·10:04") {
		t.Errorf("source and time should be right-aligned: %q", line)
	}
	if !strings.Contains(line, "timed") {
		t.Errorf("row should contain the title: %q", line)
	}
	if !strings.HasPrefix(line, "  ") {
		t.Errorf("row should keep the cursor gutter: %q", line)
	}
}
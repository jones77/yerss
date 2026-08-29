package ui

import (
	"path/filepath"
	"testing"

	"yerss/internal/config"
	"yerss/internal/store"
)

func newTestModel(t *testing.T) (*Model, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	m := New(config.Default(), st)
	m.width = 80
	m.height = 24
	return m, st
}

func insertArticle(t *testing.T, st *store.Store, title string, tags []string) int64 {
	t.Helper()
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-" + title,
		Title:   title,
		Content: "<p>" + title + "</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle: %v", err)
	}
	if err := st.SetArticleTags(id, tags); err != nil {
		t.Fatalf("SetArticleTags: %v", err)
	}
	return id
}

func TestTagCountText(t *testing.T) {
	cases := []struct {
		unread, total int
		plain, bold   string
	}{
		{60, 60, "", "(60)"},
		{0, 60, "(60)", ""},
		{15, 60, "/60)", "(15"},
	}
	for _, c := range cases {
		plain, bold := tagCountText(c.unread, c.total)
		if plain != c.plain || bold != c.bold {
			t.Errorf("tagCountText(%d,%d) = (%q,%q), want (%q,%q)",
				c.unread, c.total, plain, bold, c.plain, c.bold)
		}
	}
}

func TestTagFiltering(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "a1", []string{"tech"})
	insertArticle(t, st, "a2", []string{"news"})

	m.loadList()
	if m.articleCount() != 2 {
		t.Fatalf("expected 2 articles, got %d", m.articleCount())
	}

	m.popupData = popupState{
		tags:   []store.TagCount{{Name: "tech", Total: 1, Unread: 1}},
		cursor: 0,
	}
	m.confirmTagSelection()
	if m.list.filter != "tech" {
		t.Errorf("filter = %q, want tech", m.list.filter)
	}
	if m.articleCount() != 1 || !containsID(loadedIDs(m), id1) {
		t.Fatalf("expected only tech article, got %v", loadedIDs(m))
	}

	m.clearFilter()
	if m.list.filter != "" {
		t.Errorf("filter after clear = %q", m.list.filter)
	}
	if m.articleCount() != 2 {
		t.Fatalf("expected all articles after clear, got %d", m.articleCount())
	}
}

func loadedIDs(m *Model) []int64 {
	var ids []int64
	for gi := range m.list.groups {
		for ai := range m.list.groups[gi].articles {
			ids = append(ids, m.list.groups[gi].articles[ai].ID)
		}
	}
	return ids
}

func containsID(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestPopupWrapAround(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData = popupState{
		tags: []store.TagCount{
			{Name: "a"}, {Name: "b"},
		},
		cursor: 1,
	}
	m.movePopupCursor(1)
	if m.popupData.cursor != 0 {
		t.Errorf("wrap-down: cursor = %d, want 0", m.popupData.cursor)
	}
	m.movePopupCursor(-1)
	if m.popupData.cursor != 1 {
		t.Errorf("wrap-up: cursor = %d, want 1", m.popupData.cursor)
	}
}

func TestOpenArticleStaleCursorNoPanic(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.loadList()

	m.list.cursor = len(m.visibleRows()) + 5
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with stale-high cursor changed view to %d", m.view)
	}

	m.list.cursor = -1
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with negative cursor changed view to %d", m.view)
	}

	m.list.groups = nil
	m.list.cursor = 0
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle on empty list changed view to %d", m.view)
	}
}
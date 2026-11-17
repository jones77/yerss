package ui

import (
	"path/filepath"
	"testing"

	"yerss/internal"
	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/ui/popup"
)

func newTestModel(t *testing.T) (*Model, *internal.Store) {
	t.Helper()
	st, err := internal.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("internal.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	m := New(app.New(config.Default(), st))
	m.width = 80
	m.height = 24
	return m, st
}

func insertArticle(t *testing.T, st *internal.Store, title string, tags []string) int64 {
	t.Helper()
	a := internal.Article{
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

func TestTagFiltering(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "a1", []string{"tech"})
	insertArticle(t, st, "a2", []string{"news"})

	m.list.Load()
	if m.list.ArticleCount() != 2 {
		t.Fatalf("expected 2 articles, got %d", m.list.ArticleCount())
	}

	m.popupData.SetKind(popup.KindTagsList)
	m.popupData.SetTags([]internal.TagCount{{Name: "tech", Total: 1, Unread: 1}})
	m.popupData.SetCursor(0)
	m.confirmTagSelection()
	if m.list.Filter() != "tech" {
		t.Errorf("filter = %q, want tech", m.list.Filter())
	}
	if m.list.ArticleCount() != 1 || !containsID(loadedIDs(m), id1) {
		t.Fatalf("expected only tech article, got %v", loadedIDs(m))
	}

	m.list.ClearFilter()
	if m.list.Filter() != "" {
		t.Errorf("filter after clear = %q", m.list.Filter())
	}
	if m.list.ArticleCount() != 2 {
		t.Fatalf("expected all articles after clear, got %d", m.list.ArticleCount())
	}
}

func loadedIDs(m *Model) []int64 {
	return m.list.ArticleIDs()
}

func containsID(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestTagPopupGroupsSourcesFirst(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "a1", []string{"tech"})
	id2 := insertArticle(t, st, "a2", []string{"news"})
	_ = st.SetArticleTagsWithSource(id1, []string{"tech", "nytimes"}, "nytimes")
	_ = st.SetArticleTagsWithSource(id2, []string{"news", "theguardian"}, "theguardian")
	m.openTagPopup(popup.KindTagsList)
	var names []string
	for _, tg := range m.popupData.Tags() {
		names = append(names, tg.Name)
	}
	want := []string{"nytimes", "theguardian", "news", "tech"}
	if len(names) != len(want) {
		t.Fatalf("popup tags = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("popup tag order = %v, want %v (sources first)", names, want)
			break
		}
	}
}

func TestOpenArticleStaleCursorNoPanic(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	m.list.Load()

	m.list.SetCursor(m.list.RowCount() + 5)
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with stale-high cursor changed view to %d", m.view)
	}

	m.list.SetCursor(-1)
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle with negative cursor changed view to %d", m.view)
	}

	m.list.SetFilter("no-such-tag")
	m.list.Load()
	m.openArticle()
	if m.view != viewList {
		t.Errorf("openArticle on empty list changed view to %d", m.view)
	}
}

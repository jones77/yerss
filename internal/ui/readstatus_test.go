package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal"
)

func TestShortArticleMarkedReadOnOpen(t *testing.T) {
	m, st := newTestModel(t)
	a := internal.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g1",
		Title:   "short",
		Content: "<p>just a few words</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Read {
		t.Error("short article should be marked read on open")
	}
}

func TestLongArticleNotMarkedReadOnOpen(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>a long line of text to scroll through</p>\n", 200)
	a := internal.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g2",
		Title:   "long",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Read {
		t.Fatal("long article should NOT be marked read on open")
	}
}

func TestFirstMovementInEachDirectionMarksRead(t *testing.T) {
	keys := map[string]tea.KeyMsg{
		"up":     {Type: tea.KeyUp},
		"down":   {Type: tea.KeyDown},
		"pgup":   {Type: tea.KeyPgUp},
		"pgdn":   {Type: tea.KeyPgDown},
		"halfup": {Type: tea.KeyCtrlU},
		"halfdn": {Type: tea.KeyCtrlD},
		"top":    {Type: tea.KeyRunes, Runes: []rune("1")},
		"bottom": {Type: tea.KeyRunes, Runes: []rune("G")},
	}
	for dir, key := range keys {
		name := dir
		key := key
		t.Run(name, func(t *testing.T) {
			m, st := newTestModel(t)
			content := strings.Repeat("<p>scrolling text content here</p>\n", 200)
			a := internal.Article{
				FeedURL: "https://example.com/feed.xml",
				GUID:    "g-move-" + name,
				Title:   "move-" + name,
				Content: content,
			}
			id, err := st.UpsertArticle(a)
			if err != nil {
				t.Fatal(err)
			}
			a.ID = id
			m.article = m.article.Recompose(a)
			if got, _ := st.GetArticle(id); got.Read {
				t.Fatal("precondition: article should start unread")
			}
			m.updateArticle(key)
			got, err := st.GetArticle(id)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Read {
				t.Errorf("first %s movement should mark a long article read", name)
			}
		})
	}
}

func TestMovementMarksReadWhenPositionUnchanged(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>text at the top of a long article</p>\n", 200)
	a := internal.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g4",
		Title:   "top",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	if m.article.Viewport.YOffset != 0 {
		t.Fatal("precondition: article should start at the top")
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyUp})
	if got, err := st.GetArticle(id); err != nil {
		t.Fatal(err)
	} else if !got.Read {
		t.Error("a movement should mark read even when the position does not change")
	}
}

func TestFurtherMovementKeepsArticleRead(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>more scrolling text content here</p>\n", 200)
	a := internal.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g5",
		Title:   "read",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyDown})
	if got, err := st.GetArticle(id); err != nil {
		t.Fatal(err)
	} else if !got.Read {
		t.Fatal("precondition: first movement should mark read")
	}
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyPgDown},
		{Type: tea.KeyUp},
	} {
		m.updateArticle(key)
		got, err := st.GetArticle(id)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Read {
			t.Errorf("key %s should leave an already-read article read", key)
		}
	}
}

func TestRestoreSavedOffsetDoesNotMarkRead(t *testing.T) {
	st := openSharedStore(t)
	a := internal.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g6",
		Title:   "restore",
		Content: strings.Repeat("<p>a long line to restore a saved offset over</p>\n", 200),
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id

	m1 := modelOn(t, st)
	m1.article = m1.article.Recompose(a)
	m1.view = viewArticle
	m1.article.Viewport.ScrollDown(7)
	m1.persistSelection()
	// The session ended before the offset restore; the article is still unread.
	if err := st.SetRead(id, false); err != nil {
		t.Fatal(err)
	}

	m2 := modelOn(t, st)
	m2.list.Load()
	m2.restoreSelection()
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Read {
		t.Error("restoring a saved offset should not mark read")
	}
}

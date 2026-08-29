package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/store"
)

func TestShortArticleMarkedReadOnOpen(t *testing.T) {
	m, st := newTestModel(t)
	a := store.Article{
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
	m.article = m.newArticleState(a)
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Read {
		t.Error("short article should be marked read on open")
	}
}

func TestLongArticleNotMarkedOnOpen(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>a long line of text to scroll through</p>\n", 200)
	a := store.Article{
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
	m.article = m.newArticleState(a)
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Read {
		t.Fatal("long article should NOT be marked read on open")
	}

	// First downward scroll (down arrow) marks it read.
	m.updateArticle(tea.KeyMsg{Type: tea.KeyDown})
	got, err = st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Read {
		t.Error("long article should be marked read on first downward scroll")
	}
}

func TestDownwardScrollKeysMarkRead(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>more scrolling text content here</p>\n", 200)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g3",
		Title:   "scroll",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id

	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyPgDown},
		{Type: tea.KeyCtrlF},
		{Type: tea.KeyCtrlD},
		{Type: tea.KeySpace},
	} {
		_ = st.SetRead(id, false)
		m.article = m.newArticleState(a)
		m.updateArticle(key)
		got, _ := st.GetArticle(id)
		if !got.Read {
			t.Errorf("key %s should mark read on downward scroll", key)
		}
	}
}

func TestUpwardScrollDoesNotMarkRead(t *testing.T) {
	m, st := newTestModel(t)
	content := strings.Repeat("<p>upward text content</p>\n", 200)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g4",
		Title:   "up",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.newArticleState(a)
	m.article.viewport.ScrollDown(2)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyUp})
	got, _ := st.GetArticle(id)
	if got.Read {
		t.Error("upward scroll should not mark read")
	}
}
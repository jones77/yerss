package ui

import (
	"strings"
	"testing"
)

func TestNextPreviousArticleNavigation(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "one", nil)
	id2 := insertArticle(t, st, "two", nil)
	id3 := insertArticle(t, st, "three", nil)
	m.list.Load()

	// The list is newest first, so the display order is [id3, id2, id1].
	if ids := m.list.ArticleIDs(); len(ids) != 3 || ids[0] != id3 || ids[1] != id2 || ids[2] != id1 {
		t.Fatalf("ArticleIDs() = %v, want [%d %d %d]", ids, id3, id2, id1)
	}

	m.list.SelectArticleID(id2)
	m.openArticle()
	if m.article.ID() != id2 {
		t.Fatalf("opened article %d, want %d", m.article.ID(), id2)
	}

	// K is next = newer = up the list.
	sendKey(m, 'K')
	if m.article.ID() != id3 {
		t.Fatalf("K opened article %d, want %d", m.article.ID(), id3)
	}
	// J is previous = older = down the list.
	sendKey(m, 'J')
	if m.article.ID() != id2 {
		t.Fatalf("J opened article %d, want %d", m.article.ID(), id2)
	}
}

func TestNextPreviousWrapAtEnds(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "one", nil)
	insertArticle(t, st, "two", nil)
	id3 := insertArticle(t, st, "three", nil)
	m.list.Load()

	// At the newest article, K wraps to the oldest.
	m.list.SelectArticleID(id3)
	m.openArticle()
	sendKey(m, 'K')
	if m.article.ID() != id1 {
		t.Fatalf("K at the newest article opened %d, want oldest %d", m.article.ID(), id1)
	}

	// At the oldest article, J wraps to the newest.
	m.list.SelectArticleID(id1)
	m.openArticle()
	sendKey(m, 'J')
	if m.article.ID() != id3 {
		t.Fatalf("J at the oldest article opened %d, want newest %d", m.article.ID(), id3)
	}
}

func TestNavigationStaysWithinFilter(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "one", []string{"tech"})
	id2 := insertArticle(t, st, "two", []string{"news"})
	id3 := insertArticle(t, st, "three", []string{"tech"})

	m.list.SetFilter("tech")
	m.list.Load()
	if ids := m.list.ArticleIDs(); len(ids) != 2 {
		t.Fatalf("filtered ArticleIDs() = %v, want 2 articles", ids)
	}

	m.list.SelectArticleID(id1)
	m.openArticle()
	for _, key := range []rune{'K', 'J', 'K', 'J'} {
		sendKey(m, key)
		if got := m.article.ID(); got != id1 && got != id3 {
			t.Fatalf("%q left the filter: opened article %d (news=%d)", key, got, id2)
		}
	}
}

func TestNavigationNoOpWhenArticleNotInList(t *testing.T) {
	m, st := newTestModel(t)
	id1 := insertArticle(t, st, "one", []string{"tech"})
	insertArticle(t, st, "two", []string{"news"})
	m.list.Load()

	m.list.SelectArticleID(id1)
	m.openArticle()

	// Re-filter so the open article is no longer in the list.
	m.list.SetFilter("news")
	m.list.Load()

	sendKey(m, 'J')
	if m.article.ID() != id1 {
		t.Fatalf("J moved to %d although the open article is not in the list", m.article.ID())
	}
	sendKey(m, 'K')
	if m.article.ID() != id1 {
		t.Fatalf("K moved to %d although the open article is not in the list", m.article.ID())
	}
}

func TestArticleTopBorderShowsPosition(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	id2 := insertArticle(t, st, "two", nil)
	insertArticle(t, st, "three", nil)
	m.list.Load()

	m.list.SelectArticleID(id2)
	m.openArticle()

	lines := strings.Split(m.View(), "\n")
	if !strings.Contains(lines[0], "2/3") {
		t.Errorf("top border missing position 2/3: %q", lines[0])
	}
}

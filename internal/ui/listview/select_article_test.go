package listview

import "testing"

func TestSelectArticleIDExpandsCollapsedGroup(t *testing.T) {
	s, st := newTestState(t)
	id1 := insertArticle(t, st, "one", nil)
	id2 := insertArticle(t, st, "two", nil)
	s.Load()

	if got := s.ArticleIDs(); len(got) != 2 || got[0] != id2 || got[1] != id1 {
		t.Fatalf("ArticleIDs() = %v, want [%d %d]", got, id2, id1)
	}

	s.expandToggle() // every group is expanded, so this collapses them all
	if !s.CursorIsHeader() {
		t.Fatal("expected the cursor on a header after collapsing every group")
	}

	if !s.SelectArticleID(id1) {
		t.Fatal("SelectArticleID(id1) = false, want true")
	}
	if s.CursorIsHeader() {
		t.Error("SelectArticleID should land the cursor on the article row")
	}
	if got := s.ArticleIDAtRow(s.Cursor()); got != id1 {
		t.Errorf("cursor article = %d, want %d", got, id1)
	}

	if s.SelectArticleID(1 << 30) {
		t.Error("SelectArticleID(unknown) = true, want false")
	}
}

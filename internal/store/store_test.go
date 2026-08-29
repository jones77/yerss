package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func sampleArticle() Article {
	return Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "guid-1",
		Title:       "Hello",
		Link:        "https://example.com/1",
		Author:      "alice",
		PublishedAt: time.Now().Add(-time.Hour),
		Content:     "<p>v1</p>",
		Description: "desc",
		Categories:  []string{"tech"},
	}
}

func TestUpsertArticleDeduplicates(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	id1, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle: %v", err)
	}
	a.Title = "Updated"
	a.Content = "<p>v2</p>"
	id2, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle second: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("expected same article id, got %d and %d", id1, id2)
	}
	articles, err := st.ListArticles("")
	if err != nil {
		t.Fatalf("ListArticles: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if articles[0].Title != "Updated" {
		t.Errorf("expected title updated, got %q", articles[0].Title)
	}
	if articles[0].Content != "<p>v2</p>" {
		t.Errorf("expected content updated, got %q", articles[0].Content)
	}
}

func TestUpsertArticleInsertsNewGuid(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	if _, err := st.UpsertArticle(a); err != nil {
		t.Fatal(err)
	}
	a.GUID = "guid-2"
	if _, err := st.UpsertArticle(a); err != nil {
		t.Fatal(err)
	}
	articles, err := st.ListArticles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(articles))
	}
}

func TestReadStatusPersistsAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restart.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	id, err := st.UpsertArticle(sampleArticle())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetRead(id, true); err != nil {
		t.Fatal(err)
	}
	st.Close()

	st2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	got, err := st2.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Read {
		t.Error("expected read status true after restart")
	}
}

func TestArticleCategoriesStored(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetArticleTags(id, []string{"tech", "news"}); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %v", got.Categories)
	}
}

func TestListTagsCounts(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	id1, _ := st.UpsertArticle(a)
	a.GUID = "g2"
	id2, _ := st.UpsertArticle(a)
	_ = st.SetArticleTags(id1, []string{"tech"})
	_ = st.SetArticleTags(id2, []string{"tech", "news"})
	_ = st.SetRead(id1, true)

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", tags)
	}
	// tech: total 2, unread 1; news: total 1, unread 1
	byName := map[string]TagCount{}
	for _, tg := range tags {
		byName[tg.Name] = tg
	}
	if tech := byName["tech"]; tech.Total != 2 || tech.Unread != 1 || tech.Read != 1 {
		t.Errorf("tech counts wrong: %+v", tech)
	}
	if news := byName["news"]; news.Total != 1 || news.Unread != 1 || news.Read != 0 {
		t.Errorf("news counts wrong: %+v", news)
	}
}

func TestHasVerifiedFeedRequiresFetchTime(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	if _, err := st.UpsertArticle(a); err != nil {
		t.Fatal(err)
	}
	// A stub feed row created by article persistence has no last_fetched_at and
	// must not satisfy the gate's offline trust rule.
	verified, err := st.HasVerifiedFeed()
	if err != nil {
		t.Fatal(err)
	}
	if verified {
		t.Error("stub-only feed row must not count as a verified feed")
	}
	// After a real fetch (UpsertFeed stamps last_fetched_at), the gate trusts.
	if err := st.UpsertFeed(a.FeedURL, "feed", time.Now()); err != nil {
		t.Fatal(err)
	}
	verified, err = st.HasVerifiedFeed()
	if err != nil {
		t.Fatal(err)
	}
	if !verified {
		t.Error("feed with a recorded fetch time should count as verified")
	}
}

func TestVerifiedFeedURLs(t *testing.T) {
	st := newTestStore(t)

	urls, err := st.VerifiedFeedURLs()
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 0 {
		t.Errorf("empty table should yield an empty map, got %v", urls)
	}

	a := sampleArticle()
	if _, err := st.UpsertArticle(a); err != nil {
		t.Fatal(err)
	}
	urls, err = st.VerifiedFeedURLs()
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 0 {
		t.Errorf("stub feed row (NULL last_fetched_at) must not count as verified, got %v", urls)
	}

	if err := st.UpsertFeed(a.FeedURL, "feed", time.Now()); err != nil {
		t.Fatal(err)
	}
	urls, err = st.VerifiedFeedURLs()
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected exactly 1 verified URL, got %v", urls)
	}
	if !urls[a.FeedURL] {
		t.Errorf("fetched feed %q should be verified, got %v", a.FeedURL, urls)
	}
}

func TestArticleCount(t *testing.T) {
	st := newTestStore(t)

	n, err := st.ArticleCount()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("empty store count = %d, want 0", n)
	}

	for i := 0; i < 3; i++ {
		a := sampleArticle()
		a.GUID = fmt.Sprintf("g%d", i)
		if _, err := st.UpsertArticle(a); err != nil {
			t.Fatal(err)
		}
	}

	n, err = st.ArticleCount()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("count = %d, want 3", n)
	}

	arts, err := st.ListArticles("")
	if err != nil {
		t.Fatal(err)
	}
	if n != len(arts) {
		t.Errorf("ArticleCount %d does not match len(ListArticles) %d", n, len(arts))
	}
}

func TestDBSizeIncludesWal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "size.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpsertArticle(sampleArticle()); err != nil {
		t.Fatal(err)
	}
	// Close checkpoints and detaches; the main file is stable afterwards and
	// DBSize only stats files, so it still works.
	st.Close()

	mainFi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Ensure a clean "no wal" baseline regardless of whether Close left one.
	_ = os.Remove(path + "-wal")

	base, err := st.DBSize()
	if err != nil {
		t.Fatalf("DBSize baseline: %v", err)
	}
	if base != mainFi.Size() {
		t.Errorf("DBSize baseline = %d, want main file size %d", base, mainFi.Size())
	}

	wal := path + "-wal"
	if err := os.WriteFile(wal, make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	total, err := st.DBSize()
	if err != nil {
		t.Fatalf("DBSize with wal: %v", err)
	}
	if want := mainFi.Size() + 4096; total != want {
		t.Errorf("DBSize with wal = %d, want %d", total, want)
	}

	if err := os.Remove(wal); err != nil {
		t.Fatal(err)
	}
	noWal, err := st.DBSize()
	if err != nil {
		t.Fatalf("DBSize after wal removal: %v", err)
	}
	if noWal != mainFi.Size() {
		t.Errorf("DBSize after wal removal = %d, want main size %d", noWal, mainFi.Size())
	}
}

func TestDBSizeMissingMainFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DBSize(); err == nil {
		t.Error("expected error when main DB file is missing")
	}
}
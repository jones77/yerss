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

func TestSourceLabel(t *testing.T) {
	cases := []struct {
		link, feed, want string
	}{
		{"https://newrepublic.com/story/1", "", "newrepublic"},
		{"https://www.nytimes.com/x", "", "nytimes"},
		{"https://a.lot.of.subdomains.nytimes.com/x", "", "nytimes"},
		{"https://reallylongnewspaperdomainname.net/x", "", "reallylongnewspaperdomainname"},
		{"https://tribunemag.com./story/1", "", "tribunemag"},
		{"https://WWW.TRIBUNEMAG.COM/x", "", "tribunemag"},
		{"https://tribunemag.co.uk/story/1", "", "tribunemag"},
		{"https://www.bbc.co.uk/news/x", "", "bbc"},
		{"", "https://nytimes.com/rss", "nytimes"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := SourceLabel(c.link, c.feed); got != c.want {
			t.Errorf("SourceLabel(%q, %q) = %q, want %q", c.link, c.feed, got, c.want)
		}
	}
}

func TestSourceDomainTagListedWithCounts(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	a.FeedURL = "https://www.nytimes.com/rss"
	a.Link = "https://www.nytimes.com/1"
	id1, _ := st.UpsertArticle(a)
	a.GUID = "g2"
	a.Link = "https://www.nytimes.com/2"
	id2, _ := st.UpsertArticle(a)
	a.GUID = "g3"
	a.Link = "https://www.theguardian.com/1"
	id3, _ := st.UpsertArticle(a)
	for _, id := range []int64{id1, id2} {
		_ = st.SetArticleTagsWithSource(id, []string{"tech", "nytimes"}, "nytimes")
	}
	_ = st.SetArticleTagsWithSource(id3, []string{"news", "theguardian"}, "theguardian")
	_ = st.SetRead(id1, true)

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]TagCount{}
	var order []string
	for _, tg := range tags {
		byName[tg.Name] = tg
		order = append(order, tg.Name)
	}
	if ny := byName["nytimes"]; ny.Total != 2 || ny.Unread != 1 || !ny.IsSource {
		t.Errorf("nytimes counts wrong: %+v", ny)
	}
	if gu := byName["theguardian"]; gu.Total != 1 || gu.Unread != 1 || !gu.IsSource {
		t.Errorf("theguardian counts wrong: %+v", gu)
	}
	if tech := byName["tech"]; tech.IsSource {
		t.Errorf("category tag tech should not be a source: %+v", tech)
	}
	// Source-domain tags are grouped ahead of category tags.
	lastSource := -1
	firstCategory := len(order)
	for i, name := range order {
		if byName[name].IsSource {
			lastSource = i
		} else if firstCategory == len(order) {
			firstCategory = i
		}
	}
	if lastSource > firstCategory {
		t.Errorf("source tags should come before category tags, order: %v", order)
	}
	// The domain tag filters articles like any category tag.
	arts, err := st.ListArticles("nytimes")
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Errorf("ListArticles(nytimes) = %d articles, want 2", len(arts))
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

func TestListTagsAlphabeticalSecondarySort(t *testing.T) {
	st := newTestStore(t)
	// Equal-popularity tags sort alphabetically, case-insensitively: the byte
	// (asciibetical) order would put "Beta" and "Zoo" ahead of "apple" because
	// uppercase ASCII sorts first.
	for i, name := range []string{"Zoo", "apple", "Beta", "banana"} {
		a := sampleArticle()
		a.GUID = fmt.Sprintf("guid-%d", i)
		id, _ := st.UpsertArticle(a)
		_ = st.SetArticleTags(id, []string{name})
	}
	// A more popular tag sorts ahead of equal-popularity tags regardless of
	// spelling.
	a := sampleArticle()
	a.GUID = "guid-popular"
	id1, _ := st.UpsertArticle(a)
	a.GUID = "guid-popular-2"
	id2, _ := st.UpsertArticle(a)
	_ = st.SetArticleTags(id1, []string{"technology"})
	_ = st.SetArticleTags(id2, []string{"technology"})

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, tg := range tags {
		order = append(order, tg.Name)
	}
	want := []string{"technology", "apple", "banana", "Beta", "Zoo"}
	if len(order) != len(want) {
		t.Fatalf("ListTags order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("ListTags order = %v, want %v (alphabetical, case-insensitive)", order, want)
			break
		}
	}
}

func TestTagsCaseInsensitiveCasePreserving(t *testing.T) {
	st := newTestStore(t)
	// "energy" first, then "ENERGY": a single tag keeps the first spelling.
	a := sampleArticle()
	a.GUID = "g1"
	id1, _ := st.UpsertArticle(a)
	a.GUID = "g2"
	id2, _ := st.UpsertArticle(a)
	_ = st.SetArticleTags(id1, []string{"energy"})
	_ = st.SetArticleTags(id2, []string{"ENERGY"})

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected one tag, got %v", tags)
	}
	if tags[0].Name != "energy" || tags[0].Total != 2 {
		t.Errorf("tag = %+v, want energy with total 2 (first spelling preserved)", tags[0])
	}
	// Filtering is case-insensitive too, matching any spelling.
	arts, err := st.ListArticles("ENERGY")
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Errorf("ListArticles(ENERGY) = %d articles, want 2", len(arts))
	}
}

func TestTagsCaseInsensitiveFirstSeenWins(t *testing.T) {
	st := newTestStore(t)
	// "ENERGY" first, then "energy": the first spelling stays forever.
	a := sampleArticle()
	a.GUID = "g1"
	id1, _ := st.UpsertArticle(a)
	a.GUID = "g2"
	id2, _ := st.UpsertArticle(a)
	_ = st.SetArticleTags(id1, []string{"ENERGY"})
	_ = st.SetArticleTags(id2, []string{"energy"})

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Name != "ENERGY" || tags[0].Total != 2 {
		t.Fatalf("tags = %+v, want a single ENERGY tag with total 2", tags)
	}
}

func TestTagsCaseInsensitiveSourceFlag(t *testing.T) {
	st := newTestStore(t)
	// The category spelling differs in case from the source label; both must
	// resolve to one tag that carries the source flag with the first spelling.
	a := sampleArticle()
	id, _ := st.UpsertArticle(a)
	_ = st.SetArticleTagsWithSource(id, []string{"Nytimes"}, "nytimes")
	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || !tags[0].IsSource {
		t.Fatalf("tags = %v, want a single source tag", tags)
	}
	if tags[0].Name != "Nytimes" {
		t.Errorf("first spelling should be preserved, got %q", tags[0].Name)
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
func TestImageColumnsMigratedAndIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "img.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	cols, err := st.tableColumns("articles")
	if err != nil {
		t.Fatal(err)
	}
	if !cols["image_url"] || !cols["image_data"] {
		t.Errorf("first migration missing image columns: %v", cols)
	}
	st.Close()

	// Re-opening the same database must be a no-op (idempotent).
	st2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	cols, err = st2.tableColumns("articles")
	if err != nil {
		t.Fatal(err)
	}
	if !cols["image_url"] || !cols["image_data"] {
		t.Errorf("re-migration dropped image columns: %v", cols)
	}
}

func TestImageColumnsExistingRowsRetainData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.sqlite")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := sampleArticle()
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	// The article was written without an image URL; its row must survive the
	// additive migration with null image fields.
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
	if got.Title != a.Title || got.Content != a.Content {
		t.Errorf("existing row data lost: %+v", got)
	}
	if got.ImageURL != "" {
		t.Errorf("legacy row image_url = %q, want empty", got.ImageURL)
	}
	data, err := st2.GetImageData(id)
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Errorf("legacy row image_data = %v, want nil", data)
	}
}

func TestImageURLCapturedAndRestored(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	a.ImageURL = "https://example.com/lead.jpg"
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageURL != a.ImageURL {
		t.Errorf("image_url = %q, want %q", got.ImageURL, a.ImageURL)
	}
	list, err := st.ListArticles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ImageURL != a.ImageURL {
		t.Errorf("list scan image_url wrong: %+v", list)
	}
	// Clearing the URL on re-upsert persists NULL.
	a.ImageURL = ""
	if _, err := st.UpsertArticle(a); err != nil {
		t.Fatal(err)
	}
	got, err = st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageURL != "" {
		t.Errorf("cleared image_url = %q, want empty", got.ImageURL)
	}
}

func TestImageDataRoundtrip(t *testing.T) {
	st := newTestStore(t)
	id, err := st.UpsertArticle(sampleArticle())
	if err != nil {
		t.Fatal(err)
	}
	before, err := st.GetImageData(id)
	if err != nil {
		t.Fatal(err)
	}
	if before != nil {
		t.Errorf("fresh article image_data = %v, want nil", before)
	}
	want := []byte("fake-jpeg-bytes")
	if err := st.SetImageData(id, want); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetImageData(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("image_data roundtrip = %q, want %q", got, want)
	}
}

func TestLastSelectionRoundtrip(t *testing.T) {
	st := newTestStore(t)

	zero, err := st.LoadLastSelection()
	if err != nil {
		t.Fatalf("LoadLastSelection on empty store: %v", err)
	}
	if zero != (LastSelection{}) {
		t.Errorf("expected zero selection, got %+v", zero)
	}

	sel := LastSelection{View: "article", ArticleID: 42}
	if err := st.SaveLastSelection(sel); err != nil {
		t.Fatalf("SaveLastSelection: %v", err)
	}
	got, err := st.LoadLastSelection()
	if err != nil {
		t.Fatalf("LoadLastSelection: %v", err)
	}
	if got != sel {
		t.Errorf("roundtrip = %+v, want %+v", got, sel)
	}

	sel2 := LastSelection{View: "list", HeaderKey: "20260828"}
	if err := st.SaveLastSelection(sel2); err != nil {
		t.Fatalf("SaveLastSelection second: %v", err)
	}
	got, err = st.LoadLastSelection()
	if err != nil {
		t.Fatalf("LoadLastSelection second: %v", err)
	}
	if got != sel2 {
		t.Errorf("overwrite roundtrip = %+v, want %+v", got, sel2)
	}
}

func TestUpsertArticleReturnedIDMatchesRow(t *testing.T) {
	st := newTestStore(t)
	a := sampleArticle()
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id {
		t.Errorf("GetArticle(id).ID = %d, want %d", got.ID, id)
	}
	a.Title = "Updated"
	id2, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	if id2 != id {
		t.Errorf("conflict-update returned id %d, want %d", id2, id)
	}
}

func TestSetArticleTagsDedupesAndSkipsEmpty(t *testing.T) {
	st := newTestStore(t)
	id, err := st.UpsertArticle(sampleArticle())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetArticleTags(id, []string{"tech", "tech", "", "news", "news"}); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("categories = %v, want 2", got.Categories)
	}
	if got.Categories[0] != "news" || got.Categories[1] != "tech" {
		t.Errorf("categories = %v, want [news tech]", got.Categories)
	}
}

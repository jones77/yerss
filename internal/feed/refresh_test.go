package feed

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"yerss/internal/store"
)

func newTestFeedStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestNeedsStartupRefresh(t *testing.T) {
	g := Gate{MinInterval: 15 * time.Minute, Cooldown: 60 * time.Second}
	now := time.Now()

	cases := []struct {
		name string
		last time.Time
		want bool
	}{
		{"never refreshed", time.Time{}, true},
		{"older than interval", now.Add(-20 * time.Minute), true},
		{"exactly interval", now.Add(-15 * time.Minute), false},
		{"within interval", now.Add(-10 * time.Minute), false},
	}
	for _, c := range cases {
		if got := g.NeedsStartupRefresh(c.last, now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestManualRefreshCooldown(t *testing.T) {
	g := Gate{MinInterval: 15 * time.Minute, Cooldown: 60 * time.Second}
	now := time.Now()

	if ok, _ := g.ManualRefreshAllowed(time.Time{}, now); !ok {
		t.Error("manual refresh should be allowed when never refreshed")
	}
	if ok, _ := g.ManualRefreshAllowed(now.Add(-90*time.Second), now); !ok {
		t.Error("manual refresh should be allowed after cooldown")
	}
	// Manual refresh bypasses the 15-minute startup interval.
	if ok, _ := g.ManualRefreshAllowed(now.Add(-2*time.Minute), now); !ok {
		t.Error("manual refresh should bypass the 15-minute interval")
	}
	ok, remaining := g.ManualRefreshAllowed(now.Add(-30*time.Second), now)
	if ok {
		t.Error("manual refresh should be blocked within cooldown")
	}
	if remaining <= 0 || remaining > 30*time.Second {
		t.Errorf("expected remaining ~30s, got %v", remaining)
	}
}

// TestFetchFeedsHungServerReturns verifies FetchFeeds applies a bounded
// per-feed timeout: a server that accepts the connection but never responds is
// recorded as an error and the pass returns instead of hanging.
func TestFetchFeedsHungServerReturns(t *testing.T) {
	origPer, origOverall := refreshPerFeedTimeout, refreshOverallTimeout
	refreshPerFeedTimeout, refreshOverallTimeout = 200*time.Millisecond, 2*time.Second
	defer func() { refreshPerFeedTimeout, refreshOverallTimeout = origPer, origOverall }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	st := newTestFeedStore(t)
	start := time.Now()
	res := FetchFeeds(st, []string{srv.URL})
	elapsed := time.Since(start)

	if res.Feeds != 0 {
		t.Errorf("expected 0 feeds parsed, got %d", res.Feeds)
	}
	if len(res.Errors) == 0 {
		t.Error("expected the hung feed to be recorded as an error")
	}
	if elapsed > 2*time.Second {
		t.Errorf("FetchFeeds exceeded overall deadline: %v", elapsed)
	}
}

// TestVerifyFeedsOverallDeadline verifies the overall deadline cancels
// in-flight fetches that are still hanging.
func TestVerifyFeedsOverallDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	st := newTestFeedStore(t)
	start := time.Now()
	res, err := VerifyFeeds(st, []string{srv.URL, srv.URL}, 5*time.Second, 150*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error when no feed parses successfully")
	}
	if res.Feeds != 0 {
		t.Errorf("expected 0 feeds parsed, got %d", res.Feeds)
	}
	if elapsed > time.Second {
		t.Errorf("VerifyFeeds did not respect the overall deadline: %v", elapsed)
	}
}
// TestItemToArticleFallsBackToGUIDLink covers Atom feeds whose entries carry
// the permalink in <id> but no <link> element (e.g. jacobin.com): the article
// link must fall back to the GUID so the reader can open it.
func TestItemToArticleFallsBackToGUIDLink(t *testing.T) {
	item := &gofeed.Item{
		GUID:    "https://jacobin.com/2026/08/some-article",
		Title:   "Some Article",
		Content: "<p>body</p>",
	}
	a := itemToArticle("https://jacobin.com/feed", item)
	if a.Link != item.GUID {
		t.Errorf("Link = %q, want the entry id %q", a.Link, item.GUID)
	}
}

func TestItemToArticleKeepsRealLink(t *testing.T) {
	item := &gofeed.Item{
		GUID: "https://example.com/2026/08/id",
		Link: "https://example.com/2026/08/article",
	}
	a := itemToArticle("https://example.com/feed", item)
	if a.Link != item.Link {
		t.Errorf("Link = %q, want %q", a.Link, item.Link)
	}
}

func TestItemToArticleNonURLGUIDNoFallback(t *testing.T) {
	item := &gofeed.Item{GUID: "tag:example.com,2026:article-1"}
	a := itemToArticle("https://example.com/feed", item)
	if a.Link != "" {
		t.Errorf("Link should stay empty for a non-URL GUID, got %q", a.Link)
	}
}

func TestLeadImageURLFromEnclosure(t *testing.T) {
	item := &gofeed.Item{
		Enclosures: []*gofeed.Enclosure{
			{Type: "audio/mpeg", URL: "https://example.com/pod.mp3"},
			{Type: "image/jpeg", URL: "https://example.com/lead.jpg"},
			{Type: "image/png", URL: "https://example.com/other.png"},
		},
	}
	if got := leadImageURL(item); got != "https://example.com/lead.jpg" {
		t.Errorf("leadImageURL = %q, want the first image enclosure", got)
	}
}

func TestLeadImageURLFromMediaThumbnail(t *testing.T) {
	item := &gofeed.Item{
		Extensions: map[string]map[string][]ext.Extension{
			"media": {
				"thumbnail": {{Attrs: map[string]string{"url": "https://example.com/thumb.png"}}},
			},
		},
	}
	if got := leadImageURL(item); got != "https://example.com/thumb.png" {
		t.Errorf("leadImageURL = %q, want the media thumbnail", got)
	}
}

func TestLeadImageURLIgnoresNonImageEnclosure(t *testing.T) {
	item := &gofeed.Item{
		Enclosures: []*gofeed.Enclosure{{Type: "audio/mpeg", URL: "https://example.com/pod.mp3"}},
	}
	if got := leadImageURL(item); got != "" {
		t.Errorf("non-image enclosure should not set a lead image, got %q", got)
	}
}

func TestLeadImageURLNotFromInlineContent(t *testing.T) {
	item := &gofeed.Item{Content: `<p>see <img src="https://cdn.example.com/track.png"> now</p>`}
	if got := leadImageURL(item); got != "" {
		t.Errorf("inline content <img> must not become a lead image, got %q", got)
	}
}

func TestItemToArticleCapturesLeadImage(t *testing.T) {
	item := &gofeed.Item{
		GUID:      "g1",
		Title:     "t",
		Content:   "<p>body</p>",
		Enclosures: []*gofeed.Enclosure{{Type: "image/jpeg", URL: "https://example.com/lead.jpg"}},
	}
	a := itemToArticle("https://example.com/feed", item)
	if a.ImageURL != "https://example.com/lead.jpg" {
		t.Errorf("ImageURL = %q, want the enclosure URL", a.ImageURL)
	}
}

// TestFetchFeedsAtomEntryIDAsLink is an end-to-end check that a Jacobin-style
// Atom feed (entry <id> is the permalink, no <link>) stores an article whose
// Link is the entry id.
func TestFetchFeedsAtomEntryIDAsLink(t *testing.T) {
	atom := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Jacobin</title>
  <link href="https://jacobin.com"/>
  <entry>
    <id>https://jacobin.com/2026/08/colombia-earthquake</id>
    <title>Colombia's Far Right</title>
    <updated>2026-08-28T16:20:17Z</updated>
    <content type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml"><p>body</p></div></content>
  </entry>
</feed>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, atom)
	}))
	defer srv.Close()

	st := newTestFeedStore(t)
	res := FetchFeeds(st, []string{srv.URL})
	if res.Feeds != 1 {
		t.Fatalf("expected 1 feed parsed, got %d", res.Feeds)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	arts, err := st.ListArticles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 1 {
		t.Fatalf("expected 1 article, got %d", len(arts))
	}
	if arts[0].Link != "https://jacobin.com/2026/08/colombia-earthquake" {
		t.Errorf("article Link = %q, want the entry id URL", arts[0].Link)
	}
}

func TestFetchFeedsRecordsOutcomes(t *testing.T) {
	st := newTestFeedStore(t)

	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>OK</title><item><title>a</title><guid>g1</guid></item><item><title>b</title><guid>g2</guid></item></channel></rss>`)
	}))
	defer okSrv.Close()

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>Empty</title></channel></rss>`)
	}))
	defer emptySrv.Close()

	notFoundSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer notFoundSrv.Close()

	res := FetchFeeds(st, []string{okSrv.URL, emptySrv.URL, notFoundSrv.URL})
	if len(res.Outcomes) != 3 {
		t.Fatalf("outcomes = %d, want 3: %+v", len(res.Outcomes), res.Outcomes)
	}

	byURL := map[string]FeedOutcome{}
	for _, o := range res.Outcomes {
		byURL[o.URL] = o
	}
	if o := byURL[okSrv.URL]; o.Err != nil || o.Articles != 2 {
		t.Errorf("ok feed outcome = %+v, want 2 articles and no error", o)
	}
	if o := byURL[emptySrv.URL]; o.Err != nil || o.Articles != 0 {
		t.Errorf("empty feed outcome = %+v, want 0 articles and no error", o)
	}
	if o := byURL[notFoundSrv.URL]; o.Status != http.StatusNotFound || o.Err == nil || o.Articles != 0 {
		t.Errorf("404 feed outcome = %+v, want status 404 with an error", o)
	}
}

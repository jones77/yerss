package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setFeedsFile writes the given URLs to a temp feeds.txt and points the model
// at it.
func setFeedsFile(t *testing.T, m *Model, urls []string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "feeds.txt")
	content := ""
	for _, u := range urls {
		content += u + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m.cfg.Data.FeedsFile = path
}

func TestInitForcesRefreshOnNewFeed(t *testing.T) {
	m, st := newTestModel(t)
	// A verified feed makes LastRefreshedAt recent, so the 15-minute gate alone
	// would skip the refresh.
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}
	// feeds.txt gains a brand-new URL with no verified row.
	setFeedsFile(t, m, []string{"https://example.com/new.xml"})

	if cmd := m.Init(); cmd == nil {
		t.Error("expected a refresh command when a configured feed has never been fetched")
	}
}

func TestInitSkipsRefreshWhenAllFeedsVerified(t *testing.T) {
	m, st := newTestModel(t)
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}
	// feeds.txt contains only the already-verified URL.
	setFeedsFile(t, m, []string{"https://example.com/known.xml"})

	if cmd := m.Init(); cmd != nil {
		t.Error("expected no refresh command when all feeds are verified and the last refresh was recent")
	}
}

func TestInitEndToEndFetchesNewFeed(t *testing.T) {
	m, st := newTestModel(t)
	// A verified feed makes LastRefreshedAt recent, so the 15-minute gate alone
	// would skip the startup refresh.
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>Test Feed</title><link>http://example.com/</link><description>test</description><item><title>Hello from new feed</title><link>http://example.com/item/1</link><guid>e2e-guid-1</guid><pubDate>Sat, 28 Aug 2026 12:00:00 GMT</pubDate></item></channel></rss>`)
	}))
	defer srv.Close()

	// feeds.txt gains a brand-new URL pointing at the test server.
	feedsURL := srv.URL + "/feed.xml"
	setFeedsFile(t, m, []string{feedsURL})

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected a refresh command for the never-fetched feed")
	}

	msg := cmd()
	if got := msg.(refreshFinishedMsg); got.err != nil {
		t.Fatalf("startup refresh failed: %v", got.err)
	}

	exists, err := st.ArticleExists(feedsURL, "e2e-guid-1")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected the new feed's article to be stored after the startup refresh")
	}
}
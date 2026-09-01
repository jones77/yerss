package ui

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yerss/internal/feed"
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
	m.sess.Config().Data.FeedsFile = path
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

func TestInitSkipsRefreshForSchemelessKnownFeed(t *testing.T) {
	m, st := newTestModel(t)
	// The verified row stores the canonical https:// URL.
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}
	// feeds.txt lists the same feed scheme-less after the format change.
	setFeedsFile(t, m, []string{"example.com/known.xml"})

	if cmd := m.Init(); cmd != nil {
		t.Error("a scheme-less entry matching a verified feed should not force a refresh")
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

	// Scheme-less feeds.txt entries are looked up over HTTPS, so the test
	// server is TLS and the feed client trusts its self-signed certificate.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>Test Feed</title><link>http://example.com/</link><description>test</description><item><title>Hello from new feed</title><link>http://example.com/item/1</link><guid>e2e-guid-1</guid><pubDate>Sat, 28 Aug 2026 12:00:00 GMT</pubDate></item></channel></rss>`)
	}))
	oldFactory := feed.ClientFactory
	feed.ClientFactory = func(timeout time.Duration) *http.Client {
		return &http.Client{
			Timeout:   timeout,
			// Test server only: the self-signed certificate is trusted by design.
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
		}
	}
	t.Cleanup(func() { feed.ClientFactory = oldFactory; srv.Close() })

	// feeds.txt gains a brand-new scheme-less entry pointing at the test
	// server.
	feedsURL := feed.CanonicalURL(strings.TrimPrefix(srv.URL, "https://") + "/feed.xml")
	setFeedsFile(t, m, []string{strings.TrimPrefix(srv.URL, "https://") + "/feed.xml"})

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
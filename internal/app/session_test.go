package app

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

	"yerss/internal"
	"yerss/internal/config"
	"yerss/internal/feed"
)

func newTestSession(t *testing.T) (*Session, *internal.Store) {
	t.Helper()
	st, err := internal.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("internal.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(config.Default(), st), st
}

func setFeedsFile(t *testing.T, s *Session, urls []string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "feeds.txt")
	content := ""
	for _, u := range urls {
		content += u + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s.cfg.Data.FeedsFile = path
}

func TestRefreshLoadErrorSurfaces(t *testing.T) {
	s, _ := newTestSession(t)
	// A nonexistent feeds file must be returned as an error, not a fetch result.
	s.cfg.Data.FeedsFile = filepath.Join(t.TempDir(), "missing.txt")

	res, err := s.Refresh()
	if err == nil {
		t.Fatal("expected a feeds-file load error")
	}
	if res != nil {
		t.Fatalf("expected nil result on load error, got %+v", res)
	}
}

func TestRefreshFetchesFeeds(t *testing.T) {
	s, st := newTestSession(t)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>Test Feed</title><link>http://example.com/</link><description>test</description><item><title>Hello</title><link>http://example.com/item/1</link><guid>app-test-guid</guid></item></channel></rss>`)
	}))
	t.Cleanup(srv.Close)
	oldFactory := feed.ClientFactory
	feed.ClientFactory = func(timeout time.Duration) *http.Client {
		return &http.Client{
			Timeout:   timeout,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
		}
	}
	t.Cleanup(func() { feed.ClientFactory = oldFactory })

	feedsURL := feed.CanonicalURL(strings.TrimPrefix(srv.URL, "https://") + "/feed.xml")
	setFeedsFile(t, s, []string{strings.TrimPrefix(srv.URL, "https://") + "/feed.xml"})

	res, err := s.Refresh()
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if res.Feeds != 1 || res.New != 1 {
		t.Fatalf("Refresh result = %+v, want one new feed/article", res)
	}
	exists, err := st.ArticleExists(feedsURL, "app-test-guid")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected the fetched article to be stored")
	}
}

func TestNeedsStartupRefreshTimeGate(t *testing.T) {
	s, _ := newTestSession(t)
	// An empty feeds file means "no new feeds"; the time gate alone decides.
	setFeedsFile(t, s, nil)
	now := time.Now()

	if !s.NeedsStartupRefresh(time.Time{}, now) {
		t.Error("a zero last-refresh time should force a startup refresh")
	}
	if s.NeedsStartupRefresh(now, now) {
		t.Error("a recent refresh should not trigger the startup gate")
	}
	if !s.NeedsStartupRefresh(now.Add(-16*time.Minute), now) {
		t.Error("a stale refresh beyond MinInterval should trigger the startup gate")
	}
}

func TestNeedsStartupRefreshUnfetchedFeedBypassesGate(t *testing.T) {
	s, st := newTestSession(t)
	// A verified feed makes last-refresh recent, so the gate alone would skip.
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}
	// feeds.txt gains a brand-new URL with no verified row.
	setFeedsFile(t, s, []string{"https://example.com/new.xml"})

	now := time.Now()
	if !s.NeedsStartupRefresh(now, now) {
		t.Error("a never-fetched feed should bypass the time gate")
	}
}

func TestNeedsStartupRefreshSchemelessKnownFeed(t *testing.T) {
	s, st := newTestSession(t)
	if err := st.UpsertFeed("https://example.com/known.xml", "known", time.Now()); err != nil {
		t.Fatal(err)
	}
	// feeds.txt lists the same feed scheme-less after the format change.
	setFeedsFile(t, s, []string{"example.com/known.xml"})

	now := time.Now()
	if s.NeedsStartupRefresh(now, now) {
		t.Error("a scheme-less entry matching a verified feed should not force a refresh")
	}
}

func TestManualRefreshAllowed(t *testing.T) {
	s, _ := newTestSession(t)
	now := time.Now()

	if ok, remaining := s.ManualRefreshAllowed(time.Time{}, now); !ok || remaining != 0 {
		t.Errorf("zero last-refresh time should allow a manual refresh, got ok=%v remaining=%v", ok, remaining)
	}
	if ok, remaining := s.ManualRefreshAllowed(now, now); ok || remaining <= 0 {
		t.Errorf("a just-completed refresh should be throttled by the cooldown, got ok=%v remaining=%v", ok, remaining)
	}
	if ok, remaining := s.ManualRefreshAllowed(now.Add(-2*time.Minute), now); !ok || remaining != 0 {
		t.Errorf("a refresh older than the cooldown should be allowed, got ok=%v remaining=%v", ok, remaining)
	}
}

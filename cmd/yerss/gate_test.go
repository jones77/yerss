package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yerss/internal/config"
	"yerss/internal/store"
)

func gateTestEnv(t *testing.T, feedsContent string) (*config.Config, *store.Store) {
	t.Helper()
	dir := t.TempDir()
	feedsPath := filepath.Join(dir, "feeds.txt")
	if err := os.WriteFile(feedsPath, []byte(feedsContent), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Data.FeedsFile = feedsPath
	cfg.Data.Dir = filepath.Join(dir, "data")
	st, err := store.Open(cfg.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return cfg, st
}

func TestGateEmptyFeedsRefuses(t *testing.T) {
	cfg, st := gateTestEnv(t, "# nothing here\n\n")
	if code := runStartupGate(cfg, st); code != exitEmptyFeeds {
		t.Fatalf("expected exit %d, got %d", exitEmptyFeeds, code)
	}
}

func TestGateMissingFeedsRefuses(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Data.FeedsFile = filepath.Join(dir, "absent.txt")
	cfg.Data.Dir = filepath.Join(dir, "data")
	st, err := store.Open(cfg.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if code := runStartupGate(cfg, st); code != exitEmptyFeeds {
		t.Fatalf("expected exit %d, got %d", exitEmptyFeeds, code)
	}
}

func TestGateTrustedFeedSkipsNetwork(t *testing.T) {
	unreachable := "http://127.0.0.1:1/feed.xml"
	cfg, st := gateTestEnv(t, unreachable+"\n")
	if err := st.UpsertFeed(unreachable, "previously verified", time.Now()); err != nil {
		t.Fatal(err)
	}
	// Port 1 never accepts connections: reaching this exit 0 proves the gate
	// did not attempt a fetch.
	if code := runStartupGate(cfg, st); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
}

func TestGateFirstRunWorkingFeedStarts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Test Feed</title><item><title>Hello</title><guid>g1</guid></item></channel></rss>`)
	}))
	defer srv.Close()

	cfg, st := gateTestEnv(t, srv.URL+"\n")
	if code := runStartupGate(cfg, st); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	verified, err := st.HasVerifiedFeed()
	if err != nil || !verified {
		t.Fatalf("expected verified feed persisted, verified=%v err=%v", verified, err)
	}
	last, err := st.LastRefreshedAt()
	if err != nil || last.IsZero() {
		t.Fatalf("expected last_fetched_at stamped, last=%v err=%v", last, err)
	}
}

func TestGateFirstRunAllBadRefuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg, st := gateTestEnv(t, srv.URL+"\n")
	if code := runStartupGate(cfg, st); code != exitNoWorking {
		t.Fatalf("expected exit %d, got %d", exitNoWorking, code)
	}
}

func TestGatePrintsVerifyingNotice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>t</title></channel></rss>`)
	}))
	defer srv.Close()

	cfg, st := gateTestEnv(t, srv.URL+"\n")
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	code := runStartupGate(cfg, st)
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = old

	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if !strings.Contains(string(out), "verifying feeds...") {
		t.Errorf("missing verifying notice on stderr: %q", string(out))
	}
}

func TestGateUnreadableFeedsFileRefuses(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission checks are bypassed")
	}
	dir := t.TempDir()
	feedsPath := filepath.Join(dir, "feeds.txt")
	if err := os.WriteFile(feedsPath, []byte("https://example.com/feed.xml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(feedsPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(feedsPath, 0o644) })

	cfg := config.Default()
	cfg.Data.FeedsFile = feedsPath
	cfg.Data.Dir = filepath.Join(dir, "data")
	st, err := store.Open(cfg.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	code := runStartupGate(cfg, st)
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = old

	if code != 1 {
		t.Fatalf("expected exit 1 for unreadable feeds file, got %d", code)
	}
	if !strings.Contains(string(out), "cannot read feeds file") {
		t.Errorf("missing read-error message on stderr: %q", string(out))
	}
	if strings.Contains(string(out), "no RSS feeds configured") {
		t.Errorf("read error must not reuse the empty-feeds message: %q", string(out))
	}
}
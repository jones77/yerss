package feed

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

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
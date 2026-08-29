package ui

import (
	"errors"
	"testing"
	"time"

	"yerss/internal/feed"
)

func TestModelStoresLastRefreshOutcomes(t *testing.T) {
	m, _ := newTestModel(t)

	first := feed.FetchResult{Outcomes: []feed.FeedOutcome{
		{URL: "https://a.com/rss", Status: 404, Err: errors.New("404 Not Found")},
	}}
	m.Update(refreshFinishedMsg{result: &first, fetchedAt: time.Now()})
	got := m.FeedOutcomes()
	if len(got) != 1 || got[0].URL != "https://a.com/rss" || got[0].Status != 404 {
		t.Fatalf("outcomes after first pass = %+v", got)
	}

	// A later pass replaces the recorded outcomes wholesale.
	second := feed.FetchResult{New: 1, Outcomes: []feed.FeedOutcome{
		{URL: "https://b.com/rss", Articles: 3},
	}}
	m.Update(refreshFinishedMsg{result: &second, fetchedAt: time.Now()})
	got = m.FeedOutcomes()
	if len(got) != 1 || got[0].URL != "https://b.com/rss" || got[0].Articles != 3 {
		t.Fatalf("outcomes after second pass = %+v", got)
	}

	// A pass that fails to start (no result) leaves the outcomes standing.
	m.Update(refreshFinishedMsg{err: errors.New("boom")})
	got = m.FeedOutcomes()
	if len(got) != 1 || got[0].URL != "https://b.com/rss" {
		t.Fatalf("outcomes after failed pass = %+v, want the last successful pass", got)
	}
}

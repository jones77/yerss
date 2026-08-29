package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"yerss/internal/feed"
)

func TestFeedDiagnostics(t *testing.T) {
	base := "yerss"
	cases := []struct {
		name     string
		outcomes []feed.FeedOutcome
		want     string
	}{
		{"none", nil, ""},
		{"http error", []feed.FeedOutcome{{URL: "https://x.com/rss", Status: 404, Err: fmt.Errorf("404")}},
			"yerss: error: 404, url: https://x.com/rss\n"},
		{"other error", []feed.FeedOutcome{{URL: "https://x.com/rss", Err: errors.New("connection refused")}},
			"yerss: error: connection refused, url: https://x.com/rss\n"},
		{"no articles", []feed.FeedOutcome{{URL: "https://x.com/rss", Articles: 0}},
			"yerss: warning: no articles returned, url: https://x.com/rss\n"},
		{"healthy is silent", []feed.FeedOutcome{{URL: "https://x.com/rss", Articles: 12}}, ""},
		{"mixed", []feed.FeedOutcome{
			{URL: "https://a.com/rss", Status: 404, Err: fmt.Errorf("404")},
			{URL: "https://b.com/rss", Articles: 3},
			{URL: "https://c.com/rss", Articles: 0},
		}, "yerss: error: 404, url: https://a.com/rss\n" +
			"yerss: warning: no articles returned, url: https://c.com/rss\n"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		feedDiagnostics(&buf, base, c.outcomes)
		if buf.String() != c.want {
			t.Errorf("%s: diagnostics = %q, want %q", c.name, buf.String(), c.want)
		}
	}
}

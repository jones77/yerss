package ui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"yerss/internal/store"
)

func withLocalZone(t *testing.T, loc *time.Location) {
	t.Helper()
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

func insertTimedArticle(t *testing.T, st *store.Store, title string, pub time.Time) int64 {
	t.Helper()
	a := store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "guid-" + title,
		Title:       title,
		PublishedAt: pub,
		Content:     "<p>" + title + "</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatalf("UpsertArticle: %v", err)
	}
	return id
}

// strippedLines strips ANSI escapes from each rendered line.
func strippedLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
}
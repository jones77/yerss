package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yerss/internal/config"
	"yerss/internal/store"
)

func jsonTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Data.Dir = filepath.Join(dir, "data")
	st, err := store.Open(cfg.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// withLocal sets the process local timezone for the test and restores it.
func withLocal(t *testing.T, loc *time.Location) {
	t.Helper()
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

func insertJSONArticle(t *testing.T, st *store.Store, guid string, pub time.Time, cats []string) {
	t.Helper()
	id, err := st.UpsertArticle(store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        guid,
		Title:       "Title " + guid,
		Link:        "https://example.com/" + guid,
		Author:      "alice",
		PublishedAt: pub,
		Content:     "<p>body " + guid + "</p>",
		Description: "desc " + guid,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) > 0 {
		if err := st.SetArticleTags(id, cats); err != nil {
			t.Fatal(err)
		}
	}
}

// captureStdout runs fn with stdout redirected to a pipe and returns the
// captured output plus fn's exit code.
func captureStdout(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b), code
}

func TestJSONDumpThreeDaysDescending(t *testing.T) {
	withLocal(t, time.UTC)
	st := jsonTestStore(t)
	// 20260828 (newest)
	insertJSONArticle(t, st, "a1", time.Date(2026, 8, 28, 15, 0, 0, 0, time.UTC), nil)
	// 20260827
	insertJSONArticle(t, st, "b1", time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC), nil)
	// 20260826, two articles so within-day order matters
	insertJSONArticle(t, st, "c1", time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC), nil)
	insertJSONArticle(t, st, "c2", time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC), nil)

	out, code := captureStdout(t, func() int { return runJSONDump(config.Default(), st) })
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (out %q)", code, out)
	}

	d1, d2, d3 := strings.Index(out, `"20260828"`), strings.Index(out, `"20260827"`), strings.Index(out, `"20260826"`)
	if d1 < 0 || d2 < 0 || d3 < 0 {
		t.Fatalf("missing date keys in output: %s", out)
	}
	if !(d1 < d2 && d2 < d3) {
		t.Errorf("date keys not in descending order: %s", out)
	}

	sub := out[d3:]
	if strings.Index(sub, "2026-08-26T12:00:00Z") > strings.Index(sub, "2026-08-26T09:00:00Z") {
		t.Errorf("articles within a day not descending: %s", sub)
	}
}

func TestJSONDumpArticleFieldsAndCategories(t *testing.T) {
	withLocal(t, time.UTC)
	st := jsonTestStore(t)
	insertJSONArticle(t, st, "a1", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC), nil)
	insertJSONArticle(t, st, "a2", time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC), []string{"tech", "news"})

	out, _ := captureStdout(t, func() int { return runJSONDump(config.Default(), st) })

	var m map[string][]articleJSON
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	day := m["20260828"]
	if len(day) != 2 {
		t.Fatalf("expected 2 articles on the day, got %d", len(day))
	}

	byGUID := map[string]articleJSON{}
	for _, a := range day {
		byGUID[a.GUID] = a
	}
	plain := byGUID["a1"]
	if plain.ID == 0 || plain.FeedURL == "" || plain.Title != "Title a1" ||
		plain.Link == "" || plain.Author != "alice" || plain.PublishedAt == "" ||
		plain.Content == "" || plain.Description != "desc a1" || plain.Read {
		t.Errorf("stored fields not all present: %+v", plain)
	}
	if len(plain.Categories) != 0 {
		t.Errorf("article without categories should have [] categories, got %v", plain.Categories)
	}
	if !strings.Contains(out, `"categories":[]`) {
		t.Errorf("categories should serialize as [] not null: %s", out)
	}

	tagged := byGUID["a2"]
	if len(tagged.Categories) != 2 || tagged.Categories[0] != "news" || tagged.Categories[1] != "tech" {
		t.Errorf("categories not loaded from store: %v", tagged.Categories)
	}
}

func TestJSONDumpLocalTimezoneKeying(t *testing.T) {
	withLocal(t, time.FixedZone("UTC-5", -5*60*60))
	st := jsonTestStore(t)
	// Midnight UTC on Aug 28 is 19:00 Aug 27 local, so the article must key to
	// the previous local calendar day.
	insertJSONArticle(t, st, "a1", time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC), nil)

	out, _ := captureStdout(t, func() int { return runJSONDump(config.Default(), st) })
	if strings.Contains(out, `"20260828"`) {
		t.Errorf("UTC-midnight article should not key to 20260828: %s", out)
	}
	if !strings.Contains(out, `"20260827"`) {
		t.Errorf("UTC-midnight article should key to the previous local day 20260827: %s", out)
	}

	var m map[string][]articleJSON
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	a := m["20260827"][0]
	if a.PublishedAt != "2026-08-27T19:00:00-05:00" {
		t.Errorf("published_at should be RFC3339 with offset, got %q", a.PublishedAt)
	}
}

func TestJSONDumpEmptyStore(t *testing.T) {
	st := jsonTestStore(t)
	out, code := captureStdout(t, func() int { return runJSONDump(config.Default(), st) })
	if code != 0 {
		t.Fatalf("expected exit 0 for an empty store, got %d", code)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("empty store should emit {}, got %q", out)
	}
}

func TestJSONDumpRFC3339Times(t *testing.T) {
	withLocal(t, time.UTC)
	st := jsonTestStore(t)
	insertJSONArticle(t, st, "a1", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC), nil)

	out, _ := captureStdout(t, func() int { return runJSONDump(config.Default(), st) })
	var m map[string][]articleJSON
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	a := m["20260828"][0]
	if a.PublishedAt != "2026-08-28T15:04:05Z" {
		t.Errorf("published_at not RFC3339, got %q", a.PublishedAt)
	}
	if a.FetchedAt == "" {
		t.Error("fetched_at should be an RFC3339 string, got empty")
	}
}
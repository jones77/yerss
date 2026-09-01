package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/timeutil"
)

// articleJSON is the JSON representation of a stored article. Times are emitted
// as RFC3339 strings so they are self-describing, and categories is always a
// JSON array (empty when the article has none).
type articleJSON struct {
	ID          int64    `json:"id"`
	FeedURL     string   `json:"feed_url"`
	GUID        string   `json:"guid"`
	Title       string   `json:"title"`
	Link        string   `json:"link"`
	Author      string   `json:"author"`
	PublishedAt string   `json:"published_at"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Read        bool     `json:"read"`
	FetchedAt   string   `json:"fetched_at"`
	Categories  []string `json:"categories"`
}

func toArticleJSON(a store.Article) articleJSON {
	cats := a.Categories
	if cats == nil {
		cats = []string{}
	}
	return articleJSON{
		ID:          a.ID,
		FeedURL:     a.FeedURL,
		GUID:        a.GUID,
		Title:       a.Title,
		Link:        a.Link,
		Author:      a.Author,
		PublishedAt: formatRFC3339(a.PublishedAt),
		Content:     a.Content,
		Description: a.Description,
		Read:        a.Read,
		FetchedAt:   formatRFC3339(a.FetchedAt),
		Categories:  cats,
	}
}

func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// dateBucket groups articles under one local calendar day.
type dateBucket struct {
	date string
	arts []articleJSON
}

// dateBuckets marshals as a JSON object whose keys appear in slice order, so
// the date keys are emitted in the order they were bucketed (reverse
// chronological) rather than Go's lexicographically-sorted map ordering.
type dateBuckets []dateBucket

func (b dateBuckets) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, g := range b {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(g.date)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		ab, err := json.Marshal(g.arts)
		if err != nil {
			return nil, err
		}
		buf.Write(ab)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// bucketArticles groups articles into date buckets keyed by their local
// calendar day. ListArticles returns articles reverse chronological, so the
// first time a date is seen is its most-recent occurrence and the buckets stay
// in descending date order with each bucket's articles already descending.
func bucketArticles(arts []store.Article) dateBuckets {
	var buckets dateBuckets
	idx := map[string]int{}
	for _, a := range arts {
		t := a.PublishedAt
		if t.IsZero() {
			t = a.FetchedAt
		}
		key := timeutil.DayKey(t)
		i, ok := idx[key]
		if !ok {
			i = len(buckets)
			buckets = append(buckets, dateBucket{date: key})
			idx[key] = i
		}
		buckets[i].arts = append(buckets[i].arts, toArticleJSON(a))
	}
	return buckets
}

// runJSONDump writes all stored articles as a date-grouped JSON object to
// stdout and returns the process exit code. It reads directly from the store
// and never fetches feeds or runs the startup verification gate.
func runJSONDump(cfg *config.Config, st *store.Store) int {
	arts, err := st.ListArticles("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot list articles: %v\n", err)
		return 1
	}
	cats, err := st.ArticleCategoryMap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot load categories: %v\n", err)
		return 1
	}
	for i := range arts {
		arts[i].Categories = cats[arts[i].ID]
	}
	out, err := json.Marshal(bucketArticles(arts))
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot marshal articles: %v\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}
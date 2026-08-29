package feed

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"yerss/internal/store"
)

// FetchResult summarizes a refresh pass.
type FetchResult struct {
	Feeds   int
	New     int
	Updated int
	Errors  []error
}

// refreshPerFeedTimeout and refreshOverallTimeout bound the TUI's startup and
// manual refresh passes. They are variables so tests can shorten them.
var (
	refreshPerFeedTimeout = 15 * time.Second
	refreshOverallTimeout = 60 * time.Second
)

// maxConcurrentFetches caps the number of in-flight feed requests in a pass.
const maxConcurrentFetches = 8

// FetchFeeds downloads all feeds concurrently, persists each feed and its
// articles (deduplicated) via the store, and extracts categories. Individual
// feed failures are collected in the result rather than aborting the pass. It
// applies a per-feed HTTP timeout and an overall deadline so a hung server
// cannot stall the pass indefinitely.
func FetchFeeds(st *store.Store, urls []string) FetchResult {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{Timeout: refreshPerFeedTimeout}
	ctx, cancel := context.WithTimeout(context.Background(), refreshOverallTimeout)
	defer cancel()
	return fetchFeeds(st, urls, parser, ctx)
}

// VerifyFeeds is a bounded variant of FetchFeeds used by the startup gate. It
// fetches all feeds concurrently with a per-feed HTTP timeout and an overall
// deadline, reusing the same fetch-and-persist path. It returns an error when
// no feed parses successfully; a successful gate fetch persists the verified
// feed (stamping last_fetched_at) so the startup refresh is skipped.
func VerifyFeeds(st *store.Store, urls []string, perFeed, overall time.Duration) (FetchResult, error) {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{Timeout: perFeed}
	ctx, cancel := context.WithTimeout(context.Background(), overall)
	defer cancel()
	res := fetchFeeds(st, urls, parser, ctx)
	if res.Feeds == 0 {
		return res, fmt.Errorf("no feed parsed successfully")
	}
	return res, nil
}

// fetchFeeds downloads and persists all feeds concurrently. ctx must be
// non-nil: it carries the overall deadline, and each request is cancelled when
// it elapses. In-flight requests are capped by a semaphore so a feed set larger
// than the bound is processed in waves rather than all at once.
func fetchFeeds(st *store.Store, urls []string, parser *gofeed.Parser, ctx context.Context) FetchResult {
	var (
		mu  sync.Mutex
		res FetchResult
		wg  sync.WaitGroup
		now = time.Now()
	)
	sem := make(chan struct{}, maxConcurrentFetches)
	for _, u := range urls {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(url string) {
			defer func() {
				wg.Done()
				<-sem
			}()
			feed, err := parser.ParseURLWithContext(url, ctx)
			if err != nil {
				mu.Lock()
				res.Errors = append(res.Errors, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			res.Feeds++
			if err := st.UpsertFeed(url, feed.Title, now); err != nil {
				res.Errors = append(res.Errors, err)
			}
			mu.Unlock()

			for _, item := range feed.Items {
				art := itemToArticle(url, item)
				exists, err := st.ArticleExists(art.FeedURL, art.GUID)
				if err != nil {
					mu.Lock()
					res.Errors = append(res.Errors, err)
					mu.Unlock()
					continue
				}
				id, err := st.UpsertArticle(art)
				if err != nil {
					mu.Lock()
					res.Errors = append(res.Errors, err)
					mu.Unlock()
					continue
				}
				if err := st.SetArticleTags(id, art.Categories); err != nil {
					mu.Lock()
					res.Errors = append(res.Errors, err)
					mu.Unlock()
					continue
				}
				mu.Lock()
				if exists {
					res.Updated++
				} else {
					res.New++
				}
				mu.Unlock()
			}
		}(u)
	}
	wg.Wait()
	return res
}

func itemToArticle(feedURL string, item *gofeed.Item) store.Article {
	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}
	var pub time.Time
	if item.PublishedParsed != nil {
		pub = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		pub = *item.UpdatedParsed
	}
	author := ""
	if item.Author != nil {
		author = item.Author.Name
	} else if len(item.Authors) > 0 {
		author = item.Authors[0].Name
	}
	return store.Article{
		FeedURL:     feedURL,
		GUID:        guid,
		Title:       item.Title,
		Link:        item.Link,
		Author:      author,
		PublishedAt: pub,
		Content:     item.Content,
		Description: item.Description,
		Categories:  item.Categories,
	}
}
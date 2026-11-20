package feed

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"yerss/internal"
)

// FetchResult summarizes a refresh pass.
type FetchResult struct {
	Feeds    int
	New      int
	Updated  int
	Errors   []error
	Outcomes []FeedOutcome
}

// FeedOutcome records the per-URL result of one fetch pass: the HTTP status
// code when the fetch failed with an HTTP error, the non-HTTP error
// otherwise, and the number of articles the feed returned when it parsed
// successfully.
type FeedOutcome struct {
	URL      string
	Status   int
	Err      error
	Articles int
}

// refreshPerFeedTimeout and refreshOverallTimeout bound the TUI's startup and
// manual refresh passes. They are variables so tests can shorten them.
var (
	refreshPerFeedTimeout = 15 * time.Second
	refreshOverallTimeout = 60 * time.Second
)

// ClientFactory builds the HTTP client for a fetch pass. It is a package
// seam so tests can inject a client that trusts a self-signed test TLS
// server.
var ClientFactory = func(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// maxConcurrentFetches caps the number of in-flight feed requests in a pass.
const maxConcurrentFetches = 8

// FetchFeeds downloads all feeds concurrently, persists each feed and its
// articles (deduplicated) via the store, and extracts categories. Individual
// feed failures are collected in the result rather than aborting the pass. It
// applies a per-feed HTTP timeout and an overall deadline so a hung server
// cannot stall the pass indefinitely.
func FetchFeeds(st *internal.Store, urls []string) FetchResult {
	parser := gofeed.NewParser()
	parser.Client = ClientFactory(refreshPerFeedTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), refreshOverallTimeout)
	defer cancel()
	return fetchFeeds(st, urls, parser, ctx, nil)
}

// VerifyFeeds is a bounded variant of FetchFeeds used by the startup gate. It
// fetches all feeds concurrently with a per-feed HTTP timeout and an overall
// deadline, reusing the same fetch-and-persist path. It returns an error when
// no feed parses successfully; a successful gate fetch persists the verified
// feed (stamping last_fetched_at) so the startup refresh is skipped.
func VerifyFeeds(st *internal.Store, urls []string, perFeed, overall time.Duration) (FetchResult, error) {
	return VerifyFeedsWithProgress(st, urls, perFeed, overall, nil)
}

// VerifyFeedsWithProgress behaves like VerifyFeeds and calls progress once for
// each feed after its outcome is recorded, so callers can report progress while
// a slow feed set is verified. progress is serialized and may be nil.
func VerifyFeedsWithProgress(st *internal.Store, urls []string, perFeed, overall time.Duration, progress func(FeedOutcome)) (FetchResult, error) {
	parser := gofeed.NewParser()
	parser.Client = ClientFactory(perFeed)
	ctx, cancel := context.WithTimeout(context.Background(), overall)
	defer cancel()
	res := fetchFeeds(st, urls, parser, ctx, progress)
	if res.Feeds == 0 {
		return res, fmt.Errorf("no feed parsed successfully")
	}
	return res, nil
}

// fetchFeeds downloads and persists all feeds concurrently. ctx must be
// non-nil: it carries the overall deadline, and each request is cancelled when
// it elapses. In-flight requests are capped by a semaphore so a feed set larger
// than the bound is processed in waves rather than all at once. When progress
// is non-nil it is called once per feed with that feed's outcome.
func fetchFeeds(st *internal.Store, urls []string, parser *gofeed.Parser, ctx context.Context, progress func(FeedOutcome)) FetchResult {
	var (
		mu  sync.Mutex
		res FetchResult
		wg  sync.WaitGroup
		now = time.Now()
	)
	// record appends one outcome and reports it. Callers hold mu, so progress
	// calls never overlap and stay in completion order.
	record := func(outcome FeedOutcome) {
		res.Outcomes = append(res.Outcomes, outcome)
		if progress != nil {
			progress(outcome)
		}
	}
	sem := make(chan struct{}, maxConcurrentFetches)
	for _, entry := range urls {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(entry string) {
			defer func() {
				wg.Done()
				<-sem
			}()
			// Feeds.txt entries are scheme-less and always looked up over
			// HTTPS; the canonical URL is also the store key, so existing
			// article rows deduplicate unchanged.
			url := CanonicalURL(entry)
			feed, err := parser.ParseURLWithContext(url, ctx)
			if err != nil {
				mu.Lock()
				res.Errors = append(res.Errors, err)
				outcome := FeedOutcome{URL: url, Err: err}
				var httpErr gofeed.HTTPError
				if errors.As(err, &httpErr) {
					outcome.Status = httpErr.StatusCode
				}
				record(outcome)
				mu.Unlock()
				return
			}

			mu.Lock()
			res.Feeds++
			record(FeedOutcome{URL: url, Articles: len(feed.Items)})
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
				source := internal.SourceLabel(art.Link, art.FeedURL)
				tags := make([]string, 0, len(art.Categories)+1)
				tags = append(tags, art.Categories...)
				if source != "" {
					tags = append(tags, source)
				}
				if err := st.SetArticleTagsWithSource(id, tags, source); err != nil {
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
		}(entry)
	}
	wg.Wait()
	return res
}

func itemToArticle(feedURL string, item *gofeed.Item) internal.Article {
	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}
	link := item.Link
	if link == "" && isAbsoluteURL(guid) {
		link = guid
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
	return internal.Article{
		FeedURL:     feedURL,
		GUID:        guid,
		Title:       item.Title,
		Link:        link,
		Author:      author,
		PublishedAt: pub,
		Content:     item.Content,
		Description: item.Description,
		Categories:  item.Categories,
		ImageURL:    leadImageURL(item),
	}
}

// leadImageURL extracts the publisher-attached lead image URL for a feed item:
// the first enclosure whose MIME type is an image type, falling back to the
// media:thumbnail extension's url attribute. Inline <img> elements in the
// article HTML content are never considered — only publisher-attached
// enclosure/thumbnail metadata.
func leadImageURL(item *gofeed.Item) string {
	for _, enc := range item.Enclosures {
		if enc == nil {
			continue
		}
		if strings.HasPrefix(strings.ToLower(enc.Type), "image/") {
			if u := strings.TrimSpace(enc.URL); u != "" {
				return u
			}
		}
	}
	if exts, ok := item.Extensions["media"]; ok {
		if thumbs, ok := exts["thumbnail"]; ok {
			for _, t := range thumbs {
				if u, ok := t.Attrs["url"]; ok {
					if u = strings.TrimSpace(u); u != "" {
						return u
					}
				}
			}
		}
	}
	return ""
}

// isAbsoluteURL reports whether s parses as an absolute URL with a host. It is
// used to fall back to a feed entry's id as the article link when the entry
// carries no <link> element (common in Atom feeds where the id is the
// permalink).
func isAbsoluteURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.IsAbs() && u.Host != ""
}

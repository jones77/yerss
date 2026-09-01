// Package app is the application layer between the TUI and the store, config,
// feed, and image packages. Session owns the store, config, and image caches
// for one run, and the UI depends on it instead of reaching into those
// packages directly.
package app

import (
	"time"

	"yerss/internal/config"
	"yerss/internal/feed"
	"yerss/internal/image"
	"yerss/internal/store"
)

// Session holds the store, config, and image caches for one run. The TUI
// model accesses them through the session, keeping orchestration logic in one
// place behind a single dependency.
type Session struct {
	cfg   *config.Config
	store *store.Store

	ImgRenderer image.Renderer
	ImgCache    *image.Cache
	ImgBlocks   *image.Blocks
	ImgPhotos   *image.Photos
	ImgNatives  *image.Natives
	ImgNative   image.NativeRenderer
}

// New builds a Session bound to a config and store.
func New(cfg *config.Config, st *store.Store) *Session {
	return &Session{
		cfg:         cfg,
		store:       st,
		ImgRenderer: image.Halfblocks{},
		ImgCache:    image.NewCache(),
		ImgBlocks:   image.NewBlocks(),
		ImgPhotos:   image.NewPhotos(),
		ImgNatives:  image.NewNatives(),
		ImgNative:   image.NativeRenderer{Protocol: image.DetectProtocol()},
	}
}

// Config returns the session's config.
func (s *Session) Config() *config.Config { return s.cfg }

// Store returns the session's store.
func (s *Session) Store() *store.Store { return s.store }

// ListArticles delegates to the store.
func (s *Session) ListArticles(filterTag string) ([]store.Article, error) {
	return s.store.ListArticles(filterTag)
}

// ArticleCount delegates to the store.
func (s *Session) ArticleCount() (int, error) { return s.store.ArticleCount() }

// DBSize delegates to the store.
func (s *Session) DBSize() (int64, error) { return s.store.DBSize() }

// SetRead delegates to the store.
func (s *Session) SetRead(id int64, read bool) error { return s.store.SetRead(id, read) }

// GetArticle delegates to the store.
func (s *Session) GetArticle(id int64) (*store.Article, error) { return s.store.GetArticle(id) }

// SaveSelection persists the reader selection.
func (s *Session) SaveSelection(sel store.LastSelection) error {
	return s.store.SaveLastSelection(sel)
}

// LoadSelection restores the saved reader selection.
func (s *Session) LoadSelection() (store.LastSelection, error) {
	return s.store.LoadLastSelection()
}

// LastRefreshedAt delegates to the store.
func (s *Session) LastRefreshedAt() (time.Time, error) { return s.store.LastRefreshedAt() }

// VerifiedFeedURLs delegates to the store.
func (s *Session) VerifiedFeedURLs() (map[string]bool, error) {
	return s.store.VerifiedFeedURLs()
}

// ListTags delegates to the store.
func (s *Session) ListTags() ([]store.TagCount, error) { return s.store.ListTags() }

// FetchFeeds runs a refresh pass against the session's store and returns the
// pass summary.
func (s *Session) FetchFeeds(urls []string) *feed.FetchResult {
	res := feed.FetchFeeds(s.store, urls)
	return &res
}
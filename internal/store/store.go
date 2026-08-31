package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/charmbracelet/x/ansi"
	_ "github.com/glebarez/go-sqlite"
)

// Article is a single stored feed item.
type Article struct {
	ID          int64
	FeedURL     string
	GUID        string
	Title       string
	Link        string
	Author      string
	PublishedAt time.Time
	Content     string
	Description string
	Read        bool
	FetchedAt   time.Time
	Categories  []string
	ImageURL    string
}

// TagCount is an aggregate count for a tag. IsSource reports whether the tag is
// a source-domain tag (a news organization) rather than a feed category.
type TagCount struct {
	Name     string
	Unread   int
	Read     int
	Total    int
	IsSource bool
}

// ArticleImage is one image attached to an article, keyed by its position in
// the article. Block holds the rendered text block and Photo holds the full raw
// photo bytes, both persisted together. Width is the display width of Block in
// cells, or 0 when there is no block.
type ArticleImage struct {
	Position int
	URL      string
	Credit   string
	Block    string
	Photo    []byte
	Width    int
}

// Store wraps the SQLite connection.
type Store struct {
	db   *sql.DB
	path string
}

// Open creates (if needed) and opens the SQLite database at path, runs
// migrations, and returns a ready Store. It returns an error if the database
// cannot be created or written to.
func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create data directory %s: %w", dir, err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("cannot open database %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot open database %s: %w", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("database %s: %w", path, err)
		}
	}
	st := &Store{db: db, path: path}
	if err := st.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database %s: %w", path, err)
	}
	return st, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// DBSize returns the on-disk size in bytes of the database, including the WAL
// sidecar file when present. A missing main file is an error; a missing
// sidecar contributes zero.
func (s *Store) DBSize() (int64, error) {
	fi, err := os.Stat(s.path)
	if err != nil {
		return 0, err
	}
	total := fi.Size()
	if wal, err := os.Stat(s.path + "-wal"); err == nil {
		total += wal.Size()
	} else if !os.IsNotExist(err) {
		return 0, err
	}
	return total, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS feeds (
  url             TEXT PRIMARY KEY,
  title           TEXT,
  last_fetched_at INTEGER
);
CREATE TABLE IF NOT EXISTS articles (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  feed_url     TEXT NOT NULL REFERENCES feeds(url),
  guid         TEXT NOT NULL,
  title        TEXT,
  link         TEXT,
  author       TEXT,
  published_at INTEGER,
  content      TEXT,
  description  TEXT,
  read         BOOLEAN NOT NULL DEFAULT 0,
  fetched_at   INTEGER,
  UNIQUE(feed_url, guid)
);
CREATE TABLE IF NOT EXISTS categories (
  id     INTEGER PRIMARY KEY AUTOINCREMENT,
  name   TEXT COLLATE NOCASE UNIQUE,
  source BOOLEAN NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS article_categories (
  article_id  INTEGER REFERENCES articles(id) ON DELETE CASCADE,
  category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (article_id, category_id)
);
CREATE TABLE IF NOT EXISTS state (
  key   TEXT PRIMARY KEY,
  value TEXT
);
`)
	if err != nil {
		return err
	}
	return s.migrateImageColumns()
}

// migrateImageColumns adds the nullable image_url column to the articles table
// when it is absent, then runs the article_images migration. SQLite has no ADD
// COLUMN IF NOT EXISTS, so the guard queries PRAGMA table_info before altering.
// image_url is the retained denormalized pointer to an article's lead image;
// the legacy image_data/image_block columns are migrated into article_images and
// dropped by migrateImageTable instead of being added here.
func (s *Store) migrateImageColumns() error {
	cols, err := s.tableColumns("articles")
	if err != nil {
		return err
	}
	if !cols["image_url"] {
		if _, err := s.db.Exec(`ALTER TABLE articles ADD COLUMN image_url TEXT`); err != nil {
			return err
		}
	}
	return s.migrateImageTable()
}

// migrateImageTable creates the article_images child table and folds any legacy
// per-article image_block/image_data columns into position-0 rows, then drops
// those columns when SQLite supports column removal. It is idempotent: on an
// already-migrated database the legacy columns are absent and it is a no-op.
func (s *Store) migrateImageTable() error {
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS article_images (
  article_id INTEGER NOT NULL REFERENCES articles(id),
  position   INTEGER NOT NULL,
  url        TEXT,
  credit     TEXT,
  block      TEXT,
  photo      BLOB,
  width      INTEGER,
  PRIMARY KEY (article_id, position)
)`); err != nil {
		return err
	}
	cols, err := s.tableColumns("articles")
	if err != nil {
		return err
	}
	if !cols["image_block"] && !cols["image_data"] {
		return s.migrateCategorySource()
	}
	rows, err := s.db.Query(`
SELECT id, image_url, image_block, image_data FROM articles
WHERE (image_block IS NOT NULL AND image_block != '') OR (image_data IS NOT NULL AND image_data != '')`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id    int64
			url   sql.NullString
			block sql.NullString
			photo []byte
		)
		if err := rows.Scan(&id, &url, &block, &photo); err != nil {
			rows.Close()
			return err
		}
		width := 0
		if block.Valid && block.String != "" {
			width = blockWidth(strings.Split(block.String, "\n"))
		}
		if _, err := s.db.Exec(`
INSERT INTO article_images (article_id, position, url, credit, block, photo, width)
VALUES (?, 0, ?, '', ?, ?, ?)
ON CONFLICT(article_id, position) DO UPDATE SET
  url = excluded.url, credit = excluded.credit, block = excluded.block,
  photo = excluded.photo, width = excluded.width`,
			id, nullIfEmpty(url.String), nullIfEmpty(block.String), photo, nullIfZero(width)); err != nil {
			rows.Close()
			return err
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	// Drop the legacy columns when the SQLite build supports DROP COLUMN;
	// otherwise leave them in place, empty and unused.
	if _, err := s.db.Exec(`ALTER TABLE articles DROP COLUMN image_block`); err != nil {
		// guarded: column removal may be unsupported; leave it unused.
	}
	if _, err := s.db.Exec(`ALTER TABLE articles DROP COLUMN image_data`); err != nil {
		// guarded: column removal may be unsupported; leave it unused.
	}
	return s.migrateCategorySource()
}

// blockWidth returns the maximum cell width across the given ANSI-styled lines.
func blockWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}

// migrateCategorySource adds the source flag to the categories table when it
// is absent (pre-existing databases created before source-domain tags).
func (s *Store) migrateCategorySource() error {
	cols, err := s.tableColumns("categories")
	if err != nil {
		return err
	}
	if cols["source"] {
		return nil
	}
	_, err = s.db.Exec(`ALTER TABLE categories ADD COLUMN source BOOLEAN NOT NULL DEFAULT 0`)
	return err
}

// tableColumns returns the set of column names on the named table.
func (s *Store) tableColumns(table string) (map[string]bool, error) {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var (
			cid  int
			name string
			typ  string
			notN int
			dflt sql.NullString
			pk   int
		)
		if err := rows.Scan(&cid, &name, &typ, &notN, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}

// UpsertFeed records a feed and its last fetch time.
func (s *Store) UpsertFeed(url, title string, fetchedAt time.Time) error {
	_, err := s.db.Exec(`
INSERT INTO feeds (url, title, last_fetched_at) VALUES (?, ?, ?)
ON CONFLICT(url) DO UPDATE SET title = excluded.title, last_fetched_at = excluded.last_fetched_at`,
		url, title, fetchedAt.Unix())
	return err
}

// LastRefreshedAt returns the most recent feed fetch time across all feeds.
// It returns the zero time when no feed has ever been fetched.
func (s *Store) LastRefreshedAt() (time.Time, error) {
	var ts sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(last_fetched_at) FROM feeds`).Scan(&ts)
	if err != nil {
		return time.Time{}, err
	}
	if !ts.Valid {
		return time.Time{}, nil
	}
	return time.Unix(ts.Int64, 0), nil
}

// LastSelection records the reader's selection at quit time so a restart can
// restore the list cursor and, when applicable, the open article and its scroll
// position. View is "list" or "article"; HeaderKey is the day-group key
// ("20060102" or "undated") when the cursor sits on a day header, otherwise
// ArticleID names the selected (or open) article and ArticleOffset is its
// viewport scroll offset.
type LastSelection struct {
	View          string `json:"view"`
	ArticleID     int64  `json:"article_id,omitempty"`
	HeaderKey     string `json:"header_key,omitempty"`
	ArticleOffset int    `json:"article_offset,omitempty"`
}

// SaveLastSelection persists the reader selection under a fixed key.
func (s *Store) SaveLastSelection(sel LastSelection) error {
	data, err := json.Marshal(sel)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
INSERT INTO state (key, value) VALUES ('last_selection', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`, string(data))
	return err
}

// LoadLastSelection returns the previously saved reader selection, or the zero
// value when none has been saved.
func (s *Store) LoadLastSelection() (LastSelection, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM state WHERE key = 'last_selection'`).Scan(&v)
	if err == sql.ErrNoRows {
		return LastSelection{}, nil
	}
	if err != nil {
		return LastSelection{}, err
	}
	var sel LastSelection
	if err := json.Unmarshal([]byte(v), &sel); err != nil {
		return LastSelection{}, err
	}
	return sel, nil
}

// HasVerifiedFeed reports whether at least one feed has ever been successfully
// fetched and persisted, meaning it has a recorded fetch time (non-NULL
// last_fetched_at). Stub feed rows created by article persistence (which insert
// a URL without a fetch time) do not satisfy this check, so the startup gate
// only trusts feeds that have actually parsed.
func (s *Store) HasVerifiedFeed() (bool, error) {
	var one int
	if err := s.db.QueryRow(`SELECT 1 FROM feeds WHERE last_fetched_at IS NOT NULL LIMIT 1`).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// VerifiedFeedURLs returns the set of feed URLs that have ever been
// successfully fetched — that is, feeds rows whose last_fetched_at is set.
// Stub rows created by article persistence (NULL last_fetched_at) are excluded,
// matching HasVerifiedFeed's definition of a verified feed. The returned map is
// non-nil even when no feed qualifies.
func (s *Store) VerifiedFeedURLs() (map[string]bool, error) {
	rows, err := s.db.Query(`SELECT url FROM feeds WHERE last_fetched_at IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	urls := make(map[string]bool)
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls[u] = true
	}
	return urls, rows.Err()
}

// PruneFeeds removes every stored feed whose URL is not in keep — along with
// that feed's articles and their associated image rows — returning the
// canonical URLs of the feeds that were removed. keep should be the canonical
// feed URLs currently configured in feeds.txt. It is a no-op when keep is
// empty of the stored feeds.
func (s *Store) PruneFeeds(keep []string) ([]string, error) {
	keepSet := make(map[string]bool, len(keep))
	for _, u := range keep {
		keepSet[u] = true
	}
	rows, err := s.db.Query(`SELECT url FROM feeds`)
	if err != nil {
		return nil, err
	}
	var pruned []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			rows.Close()
			return nil, err
		}
		if !keepSet[url] {
			pruned = append(pruned, url)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(pruned) == 0 {
		return nil, nil
	}

	var b strings.Builder
	b.WriteString(`IN (`)
	args := make([]any, 0, len(pruned))
	for i, url := range pruned {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(`?`)
		args = append(args, url)
	}
	b.WriteString(`)`)
	inClause := b.String()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// article_images has no ON DELETE CASCADE, so remove those rows first.
	if _, err := tx.Exec(`
DELETE FROM article_images
WHERE article_id IN (SELECT id FROM articles WHERE feed_url `+inClause+`)`, args...); err != nil {
		return nil, err
	}
	// article_categories cascades on article delete.
	if _, err := tx.Exec(`DELETE FROM articles WHERE feed_url `+inClause, args...); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM feeds WHERE url `+inClause, args...); err != nil {
		return nil, err
	}
	return pruned, tx.Commit()
}

// UpsertArticle stores an article, deduplicating on (feed_url, guid) and
// updating content and metadata when the row already exists. It returns the
// article's ID.
func (s *Store) UpsertArticle(a Article) (int64, error) {
	if a.FetchedAt.IsZero() {
		a.FetchedAt = time.Now()
	}
	if _, err := s.db.Exec(`INSERT OR IGNORE INTO feeds (url) VALUES (?)`, a.FeedURL); err != nil {
		return 0, err
	}
	var id int64
	err := s.db.QueryRow(`
INSERT INTO articles (feed_url, guid, title, link, author, published_at, content, description, fetched_at, image_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(feed_url, guid) DO UPDATE SET
  title = excluded.title,
  link = excluded.link,
  author = excluded.author,
  published_at = excluded.published_at,
  content = excluded.content,
  description = excluded.description,
  fetched_at = excluded.fetched_at,
  image_url = excluded.image_url
RETURNING id`,
		a.FeedURL, a.GUID, a.Title, a.Link, a.Author, unixOrNull(a.PublishedAt),
		a.Content, a.Description, a.FetchedAt.Unix(), nullIfEmpty(a.ImageURL)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ArticleExists reports whether an article with the given feed URL and GUID
// is already stored.
func (s *Store) ArticleExists(feedURL, guid string) (bool, error) {
	var one int
	err := s.db.QueryRow(`SELECT 1 FROM articles WHERE feed_url = ? AND guid = ?`, feedURL, guid).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetArticleTags replaces the category associations for an article.
func (s *Store) SetArticleTags(articleID int64, names []string) error {
	return s.setArticleTags(articleID, names, "")
}

// SetArticleTagsWithSource replaces the category associations for an article
// and marks the named source-domain tag (when present among names) as a source
// tag, so news organizations can be grouped ahead of categories.
func (s *Store) SetArticleTagsWithSource(articleID int64, names []string, sourceName string) error {
	return s.setArticleTags(articleID, names, sourceName)
}

func (s *Store) setArticleTags(articleID int64, names []string, sourceName string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM article_categories WHERE article_id = ?`, articleID); err != nil {
		return err
	}

	// Tags are unique case-insensitively (the categories.name column is
	// COLLATE NOCASE) but case-preserving: the first spelling ever inserted
	// becomes the stored name forever, and later spellings resolve to the same
	// row via the NOCASE conflict target. Within one batch, only the first
	// spelling of a name is emitted, so an article never links a tag twice.
	seen := map[string]bool{}
	var catIDs []int64
	for _, name := range names {
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		source := 0
		if strings.EqualFold(name, sourceName) {
			source = 1
		}
		var catID int64
		// The NOCASE conflict target fires on any case of the same name; the
		// existing row keeps its spelling, and source is only ever promoted so
		// a tag marked as a news organization stays one.
		if err := tx.QueryRow(`INSERT INTO categories (name, source) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET source = MAX(source, excluded.source) RETURNING id`, name, source).Scan(&catID); err != nil {
			return err
		}
		catIDs = append(catIDs, catID)
	}
	if len(catIDs) > 0 {
		var b strings.Builder
		b.WriteString(`INSERT INTO article_categories (article_id, category_id) VALUES `)
		args := make([]any, 0, len(catIDs)*2)
		for i, id := range catIDs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("(?, ?)")
			args = append(args, articleID, id)
		}
		if _, err := tx.Exec(b.String(), args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SetRead persists the read flag for an article.
func (s *Store) SetRead(id int64, read bool) error {
	_, err := s.db.Exec(`UPDATE articles SET read = ? WHERE id = ?`, read, id)
	return err
}

// SetArticleImage persists a single image for an article, keyed by its
// position. Both the rendered text block and the full raw photo bytes are
// stored together in the article_images child table, replacing whatever was
// previously stored at that position.
func (s *Store) SetArticleImage(id int64, img ArticleImage) error {
	_, err := s.db.Exec(`
INSERT INTO article_images (article_id, position, url, credit, block, photo, width)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(article_id, position) DO UPDATE SET
  url = excluded.url, credit = excluded.credit, block = excluded.block,
  photo = excluded.photo, width = excluded.width`,
		id, img.Position, nullIfEmpty(img.URL), nullIfEmpty(img.Credit),
		nullIfEmpty(img.Block), img.Photo, nullIfZero(img.Width))
	return err
}

// GetArticleImages returns every stored image for an article, ordered by
// position.
func (s *Store) GetArticleImages(id int64) ([]ArticleImage, error) {
	rows, err := s.db.Query(`
SELECT position, url, credit, block, photo, width FROM article_images
WHERE article_id = ? ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	imgs := []ArticleImage{}
	for rows.Next() {
		var (
			img    ArticleImage
			url    sql.NullString
			credit sql.NullString
			block  sql.NullString
			width  sql.NullInt64
		)
		if err := rows.Scan(&img.Position, &url, &credit, &block, &img.Photo, &width); err != nil {
			return nil, err
		}
		img.URL = url.String
		img.Credit = credit.String
		img.Block = block.String
		img.Width = int(width.Int64)
		imgs = append(imgs, img)
	}
	return imgs, rows.Err()
}

// GetArticleImagePhoto returns the stored raw photo bytes for an article's
// image at the given URL, or nil when none are stored.
func (s *Store) GetArticleImagePhoto(id int64, url string) ([]byte, error) {
	var photo []byte
	err := s.db.QueryRow(`
SELECT photo FROM article_images WHERE article_id = ? AND url = ?`, id, url).Scan(&photo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return photo, nil
}

// GetArticle returns an article by ID, including its categories.
func (s *Store) GetArticle(id int64) (*Article, error) {
	a, err := s.scanArticle(s.db.QueryRow(`
SELECT id, feed_url, guid, title, link, author, published_at, content, description, read, fetched_at, image_url
FROM articles WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	names, err := s.articleCategories(id)
	if err != nil {
		return nil, err
	}
	a.Categories = names
	return a, nil
}

// ListArticles returns all articles, newest first, optionally filtered to
// those associated with a tag name.
func (s *Store) ListArticles(filterTag string) ([]Article, error) {
	query := `
SELECT a.id, a.feed_url, a.guid, a.title, a.link, a.author, a.published_at, a.content, a.description, a.read, a.fetched_at, a.image_url
FROM articles a
`
	args := []any{}
	if filterTag != "" {
		query += `
JOIN article_categories ac ON ac.article_id = a.id
JOIN categories c ON c.id = ac.category_id
WHERE c.name = ?
`
		args = append(args, filterTag)
	}
	query += ` ORDER BY COALESCE(a.published_at, 0) DESC, a.id DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		a, err := s.scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *a)
	}
	return articles, rows.Err()
}

// ArticleCount returns the total number of articles in the database,
// regardless of any tag filter.
func (s *Store) ArticleCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&n)
	return n, err
}

// SourceLabel derives the registrable organization label of a URL's host,
// used both as the list view's source identifier and as an implicit source
// tag on every stored article. It strips a leading "www." and a trailing dot,
// takes the organization label of the registrable domain (the public suffix
// plus one, so multi-part suffixes like "co.uk" are removed whole), and for
// hosts with no registrable domain (single-label hosts, IP literals) falls
// back to dropping everything from the final dot. The raw URL host is used
// when present; otherwise the feed URL host is used. It returns "" when
// neither yields a host.
func SourceLabel(rawURL, feedURL string) string {
	host := ""
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		host = u.Host
	} else if u, err := url.Parse(feedURL); err == nil && u.Host != "" {
		host = u.Host
	}
	if host == "" {
		return ""
	}
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	h = strings.TrimPrefix(h, "www.")
	if etld1, err := publicsuffix.EffectiveTLDPlusOne(h); err == nil {
		if i := strings.Index(etld1, "."); i > 0 {
			h = etld1[:i]
		} else {
			h = etld1
		}
	} else if i := strings.LastIndex(h, "."); i > 0 {
		h = h[:i]
	}
	return h
}

// ListTags returns all tags with unread/read/total counts, source-domain
// (news organization) tags first and then category tags, each group ordered by
// popularity (total article count descending) with equal-popularity tags
// ordered alphabetically (case-insensitive, so "apple" sorts before "Zoo").
func (s *Store) ListTags() ([]TagCount, error) {
	rows, err := s.db.Query(`
SELECT c.name,
       c.source,
       COUNT(ac.article_id) AS total,
       COALESCE(SUM(CASE WHEN a.read = 0 THEN 1 ELSE 0 END), 0) AS unread
FROM categories c
JOIN article_categories ac ON ac.category_id = c.id
JOIN articles a ON a.id = ac.article_id
GROUP BY c.id
ORDER BY c.source DESC, total DESC, c.name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagCount
	for rows.Next() {
		var t TagCount
		var source int
		if err := rows.Scan(&t.Name, &source, &t.Total, &t.Unread); err != nil {
			return nil, err
		}
		t.IsSource = source == 1
		t.Read = t.Total - t.Unread
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Store) articleCategories(articleID int64) ([]string, error) {
	rows, err := s.db.Query(`
SELECT c.name FROM categories c
JOIN article_categories ac ON ac.category_id = c.id
WHERE ac.article_id = ? ORDER BY c.name COLLATE NOCASE ASC`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// ArticleCategoryMap returns every article's category names keyed by article
// ID, ordered by category name within each article. It is for callers that
// need categories alongside a full article listing (ListArticles does not join
// categories).
func (s *Store) ArticleCategoryMap() (map[int64][]string, error) {
	rows, err := s.db.Query(`
SELECT ac.article_id, c.name
FROM article_categories ac
JOIN categories c ON c.id = ac.category_id
ORDER BY ac.article_id, c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64][]string)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[id] = append(m[id], name)
	}
	return m, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) scanArticle(row scanner) (*Article, error) {
	var (
		a       Article
		author  sql.NullString
		link    sql.NullString
		desc    sql.NullString
		pub     sql.NullInt64
		fetched sql.NullInt64
		imgURL  sql.NullString
	)
	err := row.Scan(&a.ID, &a.FeedURL, &a.GUID, &a.Title, &link, &author, &pub, &a.Content, &desc, &a.Read, &fetched, &imgURL)
	if err != nil {
		return nil, err
	}
	if link.Valid {
		a.Link = link.String
	}
	if author.Valid {
		a.Author = author.String
	}
	if desc.Valid {
		a.Description = desc.String
	}
	if pub.Valid {
		a.PublishedAt = time.Unix(pub.Int64, 0)
	}
	if fetched.Valid {
		a.FetchedAt = time.Unix(fetched.Int64, 0)
	}
	if imgURL.Valid {
		a.ImageURL = imgURL.String
	}
	return &a, nil
}

func unixOrNull(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Unix()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfZero(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

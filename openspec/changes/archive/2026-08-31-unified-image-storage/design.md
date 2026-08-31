## Open Questions

None. Locked: persist full photo in the database; unified per-article table;
lead image is position 0; `articles.image_url` stays as a denormalized pointer;
legacy columns are migrated then dropped (or left unused if DROP COLUMN is
unavailable).

## Options

- Unified child table (chosen) vs keeping per-article columns and adding a
  parallel table for inline images. Chosen the unified table because inline
  images (a later change) share the exact same shape; the lead image is just
  position 0.
- Persist both block and photo (chosen) vs photo only. Chosen both: the block
  gives instant offline halfblock rendering without decode/scale, the photo
  gives native re-render and re-scale at any width.

## Problems

- The fetch paths (`BlockCmd`, `PhotoCmd`) currently write only the block via
  `SetImageBlock` and discard the raw bytes. They must be changed to keep the
  raw bytes and persist both together; the photo bytes must not be re-encoded
  from the decoded image (use the bytes fetched from the network).
- A restart with a photo but a mismatched block width must re-render from the
  stored photo without a network fetch; today that path re-fetches.

## Remaining Decisions

None.

## Details

### Schema

New table (created idempotently in the existing migration chain):

```sql
CREATE TABLE IF NOT EXISTS article_images (
  article_id INTEGER NOT NULL REFERENCES articles(id),
  position   INTEGER NOT NULL,
  url        TEXT,
  credit     TEXT,
  block      TEXT,
  photo      BLOB,
  width      INTEGER,
  PRIMARY KEY (article_id, position)
);
```

`articles.image_url` remains (read by the list view and refresh). The old
`image_block` and `image_data` columns are migrated and then dropped (SQLite
3.35+ supports `ALTER TABLE ... DROP COLUMN`; the repo's `glebarez/go-sqlite`
is recent enough — if not, leave the columns empty and unused).

### Store API

```go
type ArticleImage struct {
    Position int
    URL      string
    Credit   string
    Block    string   // rendered halfblock joined by "\n"; "" when absent
    Photo    []byte   // full bytes; nil when absent
    Width    int      // BlockWidth(Block); 0 when no block
}

// SetArticleImage upserts one image row (INSERT OR REPLACE).
func (s *Store) SetArticleImage(id int64, img ArticleImage) error

// GetArticleImages returns all images for an article ordered by position.
func (s *Store) GetArticleImages(id int64) ([]ArticleImage, error)

// GetArticleImagePhoto returns the stored photo bytes for the article's image
// at the given URL, or nil when absent.
func (s *Store) GetArticleImagePhoto(id int64, url string) ([]byte, error)
```

Remove `SetImageBlock`, `GetImageBlock`, `SetImageData`, `GetImageData`. The
lead image is read/written with position 0. `GetArticleImages` is used to find
the lead image (the row with the smallest position, which is 0).

### Migration

In the migration chain (after `migrateImageColumns`), add `migrateImageTable`
that:

1. Creates `article_images`.
2. For each article with a non-empty `image_block` or `image_data`, inserts a
   position-0 row (url = image_url, block = image_block, photo = image_data,
   width = BlockWidth of the split block).
3. Drops `image_block` and `image_data` (guarded: skip the drop if it errors).
4. Is idempotent (no-op on a migrated database).

### Image package changes

- `renderBlock` no longer persists directly; instead the fetch paths call a new
  helper that persists both block and photo:

  ```go
  // persistImage writes the rendered block and the raw photo bytes together.
  func persistImage(st *store.Store, id int64, url string, block []string, photo []byte) {
      _ = st.SetArticleImage(id, store.ArticleImage{
          Position: 0, URL: url, Block: strings.Join(block, "\n"),
          Photo: photo, Width: BlockWidth(block),
      })
  }
  ```

- `BlockCmd` (network path) keeps the fetched `data` bytes and calls
  `persistImage(... , block, data)` instead of `renderBlock`'s block-only write.
  Its cache-miss path replaces `st.GetImageBlock`/`st.GetImageData` with
  `st.GetArticleImages`: serve the stored block when its width matches, else
  decode the stored photo and re-render, else fetch.

- `PhotoCmd` writes the fetched photo: after `photos.Set(url, data)`, also
  persist the photo (and the halfblock block if it renders one) via
  `persistImage`, so the photo survives restarts. Its short-circuit path checks
  `GetArticleImagePhoto` before fetching (cache hit from a prior session).

- The in-memory caches (`Blocks`, `Photos`, `Natives`) are unchanged; they are
  session-level accelerators layered on top of the now-durable store.

### Tests

- Store: migration round-trip (legacy columns → position-0 row, old columns
  dropped/empty); `SetArticleImage`/`GetArticleImages`/`GetArticleImagePhoto`
  round-trip; ordered results for multiple positions.
- Image: `BlockCmd` persists block and photo together; a width-mismatched stored
  block re-renders from the stored photo without fetching (stub the HTTP layer);
  `PhotoCmd` persists the photo so a fresh session's `GetArticleImagePhoto`
  returns it without a fetch.

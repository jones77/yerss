## Why

Opening an image-heavy article (for example a long ProPublica feature) freezes
the UI for multiple seconds, and backing out of it can take ~8 seconds. Every
image load that lands re-composes the entire article synchronously on the UI
goroutine — re-running HTML→markdown conversion, a full glamour render of the
body, and link harvesting for every block message — so a photo-dense article
triggers dozens of full re-renders while its images load, and the Back keypress
queues behind them.

## What Changes

- Cache the article's derived display content (converted markdown body, inline
  image list, harvested links, and per-segment glamour render output) keyed by
  article id and content geometry, so a recompose after an image load reuses
  the expensive derivation instead of re-running it.
- Recompose on each image-load message using the cached derivation: only the
  changed image block is re-rendered and re-joined, and the reading position is
  preserved as today. Each image-load message recomposes the article at most
  once, after its cache work, so no image message re-runs HTML→markdown
  conversion or the glamour body render on the UI thread.
- Slim the list-load query (`ListArticles`) to the columns the list view
  renders, dropping the full HTML `content` column from the select, so
  returning to the list view stops copying every article body.
- Preserve all existing rendering, scrolling, snap, and read-state behavior —
  this is purely a performance change.

## Capabilities

### New Capabilities
- `article-recomposition`: defines how the article view re-composes when image
  loads land — derived article content cached per article and geometry, one
  recompose per update batch, and no full derivation re-run on the UI thread.

### Modified Capabilities
<!-- none -->

## Impact

- `internal/ui/article.go` — `newArticleState` / article state construction;
  add a derived-content cache layer.
- `internal/ui/model.go` — image message handling (`onBlockLoaded`,
  `onPhotoLoaded`, `onNativeLoaded`) to coalesce and reuse cached derivation.
- `internal/ui/image.go` — recompose path and image-source checks.
- `internal/store/store.go` — new `ListArticlesLite` column slimming; the UI
  list load stops copying article bodies while the JSON export keeps the full
  `ListArticles`.
- `internal/app/session.go` — no change expected.
- Tests in `internal/ui/*` asserting recompose behavior and list load.
## Why

The database currently persists only the rendered halfblock text block; the full
photo bytes are held in memory for the session and deliberately thrown away. That
means a native-capable terminal re-fetches a photo on every open after a restart,
and offline reading shows only the low-fidelity halfblock art. The database is
the natural home for the full photo: it makes "reopen this article from 7 months
ago" show the real image with no network, and the app already surfaces the
database size in the status bar so users can see — and `--init-db` — the cost.
Disk growth to a few hundred MB is acceptable for a pre-1.0 single-user tool.

## What Changes

- Replace the three per-article image columns (`image_url`, `image_block`,
  `image_data`) with a unified `article_images` child table keyed by
  `(article_id, position)` that stores, for each image: URL, credit, rendered
  halfblock block, full photo bytes, and the block's render width. The lead
  image is position 0; `articles.image_url` stays as a denormalized pointer for
  the list view and refresh.
- Persist the full photo bytes alongside the rendered block on every fetch, so
  later opens re-render natively (or re-render the halfblock at a new width)
  from the stored bytes without a network request.
- Migrate any legacy `image_block`/`image_data` rows into the new table on
  startup, then stop reading the old columns.
- Reverse the "raw photo bytes SHALL NOT be persisted" decision in the three
  specs that encode it.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `image-block-storage`: persists the lead image in a unified `article_images`
  table storing both the rendered block and the full photo, and migrates legacy
  columns.
- `native-image-rendering`: the fetched photo is persisted (not memory-only), so
  a later open re-renders natively from stored bytes.
- `reader-ui`: the async image load lifecycle persists the full photo bytes
  alongside the block (folds in the pending async-native-image-render "no byte
  limit" change).

## Impact

- `internal/store/store.go`: new `ArticleImage` type, `article_images` table +
  migration, and `SetArticleImage` / `GetArticleImages` / `GetArticleImagePhoto`;
  retire `SetImageBlock` / `GetImageBlock` / `SetImageData` / `GetImageData`.
- `internal/image/image.go`: `BlockCmd`/`PhotoCmd` persist both block and photo
  via the unified API; their cache-miss paths read the stored block and photo
  from `GetArticleImages` instead of the removed single-image getters.
- `internal/store/store_test.go` and `internal/image/image_test.go`: reworked
  for the new API and the block+photo round-trip.
- No new external dependencies.

**Sequencing note (required):** this change MODIFIEDs "Async image load
lifecycle" (reader-ui) and "Native photo render with block placeholder"
(native-image-rendering), which the `async-native-image-render` change also
MODIFIEDs and which is not yet archived. **Archive `async-native-image-render`
before this change** so this change's deltas apply against the post-archive spec
tree. This change's `reader-ui` delta already carries the no-byte-limit wording,
so it must not be applied before `async-native-image-render`'s delta.

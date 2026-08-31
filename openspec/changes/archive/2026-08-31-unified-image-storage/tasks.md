## 1. Store schema and API

- [x] 1.1 Add `type ArticleImage struct { Position int; URL, Credit, Block string; Photo []byte; Width int }` to `internal/store/store.go`
- [x] 1.2 Add an idempotent `migrateImageTable` (create `article_images`, copy legacy `image_block`/`image_data` into position-0 rows with url=`image_url` and width=`BlockWidth`, drop the old columns if `ALTER TABLE ... DROP COLUMN` succeeds) and call it from the migration chain
- [x] 1.3 Add `SetArticleImage(id, ArticleImage)` (INSERT OR REPLACE), `GetArticleImages(id) ([]ArticleImage, error)` (ordered by position), and `GetArticleImagePhoto(id, url) ([]byte, error)`
- [x] 1.4 Remove `SetImageBlock`, `GetImageBlock`, `SetImageData`, `GetImageData` and update every caller/tests to the unified API
- [x] 1.5 Store tests: migration round-trip (legacy columns → position-0 row, old columns dropped/empty); `SetArticleImage`/`GetArticleImages`/`GetArticleImagePhoto` round-trip; multiple positions ordered

## 2. Image package: persist both block and photo

- [x] 2.1 Add a `persistImage(st, id, url, block []string, photo []byte)` helper that writes a position-0 `ArticleImage` (block joined by "\n", width = `BlockWidth(block)`, photo bytes)
- [x] 2.2 In `BlockCmd`, keep the fetched `data` bytes and call `persistImage(..., block, data)` instead of the block-only `SetImageBlock` write; replace the cache-miss `GetImageBlock`/`GetImageData` reads with `GetArticleImages` (serve stored block when width matches; decode stored photo and re-render on width mismatch; else fetch)
- [x] 2.3 In `PhotoCmd`, persist the fetched photo via `persistImage` (alongside the halfblock block when one is rendered) so it survives restarts; short-circuit on `GetArticleImagePhoto` before fetching
- [x] 2.4 Keep the in-memory `Blocks`/`Photos`/`Natives` caches as session accelerators (unchanged)
- [x] 2.5 Image tests: `BlockCmd` persists block and photo together; width-mismatched stored block re-renders from the stored photo with no fetch (stub HTTP); `PhotoCmd` persists the photo so `GetArticleImagePhoto` returns it after a simulated restart

## 3. Verification

- [x] 3.1 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [x] 3.2 Run `openspec validate unified-image-storage`

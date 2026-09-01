## 1. Store block-only upsert

- [x] 1.1 Add `Store.SetArticleImageBlock` that upserts `url`, `block`, `width` and leaves `credit`/`photo` untouched.
- [x] 1.2 Add a store test: a block-only write after a full write preserves the original photo bytes and updates the block/width.

## 2. Route the re-render through the block-only write

- [x] 2.1 In `BlockCmd`, change the stored-photo re-render branch (`internal/image/image.go`) to call `SetArticleImageBlock` instead of `persistImage`.
- [x] 2.2 Keep `persistImage` for the network-fetch branches in `BlockCmd` and `PhotoCmd`.

## 3. Tests

- [x] 3.1 Add an image-level test: re-rendering at a new width from a stored photo updates the block without rewriting the photo.
- [x] 3.2 Add an image-level test: a fresh fetch still writes block and photo together.

## 4. Validation

- [x] 4.1 Run `go test ./...` and `go vet ./...` and fix regressions.

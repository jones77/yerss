## 1. MarkAllRead batching

- [x] 1.1 Rewrite `MarkAllRead` to run `UPDATE articles SET read = 1 WHERE id IN (...)` in chunks (e.g. 500) within a single transaction
- [x] 1.2 Add a `chunkIDs` (or equivalent) helper
- [x] 1.3 Test: marking all across a list larger than one chunk updates every id exactly once

## 2. UpsertArticle RETURNING

- [x] 2.1 Change `UpsertArticle` to append `RETURNING id` and scan the id directly, removing the follow-up `SELECT id`
- [x] 2.2 Test: fresh insert and conflict-update both return the correct id; dedup on (feed_url, guid) unchanged

## 3. SetArticleTags batching

- [x] 3.1 Rewrite the category upsert to use `INSERT ... ON CONFLICT(name) DO UPDATE SET name = name RETURNING id`
- [x] 3.2 Insert the `(article_id, category_id)` join rows in one batched pass
- [x] 3.3 Test: tag replacement, duplicate names, and empty names behave as before

## 4. Validation

- [x] 4.1 Run `go test ./...` and `go vet ./...`; fix any failures
- [x] 4.2 `openspec validate --changes batch-store-writes --strict` passes
- [x] 4.3 Manual smoke: refresh a feed and mark all read; confirm behavior is unchanged and the DB still opens cleanly

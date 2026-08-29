# Tasks: refresh-new-feeds

## 1. Store support

- [x] 1.1 Add `Store.VerifiedFeedURLs() (map[string]bool, error)` in `internal/store/store.go` running `SELECT url FROM feeds WHERE last_fetched_at IS NOT NULL`
- [x] 1.2 Add a store test: verified rows are returned, stub rows (NULL `last_fetched_at`) are excluded, empty table returns an empty map

## 2. Startup refresh trigger

- [x] 2.1 In `internal/ui/model.go` `Init`, load configured URLs via `feed.LoadFeeds(m.cfg.FeedsFile())`; on error, fall through to the existing time gate
- [x] 2.2 Compute the unfetched count using `Store.VerifiedFeedURLs`; when it is non-zero, return `m.refreshCmd()` before the `NeedsStartupRefresh` check
- [x] 2.3 Keep the existing 15-minute time-gated path for the all-verified case unchanged

## 3. Tests

- [x] 3.1 Add a model test: a configured URL with no verified row forces a refresh command even when `lastRefreshedAt` is within 15 minutes
- [x] 3.2 Add a model test: all configured URLs verified and last refresh recent → no refresh command returned
- [x] 3.3 Confirm existing startup-gate and refresh tests still pass

## 4. Verification

- [x] 4.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [x] 4.2 Manually: with a verified DB, add a new URL to feeds.txt and launch yerss; confirm the new feed is fetched on startup without pressing R
  (replaced by the automated end-to-end test `TestInitEndToEndFetchesNewFeed`, which executes the startup refresh against a local RSS server and asserts the new feed's article is stored)

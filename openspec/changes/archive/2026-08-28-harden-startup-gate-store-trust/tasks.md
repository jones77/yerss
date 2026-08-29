## 1. Store trust tightening

- [x] 1.1 Change `HasVerifiedFeed` in `internal/store/store.go` to `SELECT 1 FROM feeds WHERE last_fetched_at IS NOT NULL LIMIT 1`, returning true iff a row is found; update its doc comment to state the non-NULL-`last_fetched_at` requirement

## 2. Gate read-error distinction

- [x] 2.1 In `cmd/yerss/gate.go`, classify the `LoadFeeds` error: import `errors`, and after the call split into `errors.Is(err, os.ErrNotExist)` (missing → exit 2 existing message) vs other error (print `cannot read feeds file <path>: <err>`, exit 1) vs no error + empty (exit 2)
- [x] 2.2 Keep the empty-file and missing-file paths both on exit 2 with the existing "no RSS feeds configured" message

## 3. Open-article bounds safety

- [x] 3.1 In `internal/ui/article.go`, replace the `len == 0` early return in `openArticle` with `if m.list.cursor < 0 || m.list.cursor >= len(m.list.articles) { return }` before indexing

## 4. Tests

- [x] 4.1 T5: add a store test that calls `UpsertArticle` only (no `UpsertFeed`) and asserts `HasVerifiedFeed()` is false; add a companion assertion that after `UpsertFeed` it is true
- [x] 4.2 T8: add a gate test with a feeds file made unreadable (chmod 000, skip on root) asserting the read-error message is printed and the exit code is distinct from `exitEmpty` (expect 1, not 2); keep the existing empty/missing tests on exit 2
- [x] 4.3 T10: add a UI test that sets `m.list.cursor` beyond `len(m.list.articles)` (and to -1) and asserts `openArticle` does not panic and does not change `m.view`

## 5. Verification

- [x] 5.1 Run `go build ./...`, `go vet ./...`, `go test ./... -race`; resolve any fallout
- [x] 5.2 Confirm the existing `cmd/yerss` gate tests still pass and the offline-trust scenario still exits 0 when a real verified feed exists

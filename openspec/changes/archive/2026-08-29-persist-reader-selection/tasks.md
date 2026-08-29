## 1. Store persistence

- [x] 1.1 Add a `state (key TEXT PRIMARY KEY, value TEXT)` table to the migration in `internal/store/store.go`
- [x] 1.2 Add the `LastSelection` struct (`View`, `ArticleID`, `HeaderKey`, `ArticleOffset`) with `SaveLastSelection`/`LoadLastSelection` (JSON blob under `last_selection`)
- [x] 1.3 Test: save/load roundtrip, overwrite, and zero value on an empty store

## 2. Persist selection on quit

- [x] 2.1 Add `persistSelection` (`internal/ui/selection.go`): records view, the article ID under the cursor (or the day-group key on a header), and the article viewport `YOffset` when in the article view
- [x] 2.2 Call `persistSelection` in the `config.Quit` branches of `updateList`, `updateArticle`, and `updatePopup`

## 3. Restore selection on start

- [x] 3.1 Add `restoreSelection`: repositions the cursor to the saved article row (by ID) or day-header row (by day key), and reopens the saved article with `SetYOffset` when the saved view is the article view
- [x] 3.2 Call `restoreSelection` from `Init` after `loadList`
- [x] 3.3 Fall back to the clamped cursor (no error) when the saved article or day group no longer exists
- [x] 3.4 Test: cursor restored for an article row and a day header; article view reopened with scroll position; `b` from the restored article lands on the selected article; fallback when the article is gone

## 4. Validation

- [x] 4.1 Run `go test ./...` and linter; fix any failures
- [x] 4.2 `openspec validate --changes persist-reader-selection --strict` passes
- [ ] 4.3 Manual smoke test: select an article, scroll the reader, quit and relaunch; confirm the list cursor, article, and scroll position are restored, and that `b` returns to the selected article
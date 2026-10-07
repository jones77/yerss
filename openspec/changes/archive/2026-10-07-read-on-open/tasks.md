## 1. Mark read on open when the article fits

- [x] 1.1 In `composeArticle` (`internal/ui/articleview/article.go`), mark the
  article read when `len(st.lines) <= vpH` (restore the fit guard; do not mark
  unconditionally)

## 2. Mark read on any movement

- [x] 2.1 In `UpdateKey` (`internal/ui/articleview/update.go`), call
  `s.MarkRead(s.sess.Store())` for every scroll/jump case — `MoveUp`,
  `MoveDown`, `PageUp`, `PageDown`, `HalfPageUp`, `HalfPageDown`, `Top`,
  `Bottom`
- [x] 2.2 In `UpdateMouse`, call `s.MarkRead(s.sess.Store())` for both wheel up
  and wheel down
- [x] 2.3 Confirm `restoreSelection` (`internal/ui/selection.go`) and
  `Recompose`/`RecomposeInPlace` do not mark read (restore is not movement)

## 3. Tests

- [x] 3.1 Keep a short-article test asserting a fitting article is marked read
  on open (`internal/ui/readstatus_test.go`)
- [x] 3.2 Restore a long-article test asserting a non-fitting article is NOT
  marked read on open
- [x] 3.3 Add a test asserting the first movement in each direction (up, down,
  page up/down, half-page up/down, top, bottom) marks a long article read
- [x] 3.4 Add a test asserting a movement marks read even when the position does
  not change
- [x] 3.5 Add a test asserting further movement leaves an already-read article
  read

## 4. Docs

- [x] 4.1 Add an AGENTS.md TODO noting light mode is not fully implemented;
  assume dark mode until a light-mode proposal lands

## 5. Verification

- [x] 5.1 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm
  all pass

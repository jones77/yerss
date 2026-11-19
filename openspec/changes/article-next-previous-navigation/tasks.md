# Tasks: article-next-previous-navigation

## 1. Keybindings

- [x] 1.1 Add `previous_article` (bound to `J`) and `next_article` (bound to
  `K`) actions to the article view in `internal/config/keys.go`, following the
  existing action/`EffectiveKeys` structure. Uppercase keys must remain
  distinct from the lowercase `j`/`k` movement keys.
- [x] 1.2 Add the two actions to the article-view keybinding help (they are
  derived from the catalog, so verify they appear under the Article column).

## 2. List view: selection by ID and filtered total

- [x] 2.1 Add `SelectArticleID(id int64) bool` to `internal/ui/listview`
  (`access.go`): locate the article's `(groupIdx, artIdx)`, expand its day
  group, set the raw cursor to the article's visible row, and return whether
  it was found.
- [x] 2.2 Change `selectedArticlePosition()` in `internal/ui/listview/list.go`
  to return `n, s.articleCount()` (the filtered count) instead of `n, s.total`.
- [x] 2.3 Remove the now-unused `total` field on `State`, its assignment in
  `Load()`, and the `sess.ArticleCount()` call there if nothing else uses it.

## 3. Article view: position and hint state

- [x] 3.1 Add `listPos`, `listTotal` fields plus `SetListPosition(n, total)`
  and `ListPosition() (n, total int)` to `internal/ui/articleview/state.go`.
- [x] 3.2 Add unexported `navHint` (none/previous/next) and `navHintExpires`
  fields, plus `NavHint() string`, `ResetNavHint()`, and `ExpireNavHint()`.
- [x] 3.3 Add the exported `NavHintExpiredMsg` message type.

## 4. Article view: scroll hint arming

- [x] 4.1 In `internal/ui/articleview` scroll handling, after each viewport
  move: set `navHint` to `next` at the top edge, `previous` at the bottom edge
  (only for a scrollable article), or clear it mid-scroll; arming returns a
  `tea.Tick(2*time.Second, ...)` command carrying `NavHintExpiredMsg`.

## 5. Article view: render chrome

- [x] 5.1 In `internal/ui/articleview/article.go` `RenderUncached`, compute the
  counter label from `ListPosition()` and the hint label from `NavHint()`, and
  pass both to `render.RenderArticleBorder`.
- [x] 5.2 Update `render.RenderArticleBorder` / `TopBorder` /
  `BottomBorder` in `internal/ui/render/border.go` to accept the right-aligned
  counter text and the dynamic left hint (replacing the hard-coded `o:` hint).

## 6. Model: navigation wiring

- [x] 6.1 Extract `openArticleByID(id int64) tea.Cmd` from `openArticle` in
  `internal/ui/model_article.go`; keep `openArticle` resolving the cursor ID
  and calling it.
- [x] 6.2 Add `setArticlePosition(id int64)` and `navigateArticle(step int)`
  (step `+1` = previous/older/down, `-1` = next/newer/up) to the model, using
  `m.list.ArticleIDs()` with wrap; no-op when the article is not in the list.
- [x] 6.3 Route `articleview.ActionPreviousArticle` / `ActionNextArticle` in
  `updateArticle` to `navigateArticle`.
- [x] 6.4 In `backToList`, after `Load()`, call `m.list.SelectArticleID(m.article.ID())`.
- [x] 6.5 Handle `articleview.NavHintExpiredMsg` in the model's `Update` by
  calling `m.article.ExpireNavHint()`.

## 7. Help popup dismiss

- [x] 7.1 Add `b` and `h` to the `KindHelp` close case in
  `internal/ui/popup/popup.go` (alongside `esc`/`enter`/` `).

## 8. Tests

- [x] 8.1 List view: assert `SelectArticleID` selects a collapsed group's
  article and reports not-found for an unknown ID; assert the status bar
  `total` is the filtered count under a tag filter.
- [x] 8.2 Article view: assert `J`/`K` open the neighbor, wrap at both ends,
  stay within a filter, and no-op when the article is not in the list; assert
  the counter label and the dynamic `J: previous` / `K: next` hint (and its
  `o:` revert) appear in the borders.
- [x] 8.3 Popup: assert `b` and `h` dismiss the help popup.

## 9. Verification

- [x] 9.1 `go vet ./...`
- [x] 9.2 `go build ./...`
- [x] 9.3 `go test ./...`

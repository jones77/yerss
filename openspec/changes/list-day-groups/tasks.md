# Tasks: list-day-groups

## 1. Store

- [ ] 1.1 Add `Store.ArticleCount() (int, error)` returning `SELECT COUNT(*) FROM articles`
- [ ] 1.2 Add a store test: count matches `len(ListArticles(""))` across empty and populated DBs

## 2. List data model

- [ ] 2.1 In `internal/ui/list.go`, add `PublishedAt time.Time` to `articleItem`; populate it in `loadList` from `ListArticles`
- [ ] 2.2 Add `dayGroup` (date, label, collapsed, articles) and replace `listState.articles` with `listState.groups []dayGroup`
- [ ] 2.3 In `loadList`, bucket the reverse-chronological result into day groups by `PublishedAt.Local()` midnight; undated articles go to a trailing `Undated` group
- [ ] 2.4 Cache the unfiltered `Store.ArticleCount` total on the model for the status bar

## 3. Date label + timestamp

- [ ] 3.1 Add `longDate(t time.Time) string` producing `Weekday Do-ordinal Month, Year` with correct English ordinals (1st/2nd/3rd/11th–13th exception/21st)
- [ ] 3.2 Add tests for `longDate` ordinal edge cases

## 4. Visible-row flattening + cursor

- [ ] 4.1 Add `visibleRow{kind, groupIdx, artIdx}` and `visibleRows()` building the flattened header/article list honoring `collapsed`
- [ ] 4.2 Move `listState.cursor` to index `visibleRows()`; rewrite `moveListCursor` to walk it with wrap
- [ ] 4.3 Rewrite `listWindow` (or its call site) to window over `len(visibleRows())`

## 5. Rendering

- [ ] 5.1 In `renderList`, remove the `yerss` header block entirely; the first row is the most recent day header (or empty-state message)
- [ ] 5.2 Render day headers with a fold glyph (expanded `▾` / collapsed `▸`) plus the `longDate` label
- [ ] 5.3 Render article rows as `cursor + title (left, truncated) + HH:MM:SS (right, local)`; keep unread-bold / read-dim styling and the cursor background
- [ ] 5.4 Update the status bar to `<n>/<total> articles` using the cached total; keep the refresh date and active-filter name

## 6. Fold key dispatch

- [ ] 6.1 In `updateList`, before the action-map switch, resolve the cursor's `visibleRow`; on a header row handle `h` (collapse), `l` (expand), `enter`/`space`/`tab` (toggle) and return without falling through
- [ ] 6.2 On an article row, handle `tab` as toggle of the article's day group and return; otherwise fall through to the existing action-map dispatch
- [ ] 6.3 Add `collapse`/`expand`/`toggle(groupIdx)` helpers on the model that mutate `groups[i].collapsed`

## 7. Tests

- [ ] 7.1 Grouping: three days → three desc groups; undated → trailing `Undated` group
- [ ] 7.2 `visibleRows`: collapsed group yields only its header; expanded yields header + articles
- [ ] 7.3 Cursor wrap across visible rows; `l`/Enter on a header do not open an article
- [ ] 7.4 Fold keys: `h` collapses, `l` expands, `enter`/`space`/`tab` toggle on a header; `tab` toggles from an article row
- [ ] 7.5 Status bar shows `n/total`; with a filter, `n` < `total` and the filter name appears
- [ ] 7.6 Article row shows `HH:MM:SS` in local time on the right

## 8. Verification

- [ ] 8.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [ ] 8.2 Visually confirm grouping, fold glyphs, the time column, `n/total`, and that collapsing a day hides its articles and that `h`/`l`/Enter/Space/Tab behave per the spec on headers vs articles

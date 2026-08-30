## 1. Keymap changes (config)

- [x] 1.1 Add `MoveLeft` / `MoveRight` catalog actions valid only in `ViewPopup`, bound to `left`/`h` and `right`/`l` in `internal/config/keys.go`
- [x] 1.2 Add `ViewPopup` to the `Top` / `Bottom` catalog actions' views in `internal/config/keys.go`
- [x] 1.3 Make `EffectiveKeys` skip `h` for `Back` in `ViewPopup` so the popup view resolves `h` → `MoveLeft`, `l` → `MoveRight`, `esc`/`b` → `Back`
- [x] 1.4 Update `internal/config/config_test.go` for the catalog change (Global action list gains Move Left / Move Right; popup view resolves left/h and right/l, and 1/g/G/ctrl+up/ctrl+down)

## 2. Grid layout and rendering (ui)

- [x] 2.1 Rework the grid-geometry computation in `internal/ui/popup.go`: rows from `h-4` margins with a min-size clamp, columns from tag count, each column sized to its own widest entry (no truncation)
- [x] 2.2 Rewrite `renderTagPopup` to render the column-major grid without abbreviating entries, showing multiple columns side by side when they fit and a horizontal window derived from the cursor when the grid is wider than the interior; verify the article-view popup uses the same geometry and scrolling
- [x] 2.3 Update `internal/ui/tags_test.go` / `render_test.go` for the new geometry and layout (multi-column fill order, 23-column cap, multiple columns visible with mixed widths, horizontal window following the cursor, tiny-terminal clamp)
- [x] 2.4 Add `elideMiddle` to `internal/ui/width.go` and apply the 23-column cap with middle elision to tag popup cells (gutter and counts kept intact, `…`/`...` ellipsis)
- [x] 2.5 Update the news-org column: `sourceID` delegates to `store.SourceLabel` (full registrable label) and the list row middle-elides it to 15 columns

## 3. Grid navigation

- [x] 3.1 Add grid movement in `internal/ui/popup.go`: vertical via `wrapIndex`, horizontal via col/row with clamping into a partial final column, and first/last jumps (viewport-free)
- [x] 3.2 Dispatch `MoveLeft` / `MoveRight` / `Top` / `Bottom` in `updatePopup` for tag popups only; keep them no-ops in the links popup so it stays single-column and linear
- [x] 3.3 Add navigation tests: down wraps last→first and continues into the next column, up wraps first→last, left/right and `h`/`l` wrap at grid edges, horizontal clamp into a partial column, `1`/`g`/`G` jumps, and the scroll window advances with the cursor
- [x] 3.4 Add links-popup regression tests confirming left/right/`h`/`l`/top/bottom are no-ops and up/down/enter behave unchanged

## 4. Source-domain tags

- [x] 4.1 Add `store.SourceLabel` (registrable organization label derivation, moved from `ui.sourceID`) with unit tests
- [x] 4.2 Add a `source` flag to the `categories` table (schema + idempotent migration) and mark source tags via `SetArticleTagsWithSource`
- [x] 4.3 Append the source-domain tag to each article at ingest in `internal/feed/fetch.go`
- [x] 4.4 Order source tags ahead of category tags in `ListTags` (each group by popularity) and add store tests

## 5. Tag popup presentation

- [x] 5.1 Inline the popup title in the top border (`inlineTitleBorder`) instead of a body line; drop the top/bottom box padding so the grid fills the interior between the borders
- [x] 5.2 Replace the `>` cursor marker with the list view's full-cell background highlight, applied per segment so it spans the whole column width (including padding) for every item in the column and changes width when the cursor moves between columns
- [x] 5.3 Render the count total in the text (white) role (list-view read color) instead of the dim role
- [x] 5.4 Drop the two-space cursor gutter so the tag sits one column from the content edge; cells padded to the column width joined with two spaces between columns
- [x] 5.5 Add UI tests: selection highlight, cell layout (edge spacing, single-space column gap, highlight width), inline title, count colors, source-first grouping
- [x] 5.6 Add a right-aligned `<percent>% · <nth>/<total>` position indicator to the tag popup's bottom border, mirroring the article view and list view status bar
- [x] 5.7 Render tag counts as `X/Y` without parentheses (all-read/all-unread collapse to a single total)

## 6. Verification

- [x] 6.1 Run `go vet ./...` and `go build ./...`, then `go test ./...` and confirm all pass

## 7. Sources/tags section subtitles

- [x] 7.1 Add non-selectable `Sources` and `Tags` section subtitles to the tag popup, laid out as in-column cells in the existing column-major grid (each shown only when its group is non-empty), rendered in the non-bold status-bar blue role and never carrying the selection cursor
- [x] 7.2 Rework the tag grid to render the display item list (subtitles + tags) while keeping tag-index cursor semantics: vertical navigation stays `wrapIndex` on the tag count, and horizontal movement maps through the item grid, skipping subtitle cells and clamping into a shorter final column
- [x] 7.3 Change the tag popup title on both the list and article views to `Tags & Sources`
- [x] 7.4 Update `internal/ui/tags_test.go` for the subtitle rows (horizontal clamp position, grid-start subtitle) and add subtitle grouping, non-selectability, blue-role color, and title tests
- [x] 7.5 Update the reader-ui delta spec for the subtitles, non-selectable headers, and the new title
- [x] 7.6 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass

## 8. Column fill and right-aligned counts

- [x] 8.1 Lower the tag popup column cap from 33 to 23 columns
- [x] 8.2 Right-align the unread/total counts in each tag cell, with the name left-aligned and leftover column space padding between them
- [x] 8.3 When the natural grid is narrower than the popup interior, widen the columns — to one equal width that fills the interior, or by equal shares when a long entry forces a wider column — and pull the right border in by any leftover that cannot be split into whole columns, so the grid never leaves space to spare; `tagGrid` returns the effective box width
- [x] 8.4 Update `internal/ui/tags_test.go` for the 23-column cap, the fill geometry, and add `TestTagPopupColumnsFillInterior` and `TestTagPopupCountsRightAligned`
- [x] 8.5 Update the reader-ui delta spec and design for the cap change, right-aligned counts, and column fill
- [x] 8.6 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass

## 9. Two-space column gap

- [x] 9.1 Increase the spacing between tag popup columns from one space to two via a `tagColGap` constant, used consistently by the fill geometry (`fillColumns`), the horizontal window prefix-sum, and the row renderer
- [x] 9.2 Update `internal/ui/tags_test.go` for the two-space gap (cell layout separation, fill geometry expectations)
- [x] 9.3 Update the reader-ui delta spec and design for the two-space column gap
- [x] 9.4 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass

## 10. Symmetric horizontal scrolling

- [x] 10.1 Anchor the tag popup's horizontal window on a stored `scrollCol` in `popupState` so it advances right only when the cursor passes the right edge and left only when it passes the left edge, staying put otherwise (replacing the eager left scroll of the cursor-derived window); wraps, jumps, and resizes still snap the window to show the cursor
- [x] 10.2 Add `TestTagPopupWindowSymmetricScroll` confirming the window stays put while moving left within it and only scrolls when the cursor reaches the first column
- [x] 10.3 Update the reader-ui delta spec and design for the symmetric window behavior
- [x] 10.4 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass

## 11. Alphabetical tiebreak for tags

- [x] 11.1 Order equal-popularity tags alphabetically (case-insensitive) in `ListTags` via `COLLATE NOCASE`, keeping popularity as the primary sort and source-first grouping; apply the same collation to the article categories query
- [x] 11.2 Add `TestListTagsAlphabeticalSecondarySort` and update the reader-ui delta spec / design
- [x] 11.3 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass

## 12. Case-insensitive case-preserving tags

- [x] 12.1 Make the `categories.name` column `COLLATE NOCASE UNIQUE` so tag names are unique case-insensitively while preserving the first spelling seen forever; the `ON CONFLICT(name)` insert resolves later spellings to the same row
- [x] 12.2 Deduplicate a batch of tag names case-insensitively and match the source label case-insensitively in `setArticleTags`; keep a tag marked as a source once set (`MAX(source, excluded.source)`)
- [x] 12.3 Add store tests for case-insensitive, case-preserving tags, first-seen spelling, and case-insensitive filtering/source flags
- [x] 12.4 Update the feed-pipeline delta spec for case-insensitive tag storage and matching

## 13. Database reinit flag

- [x] 13.1 Add a `-z` / `--init-db` flag to `cmd/yerss` that deletes the database file and WAL/SHM sidecars before opening, prints a notice, and lets a fresh store start from zero (no dedup migration needed for existing databases)
- [x] 13.2 Add `resetDatabase` and unit tests; verify the flag end-to-end with `-z -j`
- [x] 13.3 Add a configuration delta spec for the reinit flag and update the design
- [x] 13.4 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm all pass
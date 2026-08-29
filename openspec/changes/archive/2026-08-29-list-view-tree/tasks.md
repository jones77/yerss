## 1. Glyphs

- [x] 1.1 Add the `tee` field to `borderGlyphs` (`├` Unicode, `+` ASCII) in `glyphsFor`, keeping the Unicode/ASCII pairs in sync per AGENTS.md
- [x] 1.2 Extend the ASCII glyph test to cover the tee glyph

## 2. Day labels

- [x] 2.1 Add a `now time.Time` parameter to `bucketDayGroups` and prefix `today, ` / `yesterday, ` to `longDate` labels when the group's local midnight matches the reference day or the previous day; update call sites to pass `time.Now()`
- [x] 2.2 Update `daygroups_test` to pass a fixed reference time; add tests for the today/yesterday prefixes, unprefixed older days, and the unchanged `Undated` label

## 3. Tree rail

- [x] 3.1 Render `┌`/`├`/`└` prefixes per row from the whole visible-row list (first row, interior, last), with `─` after the corner on article rows and a space on header rows; remove fold markers from day headers
- [x] 3.2 Add tests: bookend glyphs on the full list, interior tees, glyphs stable across a scrolled window, collapsed-group recomputation, ASCII fallback (`+`, `-`)

## 4. Row layout and selection

- [x] 4.1 Rewrite `renderArticleRow`: rail prefix, `HH:MM` (or `--:--`) after the glyph, title, right-aligned source identifier as the only right-hand field, no bullets; truncate with ellipsis keeping at least one space before the identifier
- [x] 4.2 Replace the `> `/gutter cursor with the full-row background highlight on article rows (matching the existing day-header selected style)
- [x] 4.3 Update row-format expectations in `daygroups_test` (`TestArticleRowShowsLocalTime`), `render_test`, and `polish_test`; add tests for the truncation gap, bullet-free rows, and full-row selection styling

## 5. Verification

- [x] 5.1 Run `go build ./...` and `go test ./...` — all green
- [x] 5.2 Run `go vet ./...` and the project linter; fix findings
- [x] 5.3 Manual smoke test: tree rail with fold/collapse, today/yesterday labels across a local-midnight boundary, ASCII mode (`-a`), narrow-terminal truncation

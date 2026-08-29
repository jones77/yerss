## 1. Key aliases

- [x] 1.1 Add `b` to `Back` default keys and `o` to `OpenArticle` default keys in `DefaultKeybindings()` (`internal/config/keys.go`)
- [x] 1.2 Verify per-view resolution: `o` → `open_article` in list, `o` → `open_url` in article (no conflict via `EffectiveKeys`)
- [x] 1.3 Test: `b` triggers back in all views; `o` opens article in list view; `o` opens URL in article view; existing `enter`/`l`/`h`/`esc` unchanged

## 2. Enable mouse

- [x] 2.1 Add `tea.WithMouseCellMotion()` to `tea.NewProgram` in `cmd/yerss/main.go`
- [x] 2.2 Handle `tea.MouseMsg` in `Model.Update()` (`internal/ui/model.go`), dispatching to list-view or article-view mouse handler based on `m.view`
- [x] 2.3 Verify popups still receive key events and mouse clicks outside popups don't interfere with popup state

## 3. List view mouse handling

- [x] 3.1 Implement click handler: map screen Y to visible-row index via `listWindow` start offset; look up row kind in `visibleRows()`
- [x] 3.2 Click on article row: set `m.list.cursor` to the row index, then call `m.openArticle()`
- [x] 3.3 Click on day-header row: toggle that group's collapse/expand via `m.toggle(groupIdx)`
- [x] 3.4 Click on status bar (Y == height-1): ignore (no-op)
- [x] 3.5 Wheel down: `m.moveListCursor(1)`; wheel up: `m.moveListCursor(-1)`
- [x] 3.6 Test: click opens correct article, cursor preserved on return; click toggles day header; wheel moves cursor; status bar click ignored

## 4. Article view mouse handling

- [x] 4.1 Wheel down: `m.article.viewport.ScrollDown(1)`; wheel up: `m.article.viewport.ScrollUp(1)`
- [x] 4.2 Click events in article view: no-op (links handled by terminal via OSC 8)
- [x] 4.3 Test: wheel scrolls viewport up and down by one line

## 5. OSC 8 clickable links in article body

- [x] 5.1 Extract `<a href="url">text</a>` pairs from raw HTML and replace with `ansi.SetHyperlink(url) + "text" + ansi.ResetHyperlink()` before `html2text` (`internal/ui/html.go`)
- [x] 5.2 Verify `html2text` still appends `(url)` as plain text alongside the clickable link text
- [x] 5.3 Test: link text wrapped in OSC 8 sequences carrying the URL; URL display text preserved; non-link text unaffected

## 6. OSC 8 clickable article header URL

- [x] 6.1 Wrap `a.Link` in `ansi.SetHyperlink(a.Link) + a.Link + ansi.ResetHyperlink()` in `renderArticleContent` (`internal/ui/article.go`)
- [x] 6.2 Test: header URL line carries OSC 8 sequences; visible text unchanged

## 7. Width calc switch (runewidth → ansi)

- [x] 7.1 Switch `wrapText` to use `ansi.StringWidth` for width measurement (`internal/ui/html.go`)
- [x] 7.2 Switch `truncate` to use `ansi.Truncate` (`internal/ui/html.go`)
- [x] 7.3 Switch `padRight` to use `ansi.StringWidth` (`internal/ui/html.go`)
- [x] 7.4 Verify existing `wrapText`/`truncate`/`padRight` tests pass unchanged on plain text
- [x] 7.5 Test: `wrapText` and `truncate` produce correct widths when input contains OSC 8 escape sequences

## 8. ASCII mode interaction

- [x] 8.1 Verify OSC 8 links are emitted in `--ascii` mode (OSC 8 is an escape sequence, not a Unicode glyph; ASCII mode only swaps box-drawing glyphs)
- [x] 8.2 Test: article with links in ASCII mode renders OSC 8 sequences alongside ASCII borders

## 9. Validation

- [x] 9.1 Run `go test ./...` and linter; fix any failures
- [x] 9.2 `openspec validate --changes mouse-and-clickable-links --strict` passes
- [ ] 9.3 Manual smoke test in Ghostty: click article in list, wheel scroll list, wheel scroll article, Cmd-click a link in article body, Cmd-click article header URL, press `b` and `o`

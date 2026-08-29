## 1. Shared bullet glyph

- [x] 1.1 Add a `bullet` field to `borderGlyphs` in `internal/ui/border.go` (`·` Unicode, `.` ASCII) and set it in `glyphsFor`, per the AGENTS.md glyph-pair convention
- [x] 1.2 Switch the existing top-border ` · ` separator in `topBorder` to use `g.bullet` so the bullet is glyph-consistent and ASCII-aware
- [x] 1.3 Test: top border separator uses `·` in Unicode mode and `.` in ASCII mode

## 2. List view row layout

- [x] 2.1 Add a source-identifier helper in `internal/ui/list.go`: parse a URL host with `net/url`, strip a leading `www.`, take the label before the final dot, truncate to 12 characters; fall back to the feed URL host when the article link is empty
- [x] 2.2 Change `renderArticleRow` to compose `title · source · time` using `g.bullet`, with time formatted as `15:04` (drop seconds)
- [x] 2.3 Apply the color rule: `source` and `time` always dim (`p.Dim`); `title` bold+`p.Bold` when unread, dim when read; selection background across the whole row
- [x] 2.4 Pass the article's `Link` and the feed URL into `renderArticleRow` so the source identifier can be derived
- [x] 2.5 Test: row shows `title · source · HH:MM`; source truncates to 12 chars (`reallylongnewspaperdomainname.net` → `reallylongne`); source falls back to feed host when link is empty; read/unread title color differs while source+time stay dim

## 3. Article frame restyle

- [x] 3.1 In `renderArticleBorder`, style the left vertical `g.v` with `p.Border` (grey) instead of emitting it raw; keep content text styling unchanged
- [x] 3.2 Change the right-edge track to a grey single-line glyph (`│` Unicode / `:` ASCII) in `p.Border`
- [x] 3.3 Change the scrollbar thumb to a bright-white double-line glyph (`║` Unicode / `|` ASCII) styled bold+white, replacing the accent/red double-line thumb
- [x] 3.4 Update `glyphsFor` ASCII values: `v=":"`, `unfill=":"`, `fill="|"` (left and right borders both `:`, thumb `|`)
- [x] 3.5 Test: left border is grey; track is grey single line; thumb is bright double line; ASCII mode shows `:` borders and `|` thumb with grey `:` / white `|`

## 4. Article top border date+time

- [x] 4.1 Change the top-border date prefix in `topBorder` from `2006-01-02` to `2006-01-02 15:04:05` using `PublishedAt.Local()`, followed by `g.bullet` and the title
- [x] 4.2 Thread the published time into `renderArticleBorder`/`topBorder` (replace the `date` string parameter or add a time parameter)
- [x] 4.3 Test: top border shows `YYYY-MM-DD HH:MM:SS · title`; title truncation still kicks in with the wider prefix; ASCII uses `.` bullet

## 5. Article bottom border indicator + help hint

- [x] 5.1 Compute `<bottomLine>` = `min(YOffset + viewportHeight, totalLines)` and `<totalLines>` = `TotalLineCount()` from the article viewport state
- [x] 5.2 Replace `bottomBorder`'s centered `X% scrolled` with a left-aligned `o: open in browser` hint and a right-aligned `<percent>% · <bottomLine>/<totalLines>` indicator, dashes filling between, corners retained
- [x] 5.3 Use `g.bullet` between the percent and the line ratio
- [x] 5.4 Test: short article shows `100% · <n>/<n>`; scrolled article shows percent and a consistent `<bottomLine>/<totalLines>`; help hint is left-aligned; degenerate widths still render safely

## 6. Markdown links with OSC 8

- [x] 6.1 In `wrapLinks` (`internal/ui/html.go`), replace `<a href="url">inner</a>` with `ansi.SetHyperlink(url) + "[" + inner + "]" + ansi.ResetHyperlink() + "(url)"`
- [x] 6.2 Set `html2text.Options{OmitLinks: true}` in `HTMLToText` so html2text no longer appends its own `( url )`
- [x] 6.3 Test: a link renders as `[text](url)` with the `text` wrapped in OSC 8 carrying the URL and the URL shown as plain text; non-link text unaffected

## 7. Link-aware URL wrapping and truncation

- [x] 7.1 Make `wrapText` link-aware: when a token containing a `](url)` markdown link does not fit on the current line, break at the `]` boundary so `[text]` stays (if it fits) and `(url)` starts the next line
- [x] 7.2 When a `(url)` alone is wider than the line width, truncate the URL with `ansi.Truncate` + the ellipsis glyph (`…` / ASCII `...`) to fit one line
- [x] 7.3 Test: a link whose markdown form does not fit breaks the URL onto a new line; an overlong URL is truncated with an ellipsis on a single line

## 8. Help popup labels

- [x] 8.1 Add an action display-name helper in `internal/ui/popup.go`: replace `_` with spaces and Title Case the result (e.g. `open_article` → `Open Article`)
- [x] 8.2 Use it for the left column in `renderHelp` instead of the raw action string
- [x] 8.3 Test: help popup labels are Title Case without underscores (`Open Article`, `Mark All Read`, `Toggle Read`)

## 9. Keybinding defaults

- [x] 9.1 In `DefaultKeybindings` (`internal/config/keys.go`), add `t` to `TagPopup` and `r` to `Refresh`; leave `G` (bottom) uppercase only
- [x] 9.2 Verify `Validate`/`EffectiveKeys` accept the widened defaults with no conflict across all views
- [x] 9.3 Test: `t` and `T` both open the tag popup; `r` and `R` both refresh; `g` stays top and `G` stays bottom; existing bindings unchanged

## 10. Validation

- [x] 10.1 Run `go test ./...` and linter; fix any failures
- [x] 10.2 `openspec validate --changes reader-ui-polish --strict` passes
- [ ] 10.3 Manual smoke test: list row source/HH:MM/colors; article border left-grey + bright-thumb + bottom indicator + help hint; markdown links clickable and wrapping; help popup labels; `t`/`T` and `r`/`R`

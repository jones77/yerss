## Why

yerss is keyboard-only today: links are inert plain text, mouse input is
ignored, and navigation relies on vi/arrow keys. Modern terminals (Ghostty,
iTerm2, Kitty) support clickable OSC 8 hyperlinks and mouse event reporting,
and bubbletea already provides the mouse event API — yerss just doesn't enable
or handle it. Adding mouse support and clickable links brings the reader in
line with how users expect terminal apps to behave in 2026.

## What Changes

- Enable mouse cell-motion reporting (`tea.WithMouseCellMotion`) so the app
  receives click and wheel events.
- Handle mouse events in the list view: a single click on an article row sets
  the cursor to that row and opens the article; a click on a day-header row
  toggles its collapse/expand; wheel up/down moves the cursor by one row; a
  click on the status bar is ignored.
- Handle mouse events in the article view: wheel up/down scrolls the article
  viewport by one line.
- Render inline article links as clickable OSC 8 hyperlinks using
  `ansi.SetHyperlink`/`ResetHyperlink` (from `charmbracelet/x/ansi`, already in
  the dependency tree). The link URL display text is kept alongside the
  clickable link text as a safe fallback for terminals without OSC 8 support.
- Render the article's own URL (shown in the reader header) as an OSC 8
  hyperlink so Cmd/Ctrl-clicking it opens the browser, matching the existing
  `o` key behavior.
- Switch width calculations in `wrapText`, `truncate`, and `padRight`
  (`internal/ui/html.go`) from `runewidth.StringWidth` to
  `ansi.StringWidth`/`ansi.Truncate` so OSC 8 escape bytes do not corrupt
  layout math. `model.go` already uses the `ansi` package for this — the
  change aligns the older helpers with the established pattern.
- OSC 8 hyperlinks SHALL function in `--ascii` mode (OSC 8 is an escape
  sequence, not a Unicode glyph, consistent with ANSI color escapes being
  kept in ASCII mode).
- Add `b` as an alias for the `back` action (same as `esc`, `h`, left arrow)
  in all views.
- Add `o` as an alias for the `open_article` action in the list view (same as
  `enter`, `l`, right arrow). In the article view `o` already opens the
  article URL in the browser (`open_url` action) — unchanged.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Keyboard navigation" requirement is modified to add `b`
  as a back alias and `o` as an open-article alias in the list view. New
  requirements cover mouse navigation (click and wheel in list and article
  views) and clickable links (OSC 8 hyperlinks in article body and header).
- `configuration`: The "Configurable keybindings" requirement is modified to
  reflect the updated default keybindings (`b` → back, `o` → open_article in
  list view).

## Impact

- **Program startup**: `cmd/yerss/main.go` — add `tea.WithMouseCellMotion()`
  to `tea.NewProgram`.
- **Event handling**: `internal/ui/model.go` — handle `tea.MouseMsg` in
  `Update()`, dispatching to list-view and article-view mouse handlers.
- **List view**: `internal/ui/list.go` — click-to-select-and-open, click on
  day-header to toggle, wheel-scroll cursor; coordinate mapping from screen
  Y to visible-row index using the existing `listWindow`/`visibleRows`
  machinery.
- **Article view**: `internal/ui/article.go` — wheel-scroll viewport.
- **HTML rendering**: `internal/ui/html.go` — inject OSC 8 around link text
  before `html2text`; switch `wrapText`/`truncate`/`padRight` from
  `runewidth` to `ansi.StringWidth`/`ansi.Truncate`.
- **Article header**: `internal/ui/article.go` — wrap `a.Link` in OSC 8.
- **Keybindings**: `internal/config/keys.go` — add `b` to `Back` defaults,
  add `o` to `OpenArticle` defaults.
- **Dependencies**: none new — `charmbracelet/x/ansi` (v0.11.6) and
  `charmbracelet/bubbletea` (v1.3.10) already in `go.mod`.
- **Tests**: mouse coordinate mapping, click-open, day-header toggle,
  wheel-scroll, OSC 8 link injection, width calc with OSC 8 bytes, key alias
  resolution.

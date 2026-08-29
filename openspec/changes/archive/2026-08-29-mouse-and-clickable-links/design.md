## Context

The article reader renders HTML as plain text via `html2text` with link URLs
shown inline as plain text (`internal/ui/html.go`). The viewport is a
`bubbles/viewport.Model` scrolling by line, framed by `renderArticleBorder`
(`internal/ui/border.go`). Width math uses `mattn/go-runewidth`; `model.go`
already uses `charmbracelet/x/ansi` for ANSI-aware width/truncation in the
overlay compositor. Mouse events are not enabled — `tea.NewProgram` runs with
only `tea.WithAltScreen()`. Keybindings live in `internal/config/keys.go` with
a per-view resolution system that prevents cross-view conflicts.

`charmbracelet/x/ansi` v0.11.6 (already in `go.mod`) provides
`SetHyperlink`/`ResetHyperlink` (OSC 8), `StringWidth` (ANSI-aware width),
`Truncate` (ANSI-aware truncation), and `Wordwrap` (ANSI-aware word wrap that
preserves OSC 8 sequences across line breaks — verified by its test suite).
`charmbracelet/bubbletea` v1.3.10 provides `tea.WithMouseCellMotion`,
`tea.MouseMsg`/`MouseEvent` with `X`, `Y`, `Button`, `Type`, and
`MouseWheelUp`/`MouseWheelDown`/`MouseLeft` event types.

## Goals / Non-Goals

**Goals:**
- Enable mouse click and wheel navigation in list and article views.
- Render article body links and the article header URL as OSC 8 clickable
  hyperlinks with plain-text URL fallback.
- Add `b` and `o` as key aliases matching existing navigation actions.
- Keep layout math correct when content carries OSC 8 escape bytes.

**Non-Goals:**
- Hover effects (requires `WithMouseAllMotion`, which not all terminals
  support and which can interfere with text selection).
- Mouse drag selection or scrolling by pixel.
- Clickable links in the list view (list rows are not links).
- Changing the URL display-text policy (URLs remain visible alongside
  clickable link text).

## Decisions

### D1: Cell-motion mouse, not all-motion

`tea.WithMouseCellMotion()` enables click, release, and wheel events. It does
not report motion unless a button is held. `WithMouseAllMotion` would enable
hover but is less widely supported and can interfere with terminal text
selection. Cell-motion is the right tradeoff: click and wheel work everywhere
modern, and users can still select text by dragging (the terminal handles it
locally when the app doesn't claim all-motion).

### D2: Click-to-open (single click, not double)

A single left-click on an article row sets the cursor to that row and
immediately opens the article. This matches newsboat-style behavior and is the
most direct mapping. Double-click-to-open was considered and rejected — it
adds latency and complexity (click-count tracking) for no gain in a TUI where
the list is the primary navigation surface.

### D3: List click coordinate mapping

The list renders `visible` rows starting at `start` (from `listWindow`). A
click at screen Y maps to visible-row index `start + Y`. The row at that index
is looked up in `visibleRows()` to determine if it's a header (toggle
collapse) or an article (set cursor + open). The bottom line (Y == height-1)
is the status bar and is ignored. This reuses the existing `listWindow`/
`visibleRows` machinery — no new coordinate system.

### D4: OSC 8 injection before html2text

Link `<a href="url">text</a>` pairs are extracted from the raw HTML and
replaced with `SetHyperlink(url) + "text" + ResetHyperlink()` before
`html2text` processes it. `html2text` preserves the escape sequences as
opaque text (they're not HTML entities), and its link-URL-display option
still appends `(url)` as plain text. Result: clickable link text + visible
URL, in one pass.

**Alternative considered:** post-process the `html2text` output to find link
text and wrap it. Rejected — `html2text` loses the href→text association
during conversion, making it impossible to reliably re-associate URLs with
their link text afterward.

### D5: Switch width math from runewidth to ansi

`wrapText`, `truncate`, and `padRight` in `html.go` use
`runewidth.StringWidth`, which counts every byte including ANSI escape
sequences. Once content carries OSC 8 bytes, these functions would
miscount and corrupt the layout (over-padded lines, broken truncation,
wrap-at-wrong-column). Switching to `ansi.StringWidth` and `ansi.Truncate`
(excluding escape sequences from the count) fixes this. `model.go` already
uses these functions for the overlay compositor — this aligns the older
helpers with the established pattern, not introducing a new one.

### D6: Article header URL as OSC 8

`renderArticleContent` writes `a.Link` as a plain text line. Wrapping it in
`SetHyperlink(a.Link) + a.Link + ResetHyperlink()` makes it clickable. The
visible text is unchanged (the URL string itself), so terminals without OSC
8 see the same output. This mirrors `o`/`OpenURL` but via mouse.

### D7: Key aliases via DefaultKeybindings only

`b` is added to the `Back` action's default key list; `o` is added to the
`OpenArticle` action's default key list. The per-view resolution in
`EffectiveKeys` already handles `o` resolving to `open_article` in the list
view and `open_url` in the article view (they're different actions in
different view action-sets, so no conflict). No new actions, no new action
constants, no changes to the resolution logic.

## Risks / Trade-offs

- **[Mouse reporting may intercept OSC 8 link clicks in some terminals]** →
  Modern terminals (Ghostty, iTerm2, Kitty, WezTerm) handle modifier+click on
  OSC 8 links before sending the event to the app, so links remain clickable.
  Some older or less capable terminals may swallow all clicks when mouse
  reporting is on. yerss targets modern terminals; this is documented as a
  known limitation, not a blocker.
- **[OSC 8 support is terminal-dependent]** → The plain-text URL fallback
  ensures non-supporting terminals look identical to today. No graceful
  detection is attempted — OSC 8 sequences are silently ignored by
  non-supporting terminals, which is the correct fallback behavior.
- **[runewidth → ansi switch may change edge-case widths]** → Both libraries
  measure visible display width; the switch only changes behavior for strings
  containing ANSI escapes (which the old helpers miscounted anyway). Existing
  tests validate wrap/truncate behavior on plain text and should pass
  unchanged.
- **[Click on collapsed day-header vs article-row ambiguity]** → The
  `visibleRows()` flatten already distinguishes `rowHeader` from `rowArticle`,
  so the click handler dispatches correctly with no ambiguity.

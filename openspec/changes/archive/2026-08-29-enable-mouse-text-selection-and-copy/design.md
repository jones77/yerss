## Context

The article reader renders HTML as plain text via `HTMLToText`
(`internal/ui/html.go`), wraps it to the content width, and feeds it to a
`bubbles/viewport.Model` (`internal/ui/article.go`). Lines carry ANSI from the
scrollbar/border chrome and OSC 8 hyperlink sequences on link lines (added in
`renderArticleContent`), so all width math on content is already ANSI-aware
via `charmbracelet/x/ansi` (`truncate`, `padRight`, `wrapText`). The app runs
with `tea.WithMouseCellMotion()` (`cmd/yerss/main.go`), so it already receives
press, release, wheel, and drag-while-held mouse events. The article view's
mouse handler ignores clicks by design (`internal/ui/article.go:140`). Mouse
coordinates are terminal cells; the content area begins at
column `padX+1` and row `1+effPadY` inside the border frame
(`internal/ui/border.go`). The clipboard helper in `internal/ui/open.go`
pipes a value to `pbcopy`/`clip`/`xclip` via stdin without a shell.

## Goals / Non-Goals

**Goals:**
- In-app mouse text selection in the article view, rendered with an inverted
  highlight, using the existing cell-motion mouse mode.
- Automatic copy of the selected text to the clipboard on release.
- A `C` (shift-c) key that copies the full article text, distinct from `c`
  (`copy_url`).
- All copied text is plain — ANSI styling and OSC 8 escapes stripped.
- Reuse the existing non-shell clipboard mechanism for all copies.

**Non-Goals:**
- Text selection in the list view (list rows remain click-to-open).
- Hover/passive highlighting (requires all-motion mode, which interferes with
  the drag tracking this feature needs).
- Auto-scrolling the article when a drag reaches the viewport edge.
- Visual feedback (status bar) beyond the existing message pattern.
- Changing what counts as "article text" for the `C` key (the same content
  `renderArticleContent` produces, escapes stripped).

## Decisions

### D1: App-owned selection via cell-motion events, not terminal-native selection

`WithMouseCellMotion()` already claims drag events, which disables the
terminal's native selection. The app therefore tracks the selection itself:
press anchors, motion extends, release finalizes and copies. The alternative —
relying on terminal-native selection — does not work because the app owns the
mouse. Selection state is stored on `articleState` as an anchor cell and a
current cell in viewport coordinates, normalized to a range at render time.

### D2: Selection is a cell range over the visible wrapped lines

Mouse coordinates map to viewport cells via
`x = X - 1 - padX`, `y = Y - 1 - effPadY`, where `effPadY` is the clamped
vertical padding computed by `renderArticleBorder` (only differs from
`cfg.Display.PaddingY` on degenerate terminals). Presses outside the content
area (borders, padding, status bar, popup) are ignored. `x` is clamped to the
line width and `y` to the visible viewport height. The range spans the visible
window of `articleState.lines`, so it tracks scroll automatically because the
viewport's `View()` output is what's rendered.

The alternative — tracking selection against the underlying unwrapped text —
was rejected: mapping a cell range back through wrap positions is more complex
and produces the same visible result, since the user selects what they see.

### D3: Rendering the highlight with ansi.Cut

Each visible content line may embed ANSI (OSC 8). To invert a cell subrange
without corrupting escape sequences, split the line into before/selected/after
with `ansi.Cut` (cell-width aware), wrap the middle segment in a lipgloss
`Inverted(true)` style, and concatenate. This is the same ANSI-aware approach
the rest of the render path uses. Cell-to-cell width math is correct for wide
runes (CJK) because `ansi.Cut` counts cells, not runes.

### D4: Copying selected text by stripping then slicing cells

On release, for each line in the selection's row range: `ansi.Strip` the line
(removing OSC 8 and styling), slice the cell range `[lo, hi)` with `ansi.Cut`
on the stripped string, trim trailing padding spaces, and join lines with
`\n`. Feed the result to the clipboard via stdin. A zero-width selection
(`lo == hi`) copies nothing and sends no command.

### D5: Generalize the clipboard helper

Refactor `copyURLCmd` (`internal/ui/open.go`) into `copyTextCmd(text string)`
that pipes any text to the platform clipboard helper via stdin; `copyURLCmd`
becomes `copyTextCmd(url)`. The `C` key and selection release both call
`copyTextCmd`. The existing `urlActionMsg` is extended with a text-copy action
so the status line reads e.g. "copied article text" / "copied selection"
instead of "copied URL".

### D6: Modifier-clicks are not selections

The spec requires Cmd/Ctrl-clicks to be left alone so terminals can activate
OSC 8 hyperlinks natively. Selection starts only on left-press events with
`Mod == ModNone`; drag/release with a modifier are likewise ignored. Wheel
behavior is unchanged.

### D7: New action in the config key system

Add `CopyArticleText Action = "copy_article_text"` to `internal/config/keys.go`,
bound to `{"C"}` by default, registered in `allActions` and
`actionsForView(ViewArticle)`. `C` is currently unbound, so there is no
conflict; case-sensitivity keeps it distinct from `c` (`copy_url`). The help
popup derives the label "Copy Article Text" automatically from the action id.
`updateArticle` dispatches `config.CopyArticleText` to `copyTextCmd` on the
full article text (`renderArticleContent`, `ansi.Strip`ped).

## Risks / Trade-offs

- [Some terminals deliver drag motion only after a threshold, or map release
  with `Button: None`] → Clamp release handling to the last known position;
  treat a release event as finalizing regardless of its button field.
- [OSC 8 escapes on a link line make per-line width math subtle] → Reuse
  `ansi.Cut`/`ansi.Strip` (already exercised by the wrap/truncate path); add
  a test covering selection over a link line.
- [Selection survives scroll, so a drag followed by wheel scroll leaves
  highlight on different text] → Acceptable: the spec only requires the
  highlight to track the viewport at the current offset; a new press replaces
  the selection.
- [Degenerate terminal sizes (height 0) could panic in coordinate math] →
  Clamp `y`/`x` against the floored viewport/content dimensions before
  indexing, matching the existing render-safety requirement.

## Migration Plan

No storage, config-file, or data migration. Rollback is a revert of the
implementation commit. `C` is a new default binding; users who already bind
`C` in their config would now conflict at load time — acceptable for a
previously-unbound key, and remappable via the existing keybinding config.

## Open Questions

None.
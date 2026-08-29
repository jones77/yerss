## Why

Since `tea.WithMouseCellMotion()` was enabled, the app claims mouse events for
the whole screen, which disables the terminal's native click-and-drag text
selection inside the article reader. Users can no longer grab a passage with
the mouse, and there is no way to copy an article's text to the clipboard —
only its URL (`c`). Reading is a reading/copying workflow; restoring
in-app selection and adding one-key copy closes the gap.

## What Changes

- Track mouse text selection in the article view: pressing the left button in
  the content area anchors the selection, dragging extends it, and releasing
  finalizes it. The selected region is rendered with an inverted/highlight
  style while active and immediately after release.
- On selection release, automatically copy the selected text to the system
  clipboard (no extra key needed), so drag-selecting text is the whole copy
  flow.
- Add a new `copy_article_text` action bound to `C` (shift-c) in the article
  view that copies the article's full text content to the system clipboard.
  Lowercase `c` stays bound to the existing `copy_url` action; the two are
  distinct keys.
- Copy uses the existing non-shell clipboard helper pattern (feed text via
  stdin to `pbcopy`/`clip`/`xclip`), which is already proven safe for
  feed-controlled data.
- Copied text is plain text with no ANSI styling and no OSC 8 escape
  sequences; the article text is the same content `renderArticleContent`
  produces (title, author, link, body) with escapes stripped.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Mouse navigation" requirement is modified so left-click in
  the article view starts a text selection (instead of being a no-op) and
  wheel behavior is unchanged. New requirements cover in-view text selection
  with automatic copy on release, and the `C` copy-article-text key in the
  "Quit and utility keys" requirement.
- `configuration`: The "Configurable keybindings" requirement is modified to
  add `C` as the default binding for the new `copy_article_text` action, kept
  distinct from lowercase `c` (`copy_url`) per the case-sensitivity rule.

## Impact

- **Article view**: `internal/ui/article.go` — selection state, press/drag/
  release handling in `updateArticleMouse`, selected-cell rendering in
  `renderArticle`, and text extraction from the wrapped viewport content.
- **Model**: `internal/ui/model.go` — selection state storage on
  `articleState`, dispatch of `copy_article_text`.
- **Clipboard**: `internal/ui/open.go` — generalize the existing
  URL-copy helper into a text-copy helper reused by both `copy_url` and the
  new article-text copy.
- **Keybindings**: `internal/config/keys.go` — new `CopyArticleText` action,
  default `C` binding, view-actions membership for the article view, and help
  popup label.
- **Tests**: selection anchor/extend/release mapping to viewport lines,
  auto-copy on release, `C` copies full article text, ANSI/OSC 8 stripping in
  copied text.
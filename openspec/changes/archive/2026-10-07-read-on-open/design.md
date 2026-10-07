## Context

See `proposal.md` — Why for motivation. Current state (from code):

- `composeArticle` (`internal/ui/articleview/article.go`) ends with
  `if len(st.lines) <= vpH { st.MarkRead(s.sess.Store()) }`, so an article that
  fits the viewport is marked read on open and a longer one is not.
- `UpdateKey` (`internal/ui/articleview/update.go`) calls
  `s.MarkRead(s.sess.Store())` for `MoveDown`, `PageDown`, and `HalfPageDown`
  only; `UpdateMouse` calls it for a downward wheel only. Upward scrolls, page
  ups, half-page ups, and top/bottom jumps do not mark read.
- `MarkRead` (`internal/ui/articleview/selection.go`) is idempotent: it returns
  early when `s.readMarked || s.article.Read`, then sets both and writes
  `SetRead(s.id, true)`.
- `Recompose` routes through `composeArticle` and is called on open, on image
  loads, on resize, and on startup restore; `base()` does not carry `readMarked`
  across recomposes, so the guard is the persisted `s.article.Read` for repeats
  of the same article.

## Goals / Non-Goals

**Goals:**
- An article that fits the viewport is marked read on open.
- An article that does not fit is marked read on the user's first movement in
  any direction (line/page/half-page scroll, top/bottom jump, or wheel).
- A movement marks read even when the scroll position does not change.
- Restoring a saved offset does not mark read.
- Exactly one `SetRead` write per article (no repeat writes on recomposes).

**Non-Goals:**
- Any toggle-read or mark-all-read binding (still absent by design).
- Changing the read/unread list styling or the light/dark palette.
- Changing how `SetRead` persists or how `backToList` reloads the list.

## Decisions

### Keep the fit-on-open guard; mark read on any movement

`composeArticle` keeps the `if len(st.lines) <= vpH` guard, so a fitting article
is read on open and a scrollable article is not. Every scroll/jump action — up,
down, page up/down, half-page up/down, top, bottom, and both wheel directions —
calls `MarkRead`, so the first user movement through a long article marks it
read. `MarkRead`'s guard makes subsequent calls (including the one
`composeArticle` would fire for a fitting article) no-ops.

### Treat a movement as the action, not an offset change

Marking is keyed on the user issuing a scroll/jump action, not on the viewport
offset actually changing. This keeps the rule simple and avoids edge cases
around `Up` at the top or `Down` at the bottom of a long article.

### Restore is not movement

`restoreSelection` and image-load recomposes that re-apply a saved offset must
not mark read, because they are not user movement. The `MarkRead` calls live
only in the key/mouse scroll handlers and the fit-on-open branch, never in
`Recompose`/`restoreSelection` themselves.

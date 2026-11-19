## Context

The article view renders a single article with no way to step to the next or
previous article. The list view already orders articles newest-first under an
optional tag filter, and `j`/`k` move down (older) / up (newer) through it. The
user wants `J`/`K` in the article view to step between articles in the same
order, plus a `n/total` position counter, a transient end/beginning hint, and a
fix for a help-popup key bug.

Two facts shape the design:

- The list already holds the ordered, filtered article IDs in memory. The
  article view does not need to know why a list is filtered — it can walk
  `listview.State.ArticleIDs()` and inherit the active filter for free.
- `bucketDayGroups` resets every day group's `collapsed` flag to `false` on
  every `Load()`. `backToList` calls `Load()`, so the list is always fully
  expanded when returning from an article. The list cursor therefore cannot be
  trusted to keep pointing at the same article across a `Load()` when groups
  were collapsed; selection must be restored by article ID, not by raw cursor
  index.

## Goals / Non-Goals

Goals:

- `J` (shift-j) steps to the previous article (older, down the list); `K`
  (shift-k) steps to the next article (newer, up the list). Wrapping at both
  ends.
- Navigation follows the active filter: walking a source/tag list stays inside
  that source/tag.
- A `n/total` counter in the article top border (right-aligned), where `n` is
  the 1-based position among the shown articles and `total` is the shown count.
- Fix the list status bar so `total` is the filtered count, not the whole-DB
  count.
- A transient bottom-border hint: `J: previous` at the end of a scrollable
  article, `K: next` at its beginning, replacing the `o:` hint for 2 seconds
  then reverting.
- `b` and `h` dismiss the help popup.

Non-Goals:

- No per-source vs. global dual scope. The source/tag filter already selects the
  scope; `J`/`K` inherit it.
- No persistence of the filter across restarts (that stays non-persistent, now
  made explicit in the spec).
- `J`/`K` do not mark the article read or fire image loads for anything but the
  article they open (the normal open path already handles the target article's
  images and read marking).

## Decisions

### 1. Direction and wrap semantics

`J` and `K` mirror the list's `j`/`k` exactly. The list orders newest first, so
a higher index is older.

- `J` = previous = older = `index + 1` in `ArticleIDs()`.
- `K` = next = newer = `index - 1` in `ArticleIDs()`.

Both wrap: from the newest article `K` wraps to the oldest, from the oldest `J`
wraps to the newest. The wrap matches the list's own wrapping behavior and
guarantees `J`/`K` always do something on a non-empty list.

### 2. Neighbor resolution and the "right thing" under a filter

The neighbor is resolved from `listview.State.ArticleIDs()`, the in-memory,
already-filtered, display-ordered ID list. The model finds the open article's
index, applies the step with wrap, and opens the neighbor by ID. No new store
query is needed; the scope is whatever filter the list currently holds, which
is exactly the "right thing" for a tag/source list.

When the open article's ID is not present in `ArticleIDs()` (article pruned
from the store, or an inconsistent restore), `J`/`K` are a no-op.

### 3. Opening an article by ID (shared by open and navigate)

`openArticle` currently opens `list.ArticleAtCursor()`. `J`/`K` need to open an
arbitrary ID. Extract a helper `openArticleByID(id int64) tea.Cmd` that loads
the full article, recomposes it, switches to the article view, and fires the
image load — then have both `openArticle` and the navigation path call it.
`openArticle` keeps resolving the ID from the cursor; navigation resolves it
from `ArticleIDs()`.

### 4. List cursor sync on return

Because `Load()` re-expands every group, `backToList` can no longer rely on the
raw cursor to reselect the open article. Add `listview.State.SelectArticleID(id
int64) bool`: find the article's `(groupIdx, artIdx)`, expand that group, set
the raw cursor to the article's visible row, and report whether it was found.
`backToList` calls `SelectArticleID(m.article.ID())` after `Load()`. This is a
strict improvement over the current cursor-preservation behavior (which can
drift when groups were collapsed) and keeps the existing "return to the same
article" spec true for all paths: open, navigate, and restore.

### 5. Position counter (`n/total`)

The model computes the position, because only it knows the list. After opening
an article, `setArticlePosition(id)` finds the article's index in
`ArticleIDs()` and pushes `n = index+1`, `total = len(ids)` into the article
view. When the article is not in the list it pushes `(0, 0)`.

The article view stores the position as `listPos`/`listTotal` and renders a
right-aligned `n/total` in the top border when `listPos > 0`. This mirrors the
list's bottom-right counter, but in the top border where the article view has
room.

### 6. Fix the list status bar total

`listview.State.selectedArticlePosition()` returns `n` (filtered position) and
`total = s.total` (whole-DB `ArticleCount`). Change it to return
`s.articleCount()` (the filtered, loaded count) so the list and article
counters agree. The now-unused `total` field and its `ArticleCount()` store
call in `Load()` are removed.

### 7. Transient hint and its timer

The hint is owned by the article view, which is the only place that knows the
scroll position. It adds two unexported fields — `navHint` (none/previous/next)
and `navHintExpires` — plus:

- A scroll action that lands at the top edge of a *scrollable* article sets
  `navHint = next`; landing at the bottom edge sets `navHint = previous`;
  landing anywhere else clears it. A scrollable article is one where
  `Viewport.TotalLineCount() > Viewport.Height`; a short article never shows
  the hint.
- Arming the hint also returns a `tea.Tick(2*time.Second, ...)` command
  carrying `articleview.NavHintExpiredMsg`. The model handles that message by
  calling `articleview.ExpireNavHint()`, which clears the hint only when
  `time.Now() >= navHintExpires` (so a re-arm keeps the hint alive and stale
  ticks are harmless).
- `NavHint()` returns the border label: `J: previous`, `K: next`, or empty.
  `RenderUncached` passes this to the border renderer, which substitutes it for
  the `o: open article in browser` hint when non-empty.
- `ResetNavHint()` is called when an article is opened (open or navigate), so a
  freshly opened article never flashes a hint.

The bottom-border renderer changes from a hard-coded `o:` string to a
`hint string` parameter threaded through `BottomBorder` and
`RenderArticleBorder`.

### 8. Keybindings

Two new configuration actions are added to the article view, distinct by case
from the lowercase movement keys:

- `previous_article` bound to `J`.
- `next_article` bound to `K`.

`articleview.UpdateKey` maps these configuration actions to two new article-view
actions (`ActionPreviousArticle` → step `+1`, `ActionNextArticle` → step `-1`)
that the model routes to the navigation path. `J`/`K` are uppercase and do not
collide with the article view's existing lowercase `j`/`k` scroll keys.

### 9. Help popup `b`/`h` dismiss

The `KindHelp` case in `popup.Update` hard-codes its dismiss keys. Add `b` and
`h` to the existing `esc`/`enter`/` ` close case. This is a one-line change;
the help popup is shared by list and article, so both dismiss on `b`/`h`.

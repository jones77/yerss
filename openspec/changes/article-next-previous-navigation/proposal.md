## Why

Reading a feed means constant back-and-forth: to reach the next article the
reader must leave the article view, move the list cursor, and open again. A
reader that lets you step directly to the previous (older) or next (newer)
article — and tells you where you are in the list — removes that friction.
Because a source tag already captures "articles for a publication", the same
step naturally stays within a publication or tag when one is active.

## What Changes

- The article view gains `J` (previous = older, down the list) and `K` (next =
  newer, up the list), mirroring the list view's `j`/`k` direction. Both wrap
  at the ends and follow the currently loaded (possibly filtered) list.
- The article top border shows a right-aligned `n/total` position counter,
  mirroring the list status bar.
- The list status bar's `total` is fixed: it SHALL be the number of articles
  currently shown after any active tag filter, not the whole-database count.
- Scrolling to the end of an article shows `J: previous`, and to the start
  shows `K: next`, in the bottom-border hint. The hint replaces `o: open
  article in browser` for two seconds, then the `o:` hint returns.
- The help popup dismisses on `b` and `h` (bug fix).
- Spec drift is reconciled: `q` now quits from a popup (it no longer closes
  the popup), matching the implemented behavior.
- The de-facto behavior "the active tag filter is not restored on relaunch; the
  list always reopens unfiltered" is made explicit in the spec.

## Why Not

- Not a two-axis "left/right = publication, up/down = within-publication"
  scheme: a source tag already scopes a list, and publications have no natural
  ordering. One key pair over the active filter is simpler and reuses the
  list's ordering.
- No separate all-articles vs publication counter: the position counter reads
  over whatever list is currently shown, so it is inherently filter-aware.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: next/previous article navigation, the article position counter,
  the fading navigation hint, the help-popup `b`/`h` dismiss, the `q`-quits
  reconciliation, the fixed filtered-count `total`, and the explicit
  "filter is not restored" rule.
- `configuration`: the new `next_article` (`K`) and `previous_article` (`J`)
  default bindings in the article view.

## Impact

- **UI layer** (`internal/ui`): a new `navigateArticle` path in `model.go` /
  `model_article.go` steps by neighbor index and moves the list cursor; the
  list gains a select-by-article-ID helper; the article view carries a
  position/total pair and a transient navigation hint with a two-second timer
  (the app's first timer).
- **Rendering** (`internal/ui/render`): `TopBorder` gains a right-aligned
  `n/total` segment; the article bottom-border hint becomes dynamic
  (`o:` ↔ `J: previous` / `K: next`).
- **List view** (`internal/ui/listview`): `selectedArticlePosition` reports the
  filtered count as `total` instead of the whole-database count.
- **Popup** (`internal/ui/popup`): the help popup dismisses on `b`/`h`.
- **Config** (`internal/config`): new `next_article`/`previous_article` actions
  and their `K`/`J` article-view bindings.
- **Tests**: navigation direction, wrap, filter scope, and no-op; the filtered
  counter; the fading hint; the help-popup dismiss; the `q`-quits behavior.

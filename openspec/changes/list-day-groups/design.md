# Design: list-day-groups

## Data model

`articleItem` gains the publication time, and the list becomes day-bucketed:

```go
type articleItem struct {
    ID          int64
    Title       string
    Read        bool
    PublishedAt time.Time   // new
}

type dayGroup struct {
    date      time.Time      // local midnight of the day
    label     string         // "Saturday 20th August, 2025"
    collapsed bool
    articles  []articleItem
}

type listState struct {
    groups []dayGroup
    cursor int   // index into the flattened visible-rows list
    filter string
}
```

`Store.ListArticles` already returns articles reverse chronological
(`ORDER BY published_at DESC, id DESC`), so `loadList` only needs to bucket
consecutive articles into day groups by `PublishedAt.Local()`.

## Visible-row flattening

Rendering and navigation operate on a flattened list of typed rows so collapse
hides articles uniformly:

```go
type rowKind int
const ( rowHeader rowKind = iota; rowArticle )

type visibleRow struct {
    kind     rowKind
    groupIdx int
    artIdx   int   // valid for rowArticle
}
```

```
visibleRows(): for each group (desc):
                 emit rowHeader(groupIdx)
                 if !group.collapsed: emit rowArticle for each article
```

`cursor` indexes `visibleRows()`. `moveListCursor` walks this list with wrap.
`renderList` iterates the same slice (windowed around the cursor as today).

## Date label and ordinal

```go
func longDate(t time.Time) string {
    d := t.Day()
    suf := "th"
    switch d % 10 { case 1: suf="st"; case 2: suf="nd"; case 3: suf="rd" }
    if d/10 == 1 { suf = "th" }           // 11th–13th
    return t.Format("Monday ") + fmt.Sprintf("%d%s ", d, suf) + t.Format("January, 2006")
}
```

Day bucketing uses `t.Local()` truncated to midnight; the timestamp column uses
`t.Local().Format("15:04:05")`.

## Row layout

```
  Saturday 28th August, 2026          ← header (dim/accent), indented none
>   Article one              14:32:01   ← cursor "> ", title, right-aligned time
    Article two              13:05:44
```

Width math: `timeW = 8`; gap `= 1`; `titleW = m.width - 2(cursor) - timeW - gap`.
Title truncated to `titleW`; the time column is right-aligned at the right edge.
Headers render with a fold glyph (`▾` expanded / `▸` collapsed) so collapse
state is visible; ASCII fallback uses `v`/`>` (see ascii-flag change for the
glyph-pair convention).

## Status bar `<n>/<total>`

New `Store.ArticleCount() (int, error)`:
```sql
SELECT COUNT(*) FROM articles
```
`n = len(m.list.articles)` (filtered, summed across groups); `total` fetched
once in `loadList` and stored on `listState` (or `Model`). Format
`fmt.Sprintf("%d/%d articles", n, total)`.

## Fold key dispatch

Fold keys are **context-sensitive**, dispatched in `updateList` by inspecting the
row under the cursor rather than via the action map, because `h`/`l`/Enter/Space
already map to `Back`/`OpenArticle`/`HalfPageDown` and the keymap validator
rejects one key bound to two actions in a view:

```go
row := m.visibleRows()[m.list.cursor]
if row.kind == rowHeader {
    switch msg.String() {
    case "h":       m.collapse(row.groupIdx)
    case "l":       m.expand(row.groupIdx)
    case "enter", " ", "tab": m.toggle(row.groupIdx)
    }
    return m, nil           // do not fall through to article actions
}
if msg.String() == "tab" {
    m.toggle(row.groupIdx)
    return m, nil
}
// otherwise: existing action-map dispatch (h clears filter, l/enter open, space half-page)
```

Tradeoff: fold keys are not user-remappable through `config.toml` (they reuse
existing keys contextually). Accepted for now; a future change could introduce
dedicated `Collapse`/`Expand`/`ToggleFold` actions if remappability is wanted.

## Edge cases

- **Undated articles** (`PublishedAt` zero): bucket into a trailing `Undated`
  group (sorts last by `COALESCE(published_at,0)` already), labeled `Undated`.
- **Filter active**: `loadList` filters via `ListArticles(filter)`; groups are
  rebuilt from the filtered set; `n` is the filtered count, `total` the unfiltered
  DB count.
- **Collapse then open**: opening an article re-uses `articleState` keyed by ID;
  collapse state is unaffected and in-memory only.
- **Windowing**: `listWindow` operates on `len(visibleRows)` and `cursor` instead
  of `len(articles)`.

## Tests

- `longDate`: ordinal correctness (1st, 2nd, 3rd, 11th–13th, 21st) and format.
- `loadList`/grouping: articles on three days → three groups, desc; undated →
  trailing `Undated` group.
- `visibleRows`: collapsed group emits only its header; expanded emits header +
  articles.
- Cursor wrap across visible rows; fold keys on header collapse/expand/toggle;
  Tab from article row toggles its group; `l`/Enter on header do NOT open.
- Status bar renders `n/total`; filter shows `n` < `total` plus filter name.
- `Store.ArticleCount` matches `ListArticles("")` length.

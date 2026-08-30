## Context

See `proposal.md` — Why for motivation. Current state (from code):

- `renderTagPopup` renders one tag per line in a single column, sized as a
  percentage of the screen (80% on the list view, 40% on the article view),
  so a large tag set runs past the bottom of the box.
- Cursor is a single linear `cursor` index into `popupState.tags`, moved with
  `movePopupCursor` → `wrapIndex(cursor, ±1, n)`.
- `updatePopup` dispatches only `Quit`, `Back`, `MoveUp`, `MoveDown`, `OpenURL`
  for popups; the popup view keymap (`config.ViewPopup`) binds no left/right
  keys and does not include `Top`/`Bottom`.
- Tags are already ordered by popularity (`store.ListTags`, total descending),
  so grid order falls out of the existing slice order.

## Goals / Non-Goals

**Goals:**
- Explorer-style column-major grid for both tag popups, all tags visible.
- Four-direction wrap-around navigation + first/last jumps.
- Full-screen geometry via fixed margins (12 columns left/right, 2 rows top/bottom).
- Keymap-consistent: keys stay remappable catalog actions, not hardcoded.

**Non-Goals:**
- Mouse navigation inside the popup.
- Scrolling/virtualization of the tag grid (truncation keeps everything visible).
- Any change to the article links popup (stays single-column, linear).

## Decisions

### D1. Geometry: fixed margins replace percentage sizing

Tag popup box becomes `h = max(3, m.height-4)`, `w = max(12, m.width-24)`,
applied to both list-view and article-view tag popups. The existing
min-size clamps are kept so a tiny terminal yields a minimal usable box
instead of a negative/zero-size popup.

**Alternatives:** keep 80%/40% sizing — rejected, the request is explicit
full-screen-minus-margins; two different sizes no longer make sense once the
grid and navigation are shared.

### D2. Grid model: column-major fill with a linear cursor

With interior grid rows `R = max(1, h-3)` (box height minus two border rows
minus the title line) and `n` tags, columns `C = ceil(n/R)`. The `c`-th
column holds `min(R, n - c*R)` tags. Index mapping is `col = i / R`,
`row = i % R`, so the cursor stays a plain index into `Tags` and
`confirmTagSelection` is untouched.

**Vertical movement is `wrapIndex(cursor, ±1, n)`** — in column-major
enumeration the next index is exactly "down one row, or the top of the next
column when at a column's bottom", and the existing helper already wraps the
global ends (last tag ↔ first tag). This is the Windows Explorer list-view
linear order the user cited.

**Horizontal movement** converts to `(col,row)`, moves `col ± 1` modulo `C`,
and re-derives the index with `row` clamped to the destination column's row
count (`min(row, rowsInCol-1)`), which handles the shorter final column. With
`C == 1` left/right are identity moves (no-op).

**Top/Bottom:** `1`/`g` sets cursor `0`, `G` sets cursor `n-1`.

### D3. Per-column widths capped at 23, middle elision, horizontal scrolling

Each column is sized to fit its own widest entry (2-column cursor gutter +
name + space + counts), capped at 23 columns — not a single shared width — so
one long tag name does not widen every column and collapse the viewport to a
single column. A name wider than its column is middle-elided
(`elideMiddle`, `internal/ui/width.go`): the start and end stay readable and
the middle is replaced with `glyphs().ellipsis` (`…` / `...`), with the gutter
and counts kept intact. Multiple columns show side by side whenever they fit.

When the natural grid is narrower than the interior, the spare space widens the
columns rather than becoming a trailing gap: all columns grow to one equal
width that fills the interior (`equalW = (textW - gaps) / cols`); when a single
long entry forces a column wider than that equal width, every column instead
grows by an equal share of the spare. Any leftover that cannot be split into
whole columns (spare mod cols) is removed by pulling the popup's right border
in, so the grid exactly fills the interior with no space to spare. When the
grid is wider than the interior, the same fill is applied to the visible
horizontal window at render time, so the window also spans the interior with no
trailing space. `tagGrid` returns the effective box width so the renderer draws
borders at the shrunk size.

When the full grid is wider than the popup interior, the popup shows a
horizontal window of the columns that fit and scrolls it to keep the cursor
visible. Column left edges are tracked with a prefix-sum of
`colWidths[c] + tagColGap` (two-space gap).

The window is anchored by a stored scroll column (`popupState.scrollCol`) so
horizontal movement is symmetric: it advances right only when the cursor passes
the window's right edge and left only when it passes the window's left edge,
staying put while the cursor moves within it. This replaces the earlier
cursor-derived window, which picked the leftmost column from which the cursor
fits and therefore scrolled left eagerly on every left move. The stored column
is recomputed at render time (and written back) so wraps, jumps, and resizes
still snap the window to show the cursor; the cursor itself stays a plain
linear index and `moveTagCursorHorizontal` needs no awareness of the window.

**Alternatives:** share one column width sized to the widest entry — rejected
by the user; a single long tag collapses the viewport to one column. Render
long entries in full with no cap — rejected by the user, who asked for a
23-column cap with middle elision. Keep the cursor-derived follow window —
rejected by the user, who wanted the window to scroll left only when the
cursor passes its left edge, symmetric with the rightward behavior.

### D5. Source-domain tags and the news-org column

Every article is associated with a source-domain tag at ingest
(`internal/feed/fetch.go` calls `SetArticleTagsWithSource(id, tags, source)`),
marked with a `source` flag on the `categories` table (added by an idempotent
migration, like the image-column migrations). `store.SourceLabel` (moved to
`internal/store`, reusing the `publicsuffix` derivation previously in
`ui.sourceID`) yields the registrable organization label — `www.nytimes.com`
and `a.lot.of.subdomains.nytimes.com` both give `nytimes`. `ListTags` orders
source tags ahead of category tags (each group by popularity, with
equal-popularity tags tied alphabetically via `COLLATE NOCASE`), so the tag
popup shows news organizations first. The same label is reused for the list
view's right-aligned news-org column, now middle-elided to 15 columns instead
of truncated to 12.

**Alternatives:** synthesize domain tags at query time in `ListTags` with a
separate filter path — rejected, storing them as marked categories reuses the
counts, filtering, source-first ordering, and article-view popup machinery
unchanged.

### D6. Tag popup presentation

The popup title is inlined into the top border (`inlineTitleBorder`, the same
helper the help popup uses) instead of occupying a body line, freeing a grid
row, and the bottom border carries a right-aligned position indicator
(`tagBottomBorder`): `<percent>% · <nth>/<total>` with the selected tag's
place among all tags, mirroring the article view's bottom border and the list
view's status bar. The bottom border line and its bullet render in the same
chrome role as the rest of the popup border (the top and vertical borders),
so the frame reads as one color. The box is built manually (as the help popup is) with one space
of left/right padding and no top/bottom padding, so the grid fills the whole
interior between the borders. Grid cells carry no cursor gutter: the tag name
sits
one column from the popup's content edge, cells are padded to the column width
(the widest entry, capped at 23) and joined with a single space, so adjacent
columns are always separated by exactly one space. Within a cell the tag name
sits on the left and the unread/total counts right-align to the column edge,
with any leftover space padding between them. Selection mirrors the list
view: the selected cell gets the full `#707070` background highlight across
the column width, starting exactly at the tag, with no `>` marker. The
background is applied per styled segment (name, spacer, counts, padding),
exactly as the list rows do — a single outer wrap's reset would be cleared by
the count style's inner `\x1b[0m`, leaving the trailing padding unhighlighted
and shrinking the highlight to the item's own text length. The count
display uses the list view's read/unread colors — the unread number bold
bright, the total in the text (white) role — instead of the dim role.

**Alternatives:** keep the `>` cursor marker and a two-space gutter — rejected
by the user, who asked for the list-view selection style, a tag one space from
the edge, and two spaces between columns.

### D4. Keymap: extend Top/Bottom to the popup view; add MoveLeft/MoveRight with hjkl

- `config/keys.go`: add `ViewPopup` to `Top` and `Bottom` (defaults
  `1`/`g`/`ctrl+up` and `G`/`ctrl+down`), and add new catalog actions
  `MoveLeft` (`left`, `h`) and `MoveRight` (`right`, `l`), valid only in
  `ViewPopup`.
- `h` is currently bound to `Back` in the popup view (`esc`/`h`/`b`). To keep
  vim left/right working, `EffectiveKeys` skips `h` for `Back` in `ViewPopup`
  (mirroring the existing `enter` skip for `Back` in the list view), so the
  popup view resolves `h` → `MoveLeft`, `l` → `MoveRight`, and `esc`/`b` →
  `Back`. The help popup continues to show `Back` as `esc, h, b` (it renders
  the catalog keys, not per-view exclusions — the same quirk that already
  applies to the `enter` exclusion).
- `updatePopup` dispatches `MoveLeft`/`MoveRight` to the grid movement and
  `Top`/`Bottom` to first/last for the **tag popups only**; in the links popup
  they are no-ops, preserving its single-column linear navigation.

**Alternatives:** keep left/right as arrows only — rejected by the user, `hjkl`
must mirror the arrows. Hardcode the keys in `updatePopup` — rejected, breaks
the keymap's remapability and the help popup's catalog-driven rendering. New
actions appear under the help popup's Global section automatically (they are
valid in neither the list nor article view, so `GlobalActions` includes them);
the help requirement already says it lists *each* action, so no help spec delta
is required.

### D7. Sources/tags section subtitles and title

The popup title becomes `Tags & Sources` on both the list and article views.
Each group gets a non-selectable section subtitle: `Sources` above the
source-domain tags and `Tags` above the category tags, each rendered only when
its group is non-empty, in the status-bar blue role (non-bold), occupying a
single in-column grid cell centered within its column.

The subtitles are implemented as a display item list (`tagItems`) built from
the tag slice in order, inserting one subtitle cell before each group. The
column-major grid geometry, column widths, horizontal window, and cell
rendering all operate on this item list; the selection cursor stays a plain
index into `Tags`, so `confirmTagSelection` and the position indicator are
untouched. Vertical navigation is unchanged — `wrapIndex(cursor, ±1, n)`
still yields the next tag in column-major order because the tags appear in
order in the item list and subtitles only sit at group boundaries. Horizontal
movement maps the cursor to its item-grid cell, moves the column, and re-derives
the tag via `tagAtColumnRow`, which skips subtitle cells (landing on the
nearest tag, preferring the one below) and clamps into a shorter final column
exactly as before. Subtitles are never highlighted and navigation never lands
on them.

**Alternatives:** full-width section rows spanning the visible grid — rejected
by the user, who chose single in-column header labels. Keeping a flat tag-only
grid and prefixing each group's first cell with its label — rejected, that
would hide a tag; an inserted subtitle cell is the only layout that keeps every
tag visible.

### D8. Rendering

`renderTagPopup` builds row strings by walking `row` from 0 to `R-1` and
concatenating each column's cell for that row (padded to `colW`, gap 1),
omitting cells past a partial column's count. The title line and count
formatting (`tagCountText`) are unchanged. The links popup renderer is
untouched.

### D9. Case-insensitive, case-preserving tags

Tag names are stored uniquely case-insensitively but preserve the first
spelling seen forever. The `categories.name` column is declared
`TEXT COLLATE NOCASE UNIQUE`, so SQLite's unique constraint (and the
`INSERT ... ON CONFLICT(name)` conflict target, and `WHERE c.name = ?`
comparisons in filtering) all resolve tags case-insensitively, while the
existing row's spelling is kept on conflict. Within one batch,
`setArticleTags` deduplicates names on `strings.ToLower` so an article never
links the same tag twice, matches the source label with `strings.EqualFold`
so a case-variant category still becomes the source tag, and only promotes the
source flag (`MAX(source, excluded.source)`) so a tag marked as a news
organization stays one. Because `CREATE TABLE IF NOT EXISTS` does not alter an
existing table, databases created before this change keep case-sensitive tag
names; the new `-z`/`--init-db` flag deletes the database (and WAL/SHM
sidecars) so a fresh store picks up the NOCASE schema — the stored data is
transient and may be wiped freely during development.

**Alternatives:** a normalized `lower_name` column with a case-insensitive
unique index — rejected, the `COLLATE NOCASE` column collation is simpler and
makes filtering and conflict resolution automatic. A table-rebuild migration to
dedupe existing case-variant tags — rejected by the user, who preferred the
`-z` reset flag over a dedup migration.

### D10. Database reinit flag

`yerss -z` / `--init-db` deletes the database file and its `-wal`/`-shm`
sidecars before `store.Open`, so the next open starts from zero. `main` prints
a notice naming the deleted path. This is the development escape hatch that
makes the case-insensitive schema change safe without a dedup migration.

**Alternatives:** SQL statements to `DROP TABLE` everything and recreate —
rejected, deleting the file (and sidecars) is simpler and also clears WAL
artifacts.

## Risks / Trade-offs

- [Very wide grids need many horizontal scroll steps to reach the least
  popular tags] → per-column widths (capped at 23) keep most columns narrow, so
  multiple columns stay visible; the follow-cursor window advances one column
  per step and `G` jumps straight to the last column.
- [Elided tag names and domain labels lose their middles] → start and end are
  kept, so tags/domains stay recognizable; this is the requested trade-off.
- [Source-domain tags duplicate a category name or are visible in the
  article-view popup] → deduplicated at ingest; showing the source in the
  article popup is consistent with treating sources as tags.
- [Existing articles lack source tags until a refresh re-tags them] → refresh
  re-associates tags on every ingest, so one refresh backfills the domain tags.
- [Help popup Global section gains rows for Move Left / Move Right] → adds
  ~10 columns of width; well under the 70-column cap.
- [Back's `h` binding is silently inactive in the popup view] → the help popup
  still lists `esc, h, b` for Back, mirroring the existing `enter` exclusion
  quirk; `b` and `esc` still close the popup.
- [Down-wrap interpretation: linear column-major continuation rather than
  column-local wrap] → matches Explorer list-view behavior and the request's
  "wrap at the top/bottom/left/right of the screen"; adjustable if the user
  prefers column-local wrap.
- [Behavior change for existing tag-popup navigation] → purely additive grid;
  up/down wrap semantics at the global ends are preserved.

## Migration Plan

Pure UI change — no data or config migration. Rollback is a revert of the
implementation commit.

## Open Questions

None.
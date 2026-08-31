## Why

The tag popup renders every tag in one long column that runs past the bottom of
the screen ("continues down forever and ever"), so navigating a large tag set
means holding `j` and watching the cursor march down a single column. An
Explorer-style multi-column grid with four-direction navigation makes the whole
tag set visible and reachable at a glance.

## What Changes

- Tag popups (both list-view and article-view) render their tags in a
  Windows Explorer-style multi-column grid: entries are ordered by popularity
  (total article count descending) and fill each column top to bottom,
  continuing into the next column when the bottom is reached.
- Tag popup geometry changes from percentage sizing (80% of screen on the list
  view, 40% on the article view) to fixed margins: the popup covers the screen
  with a 12-column margin on the left and right and a 2-row margin at the top
  and bottom.
- Navigation becomes four-directional: up/down follow the column-major order
  and wrap around at both ends (last tag → first tag), left/right move between
  columns and wrap around at the grid's left/right edges, clamping the row when
  entering a shorter final column.
- `1`/`g` move to the first tag and `G` moves to the last tag inside the popup
  (the existing Top/Bottom actions become valid in the popup view).
- Tag popup columns are per-column sized with a 23-column cap; a name wider
  than its column has its middle elided with an ellipsis, keeping start and
  end readable. The grid may be wider than the screen, so the popup scrolls
  horizontally as the cursor moves toward the least popular tags.
- `h`/`l` move left/right in the popup, mirroring the arrows.
- Source-domain tags are marked in the database and grouped first in the
  popup, ahead of category tags.
- The popup title is inlined in the top border, the popup keeps a single space
  of padding on every side and between columns, selection uses the list view's
  background highlight (no `>` marker), and the tag counts use the list view's
  read/unread colors.
- The popup title becomes `Tags & Sources`, and the source-domain tags and the
  category tags each get a non-selectable in-column section subtitle —
  `Sources` and `Tags` respectively — rendered in the non-bold status-bar blue
  role above their group.
- The list view's news-org column shows the full registrable domain label,
  middle-elided to 15 columns (was: truncated to 12).
- Every article is associated with a source-domain tag (the registrable
  organization label), so news sources appear in the tag popup and filter the
  list like category tags.
- The article links popup keeps its existing single-column, linear navigation;
  the new popup-view keys are no-ops there.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `reader-ui`: The "Article list rendering" requirement changes — the news-org
  column derives the full registrable domain label and middle-elides it to 15
  columns. The "Tag popup on list view" and "Tag popup on article view"
  requirements change — popup geometry, multi-column grid layout with a
  23-column cap and middle elision, horizontal scrolling, four-direction
  wrap-around navigation with first/last jump keys, and source-domain tags
  listed alongside category tags.
- `feed-pipeline`: The "Automatic tagging from feed categories" requirement
  changes — every article also gets a source-domain tag derived from its link
  or feed host.

## Impact

- `internal/ui/popup.go` — `renderTagPopup` (grid layout + new geometry +
  elision + section subtitles), `updatePopup` (new actions), grid movement
  helpers, `tagItems` (subtitle item list).
- `internal/ui/width.go` — `elideMiddle` helper.
- `internal/ui/list.go` — `sourceID` delegates to `store.SourceLabel`; the
  row render middle-elides the news-org column to 15.
- `internal/store/store.go` — `SourceLabel` (registrable label derivation).
- `internal/feed/fetch.go` — ingest appends the source-domain tag.
- `internal/config/keys.go` — extend `Top`/`Bottom` to the popup view; add
  `MoveLeft`/`MoveRight` (`left`/`h`, `right`/`l`); exclude `h` from `Back`
  in the popup view.
- Tests: `internal/ui` (tags, render, links-popup regression, polish),
  `internal/config` (keymap catalog), `internal/store` (SourceLabel, source
  tags), `internal/feed` (ingest).
- The help popup's Global section will gain `Move Left` / `Move Right` rows
  because they are catalog actions; the help-popup requirement itself is
  unchanged (its scenario enumerations are illustrative, not exhaustive).

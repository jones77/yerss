## Why

`internal/ui` is ~34 files in one package, and `Model` still orchestrates list rendering, article composition, image views, and popups directly. `render/` and `compose/` already demonstrate the house style (pure, well-tested sub-packages), and `extract-image-view-state` will have carved the image concern into an `imageView` type. This change is the package split the `introduce-app-service-layer` design explicitly deferred: move each view concern into a focused sub-package so `Model` becomes a thin wiring layer.

## What Changes

- Extract four sub-packages from `internal/ui`, in dependency order:
  - `internal/ui/listview` — `listState` + list rendering and list input (`list.go`, `list_update.go`).
  - `internal/ui/popup` — `popupState` + tag/links/help popups (`popup.go`, `popup_help.go`).
  - `internal/ui/imageview` — the `imageView` type from `extract-image-view-state` (`image.go`).
  - `internal/ui/articleview` — `articleState` + composition, article rendering, and selection (`article.go`, `article_update.go`, `article_selection.go`).
- Move the remaining pure helpers in the top-level package (`overlay`, `spliceStyled`, `formatMB`, glyph/status-bar pure parts) into `render`/`compose` where they belong.
- `Model` (and its message routing in `Update`) becomes the composition root that constructs and delegates to the four views.
- No behavior change; every view keeps its existing tests (moved alongside).

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/` shrinks to `model.go` (wiring) plus the moved-in pure helpers.
- New packages: `listview`, `popup`, `imageview`, `articleview`.
- `render/`, `compose/`: absorb the last pure helpers from the top-level package.
- Test files move with their subject; package-qualified type references (`listState` → `listview.State`, etc.) are updated across the tree.

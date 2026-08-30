## Why

The help popup lists all 18 actions in one flat column, so it is hard to tell
which keys work everywhere versus which belong to a specific view. Grouping the
bindings by scope makes the help legible at a glance.

## What Changes

- `renderHelp` groups actions under three headings, in catalog order within
  each section:
  - **Global** — actions not tied to one view: all-view bindings (Quit, Back,
    Move Up, Move Down, Help) plus the navigation shared by the list and
    article views (Page Up, Page Down, Half Page Up, Half Page Down, Top,
    Bottom, Tag Popup).
  - **List view** — list-only bindings: Refresh, Open Article.
  - **Article view** — article-only bindings: Link Popup, Open URL, Copy URL,
    Copy Article Text.
- The List view and Article view sections render in a second column to the
  right of Global, and the popup including its border stays within 70 columns.
- Section headings render in the blue role; each action row keeps its current
  label/key layout.
- Grouping is derived from each action's declared views in the keybinding
  catalog, so it cannot drift from the bindings themselves.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: The "Help popup" requirement changes so the popup presents the
  bindings grouped into Global, List view, and Article view sections instead of
  a single flat list.

## Impact

- `internal/config/keys.go` — exported helpers exposing the three action
  groups derived from the catalog.
- `internal/ui/popup.go` — `renderHelp` renders the grouped sections.
- `internal/ui/keys_test.go` — `TestHelpListsEveryActionInOrder` must accept
  grouped (non-catalog) ordering.
- Spec: `openspec/specs/reader-ui/spec.md` (delta).
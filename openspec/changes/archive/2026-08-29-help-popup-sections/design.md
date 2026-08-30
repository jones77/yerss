## Context

See `proposal.md` — Why. `renderHelp` (`internal/ui/popup.go`) iterates
`config.AllActions()` and prints one flat column. The keybinding catalog
(`internal/config/keys.go:56`) already declares which views each action is
valid in, so the grouping can be derived rather than hardcoded.

## Goals / Non-Goals

**Goals:**
- Three help sections: Global, List view, Article view.
- Grouping derived from the catalog so it cannot drift from the bindings.

**Non-Goals:**
- No change to keybindings, dispatch, or conflict validation.
- No change to action labels or per-row layout.

## Decisions

### 1. Group derivation lives in the config package

Add exported helpers in `internal/config/keys.go` that classify each catalog
entry by its declared views:

- `GlobalActions()` — actions not specific to one view: valid in all views
  (Quit, Back, Move Up, Move Down, Help) plus those valid in both the list and
  article views (Page Up, Page Down, Half Page, Top, Bottom, Tag Popup). In
  catalog order.
- `ListActions()` — valid in the list view and not the article view (Refresh,
  Open Article).
- `ArticleActions()` — valid in the article view and not the list view (Link
  Popup, Open URL, Copy URL, Copy Article Text).

Classification predicate: list-only = `in List && !in Article`; article-only =
`in Article && !in List`; everything else is Global. This matches the user's
choice to fold the shared list/article navigation into Global.

Alternative considered: a single `HelpGroups()` struct. Rejected — three
simple accessors match the existing `AllActions`/`actionsForView` style and
are easier to test.

### 2. renderHelp renders two columns capped at 70 columns

`renderHelp` builds `Help - current keymap` over a two-column body. The Global
section fills the left column; the List view and Article view sections stack in
the right column, separated by a blank line. Each column's rows are padded to
the column's widest row, and the columns are joined with two spaces. Headings
render bold in the blue role (`palette.StatusBar`, ANSI 12/4), matching the
popup title and border. Action rows use a tighter `  %-11s %s` layout so both
columns fit side by side.

The box width is `min(70, m.width, contentW+4)` where `contentW` is the joined
two-column width; the box height tracks the content. Lines are padded and
truncated to the interior width so the popup, border included, never exceeds
70 columns even with longer remapped keys.

### 3. Tests

`TestHelpListsEveryActionInOrder` (keys_test.go) asserted strict catalog order,
which grouping breaks. It becomes: every action line is present and the three
headings appear in Global → List view → Article view order. A new config test
pins the exact membership of the three groups (all 18 actions, no overlap, no
gaps). `TestHelpListsLinkPopup` continues to pass unchanged.

## Risks / Trade-offs

- [The Global section is the largest (12 rows)] → It is genuinely the set of
  bindings that apply beyond one view; the smaller List/Article sections are
  what the user scans for view-specific keys.
- [Growing the popup by ~6 lines on very short terminals] → Height is a floor,
  and the popup centers; acceptable and pre-existing behavior for tall content.

## Migration Plan

None — code-only change plus the `reader-ui` spec delta. Rollback is a revert
of `config/keys.go`, `popup.go`, the affected tests, and the spec.

## Open Questions

None.
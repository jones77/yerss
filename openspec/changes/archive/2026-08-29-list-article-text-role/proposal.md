## Why

In the list view, article rows render their text in the grey role (ANSI 8 dark
grey): read titles, timestamps, and source identifiers. On many terminals that
grey is too dark to read comfortably against the article body text, which is
ANSI 7 white. Article rows should use the text role (white), matching the
article body, with grey reserved for chrome (borders, rails, day headers).

## What Changes

- In `renderArticleRow`, read article titles, publication times, and source
  identifiers render in the text role (ANSI 7 white in dark mode, ANSI 0 black
  in light mode) instead of the grey role.
- Unread titles stay in the bright role (ANSI 15 / bold black) — unchanged.
- Day headers and all other grey-role chrome are unchanged.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: The "Standard terminal color roles" and "Article list
  rendering" requirements change so list article rows (read titles,
  timestamps, source identifiers) use the text role rather than the grey role;
  the grey role is limited to chrome and day headers.

## Impact

- `internal/ui/list.go` — `renderArticleRow` switches its muted/read styling
  from `palette.Dim` to `palette.Text`.
- `internal/ui/polish_test.go` and `internal/ui/daygroups_test.go` — assertions
  on the article-row text color.
- Spec: `openspec/specs/reader-ui/spec.md` (delta).
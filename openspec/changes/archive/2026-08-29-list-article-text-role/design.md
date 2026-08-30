## Context

See `proposal.md` — Why. In `renderArticleRow` (`internal/ui/list.go`), read
titles, the publication time, and the source identifier all use
`m.palette.Dim` (ANSI 8 dark grey). Unread titles use `m.palette.Bright`
(ANSI 15) with bold weight, and day headers use `m.palette.Dim`. The article
body text uses the text role (`m.palette.Text`, ANSI 7 white).

## Goals / Non-Goals

**Goals:**
- Article rows in the list render their content in the text role (white),
  matching the article body.
- Unread titles keep the bright role; day headers and chrome keep grey.

**Non-Goals:**
- No change to the article reader border, body rendering, popups, or the
  `#333333` selection background.
- No new glyphs or layout changes.

## Decisions

### 1. Article-row muted/read styling moves to the text role

In `renderArticleRow`, the `dim` style that renders the time and source, and
the read-title style, switch from `m.palette.Dim` to `m.palette.Text`. The
style variable is renamed `dim` → `text` for accuracy. Unread titles keep
`m.palette.Bright` + bold, so read/unread still differ (white vs bright white
bold). Day headers in `renderDayHeader` keep `m.palette.Dim`.

Alternative considered: keep time and source grey and only whiten the read
title. Rejected: the user asked for the article view in color 7 wholesale, and
a half-grey row would reintroduce the readability issue.

### 2. Tests

`TestArticleRowTitleColorChangesWithRead` (polish_test.go) builds expected
escapes from `m.palette.Dim`/`m.palette.Bright` and asserts the row uses them.
It must build the read/time/source expectations from `m.palette.Text` instead.
`TestArticleRowShowsLocalTime` and `TestArticleRowSelectionBackground` build
expectations from the same palette fields, so they keep passing once
`renderArticleRow` uses `Text`.

## Risks / Trade-offs

- [Read and unread titles both read as white on some terminals] → Unread is
  additionally bold and bright-white, preserving the distinction.
- [Time and source lose the muted look] → Intended; the whole row now matches
  the article body color.

## Migration Plan

None — code-only change plus the `reader-ui` spec delta. Rollback is a revert
of `internal/ui/list.go`, the affected tests, and the spec.

## Open Questions

None.
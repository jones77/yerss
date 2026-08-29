## Context

`WrapText` was the article path's word-wrapper before `glamour-article-rendering`
moved wrapping into glamour's `WithWordWrap`. See proposal.md - Why. The function
and its helpers are now referenced only by `convert/wrap_test.go`; the article
pipeline (`internal/ui/article.go` → `renderMarkdown`) no longer touches them.

## Goals / Non-Goals

**Goals:**
- Remove `WrapText` and every helper that exists only to serve it, in one
  self-contained deletion.
- Leave `Convert` (the remaining `convert` export) and its tests untouched.

**Non-Goals:**
- No change to the article rendering pipeline, reader UI, or any behavior.
- No consolidation of ANSI-width helpers that are still in active use (that is
  a separate change: `refactor-article-rendering`).

## Decisions

**D1. Delete the file wholesale rather than delete functions one at a time.**
`wrap.go` contains only `WrapText` and its private helpers (`linkTokens`,
`truncateURL`, `isTrailingPunct`, `linkOpenSeq`); nothing else in the package
imports them. Removing the whole file is less churn than a function-by-function
edit and leaves no half-used helpers behind. Alternative considered: keep
`WrapText` but mark it deprecated — rejected, because dead code with tests is
worse than dead code without them.

**D2. Remove the tests together with the code.**
`wrap_test.go` exists solely to assert `WrapText`'s behavior. Once the function
is gone the tests cannot compile; deleting them in the same change keeps the
tree green without a stop-the-world intermediate commit.

**D3. Verify imports after deletion.**
`wrap.go` is currently the only `convert` file importing `unicode` and
`charmbracelet/x/ansi` (for `ansi.StringWidth`/`ansi.Truncate`/`ansi.ResetHyperlink`).
After deletion, drop any imports `go build`/`go vet` report as unused.

## Risks / Trade-offs

- **A future feature might have wanted `WrapText`** → If a caller reappears, the
  code is recoverable from git history; keeping 400 lines of untested-in-prod
  code indefinitely is the greater cost.
- **Import removal breaks a still-used symbol** → `go build ./...` and
  `go vet ./...` in the task list catch this mechanically.

## Migration Plan

None — pure source deletion, no data or config impact. Rollback is a revert.

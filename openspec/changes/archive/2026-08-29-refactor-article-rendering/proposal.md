## Why

The article reader's markdown→terminal step works, but its correctness rests on
two post-processors that re-parse glamour's ANSI-styled output —
`indentListContinuations` and `fixBlockquoteRewrap` — plus hand-rolled
ANSI/width helpers and duplicated selection/header logic in
`internal/ui/markdown.go` and `internal/ui/article.go`. These patches are
fragile: they operate on the *rendered* string rather than the markdown AST, so
any glamour version bump can silently break their regexes/heuristics, and their
width bookkeeping is coordinated by convention across several files
(`contentGeom`, glamour's `WithWordWrap`, the border's `truncate`/`padRight`,
`indentListContinuations`'s `+2`, and `fixBlockquoteRewrap`'s `-2`).

This is the capstone change. It is applied **last**, after
`remove-dead-wrap-text` and `harden-render-fallbacks`, so its opening research
step starts from the cleaned-up codebase. It first investigates how glamour
actually renders lists and blockquotes, then reworks the pipeline to be
declarative rather than post-hoc string surgery, and finally consolidates the
duplicated helpers.

## What Changes

- **Research / re-assessment step (explicit, first)**: before any rework, trace
  glamour's raw output for wrapped list items and blockquotes at several widths
  to confirm whether `indentListContinuations` clips continuation text, whether
  the `fixBlockquoteRewrap` orphan heuristic is sound, and whether glamour
  offers configuration to do both natively. Record findings and pick the
  declarative approach (see design.md — Decision D0).
- **Declarative rendering**: replace or shrink the two post-processors using the
  research outcome — (a) glamour style/word-wrap configuration, (b)
  markdown-source normalization before glamour, or (c) a consolidated
  post-processing pass as a fallback — so list continuations hang-indent and
  blockquotes keep a solid `│` bar without re-parsing styled output.
- **Consolidate duplicated helpers**: merge `splitBar`/`cutStyledWidth` into one
  "advance N visible cells" helper; unify the escape-sequence walkers
  (`skipEscape`, `parseLinkSpans`'s inline skip, `oscEnd`); extract the
  per-row selection-range ladder shared by `selectedText` and
  `highlightSelection`; render the article header once; and drop
  `articleState.lines` in favor of `viewport.View()`.
- **Unify width measurement**: standardize the article rendering path on one
  grapheme-aware width primitive (the `charmbracelet/x/ansi` helpers already
  imported) instead of mixing `runewidth.RuneWidth` (per-rune) and
  `ansi.StringWidth` (grapheme-aware).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

_None._ The target is the same visible rendering the `reader-ui` spec already
describes (correct list indentation, a solid blockquote bar, correct wrapping).
Any discrepancy the research uncovers — e.g. clipped continuation text — is a
bug fix that brings output *into* compliance with existing requirements, not a
new requirement, so this change declares `skip_specs: true`.

## Impact

- **UI layer**: `internal/ui` — `markdown.go`, `article.go`, `image.go`, and
  possibly `width.go`/`geometry.go` if the width unification touches shared
  helpers. No public API changes.
- **Tests**: existing `internal/ui/markdown_test.go` (blockquote/list/heading),
  `osc8_test.go` (link spans), mouse-selection, and render-safety tests are the
  safety net. The research step adds golden-style traces or focused assertions
  pinning the current (correct) output so the rework provably preserves it.
- **Sequencing**: applied last; depends on the two prior changes having landed
  so the research begins from a codebase with no dead wrapper and clean
  fallbacks.

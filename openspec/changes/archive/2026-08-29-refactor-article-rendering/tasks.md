## 1. Research and re-assessment (run first, after remove-dead-wrap-text and harden-render-fallbacks)

- [x] 1.1 Trace glamour's raw list output at several widths; confirm whether `indentListContinuations` clips the last 2 columns of continuation lines or glamour already wraps content at `contentW-2`
- [x] 1.2 Trace glamour's raw blockquote output at several widths; confirm the orphan heuristic, whether consecutive orphan lines occur, and whether `contentW-2` is the correct re-wrap width
- [x] 1.3 Investigate glamour style/word-wrap options for list-continuation hang-indent and correct blockquote wrapping
- [x] 1.4 Evaluate markdown-source normalization (pre-indent list continuations, pre-wrap blockquote paragraphs) as an alternative to post-processing
- [x] 1.5 Compare `runewidth.RuneWidth` vs `ansi.StringWidth` on combining/wide/ZWJ glyphs in the article path
- [x] 1.6 Record golden traces (or focused assertions) of the current correct list/blockquote output to hold the rework to
- [x] 1.7 Record findings, pick the declarative path (D0), and update this change's design.md accordingly

## 2. Consolidate helpers (dedup)

- [x] 2.1 Extract `styledCells(s string, n int) int` and express `splitBar` and `cutStyledWidth` in terms of it
- [x] 2.2 Consolidate `skipEscape` and `oscEnd` into one escape-skipping primitive, and reuse it in `parseLinkSpans`
- [x] 2.3 Extract `selectionRange(lo, hi cell, row, lineWidth int) (from, to int)` and use it in `selectedText` and `highlightSelection`
- [x] 2.4 Pass the computed header line count into `articleImageBlock` instead of re-rendering the header there
- [x] 2.5 Attempt to drop `articleState.lines` and derive `CopyArticleText` from `m.article.viewport.View()` — reverted: `viewport.View()` returns only the visible window and the spec requires copying the full article text (see design.md D5)
- [x] 2.6 Run `go test ./internal/ui/...` and confirm osc8/markdown/mouse-selection tests still pass

## 3. Declarative rendering (per research decision D0)

- [x] 3.1 Implement the chosen approach: glamour configuration, markdown-source normalization, or a consolidated post-processing pass
- [x] 3.2 Remove or shrink `indentListContinuations` and `fixBlockquoteRewrap` accordingly
- [x] 3.3 Confirm the rework reproduces the golden traces from 1.6 (list hang-indent and solid blockquote bar intact)
- [x] 3.4 Add/refresh regression tests for lists and blockquotes at multiple widths

## 4. Width-measurement unification

- [x] 4.1 Standardize the article path on the grapheme-aware `charmbracelet/x/ansi` width primitive, replacing per-rune `runewidth.RuneWidth` uses
- [x] 4.2 Add wide-char and combining-char regression tests for the blockquote/list paths

## 5. Verify

- [x] 5.1 Run `go build ./...` and `go vet ./...` and confirm a clean result
- [x] 5.2 Run the full `go test ./...` suite and confirm no regressions

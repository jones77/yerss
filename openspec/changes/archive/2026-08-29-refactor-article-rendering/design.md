## Context

Glamour owns wrapping and styling; `renderMarkdown` then applies two
post-processors — `indentListContinuations` and `fixBlockquoteRewrap` — that
re-parse glamour's ANSI-styled output. Alongside them sit helper functions that
walk ANSI sequences and visible cells, and the mouse-selection code that maps
selection cells to text. This change runs **last**, after
`remove-dead-wrap-text` and `harden-render-fallbacks`, so the research step below
begins from the cleaned-up codebase. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- Determine, from evidence, the most declarative way to render lists and
  blockquotes correctly, then implement it.
- One styled-width helper and one escape-skipping primitive, shared by all
  callers.
- One selection-range function shared by `selectedText` and `highlightSelection`.
- One width-measurement primitive for the article path.
- Preserve the current visible rendering for inputs that already render
  correctly.

**Non-Goals:**
- No replacement of glamour with a different renderer.
- No change to the *specified* output (styling, wrap columns, indentation
  widths, blockquote bar) — only the mechanism that produces it.

## Decisions

**D0. Research first, then commit to one declarative path. — RESOLVED: path 3.**
The implementation choice for lists/blockquotes was gated on a research step
(task group 1). The decision framework, in preference order:
1. If glamour can hang-indent list continuations and wrap blockquotes correctly
   via style/word-wrap options → configure it and delete both post-processors.
2. Else if normalizing the markdown source (pre-indenting list continuations,
   pre-wrapping blockquote paragraphs) before glamour yields identical output
   → normalize upstream and delete the post-processors.
3. Else → keep a post-processing pass but make it a single, consolidated,
   well-tested function, and document why glamour cannot do it natively.

Research findings (traced via a temporary test rendering raw glamour output at
widths 30/40/60/80; glamour style source inspected at v2.0.1):
- Glamour already wraps list-item continuations to leave room for the marker
  indent, so `indentListContinuations`'s `+2` never clips content — the border's
  truncation only removes glamour's trailing styled-space padding. The feared
  2-column content loss does not occur.
- glamour exposes only `List.LevelIndent` (per-level nesting depth); there is no
  continuation hang-indent option, so path 1 is unavailable for lists.
- Blockquote bar lines rendered with a solid `│` at every traced width; the
  orphan mis-wrap is a real but narrow glamour wrapping bug no style option can
  reach, so path 1 is unavailable for blockquotes too.
- Path 2 (markdown-source normalization) was rejected: it changes the bytes fed
  to glamour, risking output drift against the golden tests for no functional
  gain over the cleaned post-processing.

Whichever path wins, the visible output is held constant by the regression
tests added in the research step.

**D1. A single `styledCells(s string, n int) int` helper returns the byte offset
covering at most `n` visible cells.**
Both `splitBar` (split after the 2-cell `│ ` bar) and `cutStyledWidth` (trim
styled trailing padding) are the same loop. Extract it, then express each caller
in terms of it. Alternative considered: `ansi.Cut`/`ansi.Truncate` directly —
rejected because the code needs the byte-level cut point, not a rendered string.

**D2. `skipEscape` becomes the single escape-skipper, reused by the link-span
scanner.**
`parseLinkSpans` inlined the same CSI/ESC skip next to its own `oscEnd`.
`parseLinkSpans` now delegates every non-OSC sequence to the shared `skipEscape`;
`oscEnd` remains as the OSC-specific terminator finder because the scanner needs
the terminator *index* to extract the `8;…;url` payload, which a pure skipper
cannot return. `osc8_test.go` span cases are the guard against behavior drift.

**D3. Extract `selectionRange(lo, hi cell, row, lineWidth int) (from, to int)`.**
`selectedText` and `highlightSelection` encode the same first/middle/last-line
ladder. One helper, two callers, so highlight and copy cannot disagree.

**D4. Compute the header once.**
`newArticleState` already renders the full document (header + body). Change
`articleImageBlock` to receive the header line count it needs for the height cap
instead of re-rendering the header, removing a duplicate glamour render per open.

**D5. Drop `articleState.lines`. — REVERSED during implementation.**
The plan was to derive `CopyArticleText` from `m.article.viewport.View()`.
Implementation revealed two blockers: bubbles v1.0.0's `viewport.View()` returns
only the *visible window* (`m.lines[top:bottom]`) and offers no full-content
getter, so the swap would have violated the `reader-ui` requirement that `C`
copies the **full** article text; and the selection/image tests legitimately
need full-content line access. `articleState.lines` is kept as the full-content
source of truth; the header re-render removal (D4) is unaffected.

**D6. One width primitive for the article path.**
`splitBar`/`cutStyledWidth` use per-rune `runewidth.RuneWidth`, which can
miscount grapheme clusters (combining marks, emoji ZWJ sequences), while the
rest of the path uses `ansi.StringWidth`. Standardize on the grapheme-aware
`charmbracelet/x/ansi` helpers, with the decision confirmed by the research
step's wide/combining-char trace.

## Risks / Trade-offs

- **Research finds no native glamour option (path 3)** → The consolidated
  post-processing pass remains inherently output-coupled, but is at least a
  single documented, tested function rather than two ad-hoc patches.
- **`skipEscape` unification changes link-span parsing subtly** → `osc8_test.go`
  span cases pin the exact spans; no assertion may move.
- **Width unification changes edge-case wrapping** (combining/wide chars) →
  This is the intended fix; new wide-char tests lock it in, and any visible
  change is limited to those edge cases, not the common ASCII path.
- **Output drift during rework** → The research step records golden traces of
  the current correct output; the rework must reproduce them byte-for-byte
  (modulo the intended width-unification edge cases).

## Migration Plan

None — internal rework with no data/config/API impact. Because it is sequenced
last, it can be reverted independently of the two prior changes. Rollback is a
revert.

## Open Questions

All five were resolved by the research step (task group 1) and recorded in D0/D6:

1. **Does `indentListContinuations` clip?** No. Traced output shows glamour
   wraps continuations to leave room for the marker indent; the border's
   truncation removes only glamour's trailing styled-space padding.
2. **Is the orphan heuristic sound?** It missed *consecutive* orphan lines (the
   second orphan's neighbor is not a bar line). `isOrphanLine` now recognizes the
   head of a contiguous stranded-word run ending at the next bar line; the
   width threshold still guards against swallowing ordinary paragraph text. The
   `contentW-2` re-wrap width was confirmed correct at all traced widths.
3. **Native glamour options?** No. Only `List.LevelIndent` exists (nesting
   depth), and no style option reaches the blockquote re-wrap bug.
4. **Markdown-source normalization?** Rejected — it risks output drift against
   the golden tests for no functional gain over the cleaned post-processors.
5. **Width-primitive disagreement?** None for single runes: probed wide (2),
   combining (0), box-drawing (1), and emoji (2) glyphs — `ansi.StringWidth` and
   `runewidth.RuneWidth` agree on all. Unification is about one source of truth,
   not a behavior change.

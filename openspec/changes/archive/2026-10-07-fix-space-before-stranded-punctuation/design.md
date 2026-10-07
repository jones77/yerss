## Context

Compact mode (default) strips the visible href from body links and re-wraps the
surrounding paragraph (`compose.StripLinkHrefs` in `internal/ui/compose/links.go`,
`compose.ReflowShortLines`/`reflowParagraph` in `internal/ui/compose/reflow.go`,
invoked from `internal/ui/articleview/article.go`). The failing article (id 10431)
contains `...estimate of <a href="…">$195,869.52</a>, for a grand total…`. Glamour
wraps the long href onto its own line and the markdown comma follows the href on
that same line; stripping the href leaves a line beginning `, for a grand`, which
the re-flow then glues to the name line with a space, producing `$195,869.52 ,`.

## Goals / Non-Goals

**Goals:**
- Guarantee no space is inserted before closing punctuation stranded after a hidden
  URL in compact mode.

**Non-Goals:**
- No change to expanded mode, clickability, or OSC 8 wrapping.

## Decisions

**Fix in the re-flow join, not the strip.** The href drop is correct — the comma is
real text outside the link and must survive. The defect is the re-flow's
unconditional space between joined lines. `reflowParagraph` already suppresses the
join space when a whole line is closing punctuation (`trailingPunctRe`); the fix
extends this to a line that *starts* with closing punctuation
(`leadingClosingPunctRe`), which is exactly what a stranded comma produces.

**Reproduce first, fix second.** Per AGENTS.md the bug is reproduced from the local
SQLite DB (article id 10431) at the reported geometry before changing the re-flow
rules.

## Risks / Trade-offs

- [Suppressing a space a genuine word wrap needs] → the leading-punctuation set is
  limited to characters that never take a preceding space (`,.;:!?…` and closing
  quotes/brackets), so ordinary word-wrapped lines (which start with a word
  character) are unaffected.
- [Re-introducing the missing-space / double-space cases] → keep the existing
  spacing scenarios green while adding the stranded-punctuation scenario.

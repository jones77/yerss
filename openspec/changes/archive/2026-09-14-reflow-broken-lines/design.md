## Context

The article pipeline (internal/ui/article.go `computeDerivation`) renders each
markdown segment with a cached glamour `TermRenderer` at the content width, then
in compact link mode (`!linkExpand`) passes each segment through
`compose.StripLinkHrefs` before lines are split and shown in a bubbles viewport.
Two glamour line-wrapping behaviors surface as visible defects after this
post-processing:

- Blockquotes are wrapped by glamour with breakpoints `" ,.;-+|"`, so a final
  word ending in punctuation (e.g. `on.”`) splits after the period, stranding
  the closing curly quote on its own bare line. `FixBlockquoteRewrap`
  (internal/ui/compose/markdown.go) only folds a stranded line back into the
  quote when it is a "head" orphan whose successor is another quote line.
- Glamour pads every paragraph line to the content width with styled spaces.
  In compact mode `StripLinkHrefs` deletes a wrapped URL's href spans, including
  the newline and padding across the wrap boundary, leaving a short unpadded
  glued line. There is no subsequent re-wrap.

## Goals / Non-Goals

**Goals:**
- Restore full-width paragraph layout in compact mode after link-URL stripping.
- Keep blockquote trailing punctuation on a quote-bar line.
- Preserve inline styling and OSC 8 hyperlink spans through both fixes.
- Fixes work on already-rendered styled text (consistent with the existing
  `FixBlockquoteRewrap` post-processing model).

**Non-Goals:**
- Changing glamour's wrapping algorithm or themes.
- Adding a glamour option to hide href text (verified: glamour v2 has no such
  style option for regular links).
- Altering expanded-mode layout.

## Decisions

**1. Fix blockquote orphan absorption in `isOrphanLine`.**
Extend `isOrphanLine` so a line whose trimmed visible text is punctuation-only
(e.g. `”`) is treated as part of the preceding blockquote when the previous line
is a quote-bar line, regardless of what follows. Rationale: this is the exact
stranded-tail signature glamour produces; the existing "successor must be a bar
line or orphan" rule handles head-orphans but not a trailing orphan before a new
paragraph. Alternative considered: special-casing in `FixBlockquoteRewrap`'s
scan loop — rejected, the predicate is the single seam for both.

**2. Add a compact-mode re-flow pass.**
New compose helper (e.g. `ReflowShortLines(rendered string, contentW int)`)
called in `computeDerivation` after `StripLinkHrefs` when `!m.linkExpand`.
It splits into lines, groups consecutive non-blank lines into paragraphs, skips
blockquote (`│`-prefixed), list-item, and indented groups (re-wrapping would
destroy their structural prefixes), and re-flows only paragraphs containing a
line whose raw `xansi.StringWidth` is below the content width — the exact
artifact signature, since glamour pads every line to the content width.
Re-flow trims each line's trailing styled whitespace (keeping closing SGR/OSC
escapes), joins with single spaces (gluing punctuation-only orphan lines without
a space), and wraps with `charm.land/lipgloss/v2` `Wrap` at the content width
(empty breakpoints). Rationale: re-flowing only the damaged paragraphs keeps the
diff minimal and avoids perturbing hand-wrapped or intentionally short text.
Alternative considered: re-padding the glued line only — rejected, it would not
pull the following text up and would leave the artifact width wrong.

**3. Let the viewport supply final padding.**
bubbles viewport pads each visible line to the content width at render time, so
re-flowed lines need no explicit trailing padding. Rationale: avoids duplicating
padding logic and matches how all other article lines already reach full width.

## Risks / Trade-offs

- [Re-flow could join lines across a paragraph boundary] → Grouping keys on
  blank lines; glamour always separates paragraphs with a blank padded line, so
  a blank separator terminates a group.
- [Re-flow of an unusual paragraph whose short line is intentional] → Only
  paragraphs with a sub-width line are touched, and plain-paragraph groups
  (non-blockquote, non-list, non-indented) are the only ones eligible; the
  signature is specific to the strip artifact.
- [OSC 8 spans could split during re-wrap] → x/ansi is OSC-aware (escapes are
  zero-width); short link-name spans are moved whole; spans longer than a line
  would be re-emitted per line by lipgloss the same way glamour already does.
- [Importing `charm.land/lipgloss/v2` moves it to a direct dependency] →
  It is already pinned in go.mod; `go mod tidy` records it as direct.
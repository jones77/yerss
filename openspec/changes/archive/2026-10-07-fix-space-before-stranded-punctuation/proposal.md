## Why

In compact link mode (the default), punctuation that immediately followed a link
in the source can render separated from the link name by a space. In the Intercept
article "Oregon Police Are Charging a Local News Outlet $2.5M for Records About One
Cop" the sentence renders as `$195,869.52 , for a grand total` — a space before the
comma. The surrounding HTML has no space (`$195,869.52</a>, for`), so the space is
introduced by the markdown/render/strip/re-flow pipeline.

## What Changes

- Fix the compact-mode paragraph re-flow so closing punctuation stranded after a
  stripped link href butts against the preceding link name with no space.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `compact-link-mode`: extend the "Stripped link re-wrap preserves single word
  spacing" requirement so a paragraph re-wrap never inserts a space before closing
  punctuation stranded after a hidden URL.

## Impact

- `internal/ui/compose/reflow.go` (`reflowParagraph`) and
  `internal/ui/compose/markdown.go` (a new leading-closing-punctuation matcher).
- Regression test in `internal/ui/compose/reflow_test.go`.
- No new dependencies.

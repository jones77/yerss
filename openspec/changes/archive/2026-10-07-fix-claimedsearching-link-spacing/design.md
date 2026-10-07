## Context

Compact mode (default) strips the visible href from body links and re-wraps the surrounding paragraph. The relevant code is `compose.StripLinkHrefs`/`collectLinkSpans`/`groupLinkUnits` in `internal/compose/links.go` and `compose.ReflowShortLines`/`reflowParagraph` in `internal/compose/reflow.go`, invoked from `internal/ui/articleview/article.go`'s derivation pipeline. The failing article (id 10431) contains `...Department <a href="…">claimed</a> searching for records…`, which the HTML→markdown converter renders as `[claimed](…) searching`; the rendered compact-mode text shows `claimedsearching`.

## Goals / Non-Goals

**Goals:**
- Guarantee exactly one space between a stripped link name and following text in compact mode.

**Non-Goals:**
- No change to expanded mode, clickability, or OSC 8 wrapping.

## Decisions

**Reproduce first, fix second.** Per AGENTS.md the bug must be reproduced from the local SQLite DB at the reported 106x31 geometry before changing the strip/reflow rules. A task reproduces the `claimed`/`searching` case and pins whether the space is lost in the href-strip step (`StripLinkHrefs`) or earlier in the markdown→render step. The fix location follows from the reproduction; the spec only constrains the observable single-space outcome.

**Fix the separator-space handling rather than adding a post-hoc space insertion.** The correct fix restores the separator space at the point it is dropped (the href unit's leading space in `groupLinkUnits`, or the reflow join). Inserting a space after the fact risks a double space when the separator survives, which the existing scenarios already forbid.

## Risks / Trade-offs

- [Fix in the wrong layer (strip vs reflow vs converter)] → reproduce with the real article first and add a focused unit test on the failing paragraph before changing code.
- [Re-introducing the double-space case] → keep the existing double-space scenarios green while adding the missing-space scenario.

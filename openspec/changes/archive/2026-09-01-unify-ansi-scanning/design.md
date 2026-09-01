## Context

Both walkers skip `ESC` + CSI/OSC sequences and advance by display width. See proposal.md.

## Goals / Non-Goals

**Goals:** one scanner; identical link-span and styled-cut behavior.

**Non-Goals:** no new capabilities.

## Decisions

### D1 — One scanner API

`scanANSI(s string) []segment` (byte offset, display column, escape or text) used by `parseLinkSpans` and `styledCells`. Rationale: the two consumers both need "walk cells, skip escapes".

### D2 — Keep `parseLinkSpans` behavior

The OSC 8 URL extraction and BEL vs ESC-backslash termination stay as-is; only the walking is shared.

## Risks / Trade-offs

- [Subtle behavior differences] → Cover with the existing markdown and OSC 8 tests.

## Context

`internal/ui` holds ~17 source files and ~10k lines of behavior. See proposal.md.

## Goals / Non-Goals

**Goals:** smaller, concern-aligned units.

**Non-Goals:** no new features; no public API.

## Decisions

### D1 — Subpackage split, fallback to files

Target `internal/ui/render` and `internal/ui/compose` first (leaf dependencies), then components. If the model/component types create an import cycle, keep one package and split by file only. Rationale: Go import cycles are a hard constraint; files deliver most of the readability win at lower risk.

## Risks / Trade-offs

- [Import cycles] → Fall back to file-level split within one package.
- [Test churn] → Keep test files adjacent to their subject.

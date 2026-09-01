## Context

See proposal.md. The fold keys (`h`/`l`/`enter`/`space`/`tab`) are context-sensitive within the list view, so full catalog integration would require new actions plus conflict resolution; this change only centralizes them.

## Goals / Non-Goals

**Goals:** remove dead code, exact math, visible errors.

**Non-Goals:** no new keybindings; fold keys stay hardcoded but named.

## Decisions

### D1 — Inline `sourceID`

Call `store.SourceLabel` directly. Rationale: the wrapper added no value.

### D2 — Integer percent

`(n*100 + total/2) / total`. Rationale: identical rounding without float.

### D3 — Surface `LastRefreshedAt` errors

Set a status message on error rather than discarding it. Rationale: a broken store should not be silent.

### D4 — Named fold-key constants

Define the header fold keys once. Rationale: documents that they are intentional and context-sensitive.

## Risks / Trade-offs

- [Percent rounding] → Integer rounding matches the prior +0.5 behavior; covered by status-bar tests.

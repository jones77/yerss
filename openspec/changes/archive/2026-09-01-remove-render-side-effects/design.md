## Context

`nativeImageClear` is called inside `View()` and mutates `nativeSent` (and, on leaving the article, resets the map). `seen == 0` also returns a delete-all every frame. See proposal.md.

## Goals / Non-Goals

**Goals:** `View()` free of mutation; no redundant delete-all.

**Non-Goals:** no change to kitty placement cleanup semantics.

## Decisions

### D1 — Compute the clear prefix in `Update`

Fold the clear-sequence computation into the update path after each state change, storing the resulting prefix on the model for `View()` to prepend. Rationale: keeps the escape emission correct while making render pure.

### D2 — Emit delete-all only on view transitions

When the article has no native image blocks, emit the delete-all only when leaving the article view (or opening a non-native article), not every frame.

## Risks / Trade-offs

- [Ordering of cleanup vs transmit] → Keep the existing guard order; the kitty image tests pin the emitted sequences.

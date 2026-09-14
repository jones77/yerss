## Why

Five refactor proposals (`extract-link-anchor-helpers`, `extract-image-view-state`, `split-ui-package`, `complete-session-migration`, `consolidate-lipgloss-version`) are staged in `openspec/changes/`, but nothing captures the end state they converge on, nor the order in which they must land. A future session (human or build agent) opening this repo sees five unconnected changes and no target to steer toward, so it cannot sequence the work or verify each step moves toward the goal. This change writes that target down and turns the set of proposals into an executable sequence.

## What Changes

- Add an architecture overview document that describes the target `internal/ui` layout, the `internal/app` Session responsibilities, and the dependency direction each refactor enforces.
- Record the fixed execution order of the five refactors, with the rationale and the invariants each step must preserve (decode-once, O(N) render bounds, snap guards, one clear per no-native transition).
- Include a copy-paste build-agent handoff prompt so the sequence can be delegated to another session verbatim.

No production code changes.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- Documentation only; no `.go` files, specs, or dependencies change.
- Adds a durable orientation artifact for the refactor sequence (archived alongside the other changes, so it survives after the individual proposals are archived).
- Gives a build agent a self-contained brief that references the five change directories by name.

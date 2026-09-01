## Why

Two hand-rolled ANSI walkers exist: one in `markdown.go` (`skipEscape`, `styledCells`, `cutStyledWidth`) and one in `article.go` (`parseLinkSpans`, `oscEnd`). They duplicate escape-skipping and cell-advance logic that can drift.

## What Changes

- Introduce one escape-aware scanner in the ui package and route both call sites through it.
- No behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui`: new shared scanner; `markdown.go` and `article.go` refactored onto it.

## Why

`articleAttribution` and `inlineAttribution` in `internal/ui` re-derive the same caption/credit/source fallback chain with slight differences. One function with an explicit mode removes the duplication and keeps the two image kinds consistent.

## What Changes

- Fold the lead and inline attribution logic into one shared function.
- No behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/image.go`, `internal/ui/article.go`: attribution refactor.

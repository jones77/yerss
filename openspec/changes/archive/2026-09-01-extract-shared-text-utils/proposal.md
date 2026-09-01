## Why

Three small text helpers are duplicated across packages: `formatSize` exists in both `cmd/yerss` and `internal/ui`, the markdown-escape loop exists in both `convert` and `internal/ui`, and maximum-line-width is computed three different ways. The `convert` package also sits at the top level with no relation to `main`. Consolidating removes drift and shrinks the surface.

## What Changes

- Add `internal/textutil` with `FormatSize`, `EscapeMarkdown`, and `MaxLineWidth`.
- Move the top-level `convert` package to `internal/convert`.
- Delete the duplicated `formatSize` in `cmd/yerss/initdb.go` and the duplicated escape/width helpers, pointing call sites at the shared package.
- No behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/textutil`: new package.
- `convert` → `internal/convert`: moved; all import paths updated.
- `cmd/yerss`, `internal/ui`, `internal/store`, `internal/image`: import and helper updates.

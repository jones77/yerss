## Why

`convert.WrapText` — the link-aware soft wrapper that predates the glamour
renderer — has no production call site. It survives only in `convert/wrap.go`
(170 lines) and `convert/wrap_test.go` (224 lines of tests across 14 cases), so
it is dead code that gives false confidence and must be maintained against
future refactors. The archived `glamour-article-rendering` change kept it
"to avoid churning the image change's assumptions," but `add-article-images`
has since landed and does not use `WrapText`, so that rationale is obsolete.

## What Changes

- Delete `convert/wrap.go` (`WrapText`, `linkTokens`, `truncateURL`,
  `isTrailingPunct`, and the `linkOpenSeq` constant).
- Delete `convert/wrap_test.go` and its 14 tests.
- Remove the now-unused `strings`, `unicode`, and `github.com/charmbracelet/x/ansi`
  imports from the `convert` package if no remaining file uses them.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

_None._ This is a pure deletion of unused code: no spec-level behavior changes,
so this change declares `skip_specs: true`.

## Impact

- **Converter**: `convert` package shrinks to `Convert` plus its image/blank-line
  post-processing. No public API change (`Convert` is the only remaining export).
- **Tests**: `convert/wrap_test.go` is removed; `convert/convert_test.go` is
  unchanged and remains the safety net.
- **Verification**: `go build ./...`, `go test ./convert/...`, and `go vet ./...`
  must pass with no remaining references to `WrapText`.

## Why

Five pure string/link helpers (`anchorableLine`, `stripLeadingURLTokens`, `isURLToken`, `findAnchorLine`, `commonPrefix`) live in `internal/ui/image.go`, where they are unrelated to image rendering and drag `strings`/`ansi` link-anchoring logic into the file that also carries the native-image and load-scheduling paths. They are pure functions with no `*Model` receiver, so they belong with the other link/reflow helpers already extracted into `internal/ui/compose`.

## What Changes

- Move `anchorableLine`, `stripLeadingURLTokens`, `isURLToken`, `findAnchorLine`, `commonPrefix`, and the `minURLFragmentLen` constant from `internal/ui/image.go` (lines ~666–780) into a new `internal/ui/compose/anchor.go`.
- Decouple the three link-aware helpers from the `ui.articleLink` type: they only ever read the link's `url` field, so their parameter becomes `urls []string`.
- Add a tiny `urlsOf([]articleLink) []string` helper (or equivalent) in the `ui` package for the two call sites in `image.go`'s link-toggle anchor path.
- No behavior change: the anchor/search semantics are identical; only the location and the link parameter type move.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/image.go`: five functions and one constant removed; two call sites updated to pass URLs.
- `internal/ui/compose/anchor.go`: new file holding the moved helpers.
- `internal/ui/link_expand_test.go`: tests for the moved helpers adapted to the new package/parameter.

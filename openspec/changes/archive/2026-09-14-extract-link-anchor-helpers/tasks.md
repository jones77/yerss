## 1. Move helpers into compose

- [x] 1.1 Create `internal/ui/compose/anchor.go` with `anchorableLine`, `stripLeadingURLTokens`, `isURLToken`, `findAnchorLine`, `commonPrefix`, and `minURLFragmentLen`
- [x] 1.2 Change `anchorableLine`, `stripLeadingURLTokens`, `isURLToken` to accept `urls []string` (not `[]articleLink`)
- [x] 1.3 Add `urlsOf(links []articleLink) []string` in `internal/ui` (next to `articleLink`) and update the two `image.go` call sites
- [x] 1.4 Remove the moved definitions from `internal/ui/image.go`

## 2. Tests

- [x] 2.1 Move the anchor-helper unit cases from `link_expand_test.go` into `internal/ui/compose/anchor_test.go` with `[]string` URLs
- [x] 2.2 Keep the link-toggle integration cases in `internal/ui/link_expand_test.go` exercising the observable behavior

## 3. Validation

- [x] 3.1 Run `go build ./...`
- [x] 3.2 Run `go test ./...`

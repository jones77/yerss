## 1. Config grouping

- [x] 1.1 Add exported `GlobalActions`, `ListActions`, and `ArticleActions` to
      `internal/config/keys.go`, derived from each catalog action's declared
      views (list-only and article-only actions split out; everything else is
      Global), preserving catalog order

## 2. Help rendering

- [x] 2.1 Rewrite `renderHelp` in `internal/ui/popup.go` to render the Global,
      List view, and Article view sections with blue-role headings, keeping the
      existing action-row layout
- [x] 2.2 Lay the List view and Article view sections out in a second column to
      the right of Global, capping the popup (border included) at 70 columns
      and truncating rows that would overflow

## 3. Tests

- [x] 3.1 Update `internal/ui/keys_test.go` so `TestHelpListsEveryActionInOrder`
      asserts every action line appears and the three headings appear in order
- [x] 3.2 Add a `internal/config/keys_test.go` test pinning the exact membership
      of the three groups (all actions, no overlap, no gaps)

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go test ./...`
- [x] 4.2 Run `openspec validate`
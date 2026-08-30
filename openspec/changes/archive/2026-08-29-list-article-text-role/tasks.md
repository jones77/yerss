## 1. Rendering change

- [x] 1.1 `internal/ui/list.go`: in `renderArticleRow`, render read titles,
      the publication time, and the source identifier with `m.palette.Text`
      instead of `m.palette.Dim`; rename the muted `dim` style to `text`; leave
      unread titles on `m.palette.Bright` and day headers on `m.palette.Dim`

## 2. Tests

- [x] 2.1 Update `internal/ui/polish_test.go` so
      `TestArticleRowTitleColorChangesWithRead` expects the read title, time,
      and source in the text role, and re-verify the selection-background and
      time tests still pass

## 3. Verification

- [x] 3.1 Run `go build ./...` and `go test ./...`
- [x] 3.2 Run `openspec validate`
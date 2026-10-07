## 1. Keybinding

- [x] 1.1 Add `CopyDebugInfo` action constant in `internal/config/keys.go`
- [x] 1.2 Add a `catalog` entry for `CopyDebugInfo` with a default key and `ViewArticle` view, avoiding collisions with existing bindings (check `EffectiveKeys`)
- [x] 1.3 Run keybinding tests and confirm the action appears in the config template, help, and self-update registry

## 2. Snapshot & copy command

- [x] 2.1 Add a snapshot formatter that assembles the `key: value` block from `s.width`, `s.height`, `s.Viewport.YOffset`, `len(s.lines)`, `s.article.Title`, `s.article.Link`, `s.article.ID`, `s.ascii`, and `s.LinkExpand`
- [x] 2.2 Add a `CopyDebugInfoCmd` that wraps the snapshot string with the existing `CopyTextCmd` (label "copied debug info")

## 3. Dispatch

- [x] 3.1 Add a `CopyDebugInfo` case to the `Action` enum and `UpdateKey` switch in `internal/ui/articleview/update.go`
- [x] 3.2 Wire the new action to `CopyDebugInfoCmd`

## 4. Verification

- [x] 4.1 Add/update tests asserting the snapshot contains the width×height, yOffset/total-lines, URL, title, and mode flags
- [x] 4.2 Run `go build ./...` and `go test ./...` to confirm the change compiles and passes

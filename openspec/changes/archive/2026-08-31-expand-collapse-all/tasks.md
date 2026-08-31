## 1. Keymap (config)

- [x] 1.1 Add an `expand_toggle` catalog action bound to `x`, valid only in the list
  view (`internal/config/keys.go`)
- [x] 1.2 Update `internal/config/config_test.go` for the catalog change
  (default keybindings gain `expand_toggle`; List view help group gains Expand
  Toggle)

## 2. Fold-all behavior (ui)

- [x] 2.1 Add `expandToggle` to `internal/ui/list.go`: collapses all day groups
  when every group is expanded, and expands all otherwise (a mixed state
  expands); clamp the cursor afterward
- [x] 2.2 Dispatch `ExpandToggle` on `x` in `updateList`, independent of the cursor
  row
- [x] 2.3 Add UI tests: all-expanded collapses all, all-collapsed expands all,
  mixed expands all, and the toggle works from both header and article rows

## 3. Verification

- [x] 3.1 Run `go vet ./...`, `go build ./...`, and `go test ./...` and confirm
  all pass
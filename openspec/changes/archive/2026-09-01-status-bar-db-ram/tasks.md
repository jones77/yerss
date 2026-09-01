## 1. Formatter

- [x] 1.1 Add `formatMB` in `internal/ui` and unit-test floored MB output and the no-space suffix

## 2. Status bar

- [x] 2.1 Read the decoded-image byte total from the image cache counter and render `RAM <n>MB`
- [x] 2.2 Render `DB <n>MB` from the existing `dbSize` and reorder the right side to `?: help · RAM · DB · percent · position`
- [x] 2.3 Update status-bar tests for the new layout and the `0MB` case

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...` and `openspec validate status-bar-db-ram`

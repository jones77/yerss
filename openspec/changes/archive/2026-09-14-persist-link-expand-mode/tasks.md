## 1. Store layer

- [x] 1.1 Add `SaveLinkExpandMode(expanded bool)` and `LoadLinkExpandMode() (bool, error)` to `internal/store/store.go`, storing a JSON envelope under a `state`-table key like `link_expand_mode` (upsert on save, `sql.ErrNoRows` → false on load), mirroring `SaveLastSelection`/`LoadLastSelection`
- [x] 1.2 Add round-trip tests in `internal/store/store_test.go`: save true → load true, save false → load false, no row → false

## 2. Session delegation

- [x] 2.1 Add `SaveLinkExpandMode(expanded bool)` and `LoadLinkExpandMode() (bool, error)` to `internal/app/session.go` delegating to the store

## 3. UI persistence

- [x] 3.1 In `internal/ui/selection.go`, persist the mode in `persistSelection` (call the session save with `m.linkExpand`) so every quit path records it
- [x] 3.2 In `internal/ui/model.go` `Init`, load the persisted mode before `restoreSelection` and apply it via `SetLinkExpand`, so a restored article composes in the saved mode

## 4. Tests

- [x] 4.1 Add a link-expand test in `internal/ui/link_expand_test.go` (or `selection_test.go`) that toggles to expanded, persists, builds a fresh model, restores, and asserts the first article composes in expanded mode with the URL-above-title header (spec: "Restored expanded mode renders the URL-above-title header")
- [x] 4.2 Verify toggling does not touch the config file: assert the mode write only touches the `state` table (spec: "Toggling does not touch the config file")

## 5. Verification

- [x] 5.1 Run `gofmt`, `go vet ./...`, and `go test ./...` (all packages) to green
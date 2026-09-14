## 1. Tests

- [x] 1.1 Add a test asserting the header URL line is absent and the title is the first content line in compact mode
- [x] 1.2 Add a test asserting the header URL line is present (first line, blue role) in expanded mode

## 2. Implementation

- [x] 2.1 Thread the link display mode into `renderHeader` and skip `renderHeaderLink` when compact (keeping the `a.Link == ""` early return)

## 3. Verification

- [x] 3.1 Run `go test ./internal/ui/...` and confirm the header and link tests pass
- [x] 3.2 Run `go vet ./...` and confirm no new findings
- [x] 3.3 Run `openspec validate --all --strict` and confirm the change validates

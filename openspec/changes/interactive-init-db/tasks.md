## 1. Flag wiring and confirm helper

- [x] 1.1 In `cmd/yerss/main.go`, add `var interactive bool` and register `pflag.BoolVarP(&interactive, "interactive", "i", false, "prompt before --init-db")`
- [x] 1.2 Add a `confirmInitDB(path string) (code int, ok bool)` helper that stats the file, counts articles (`SELECT COUNT(*) FROM articles` on a temporarily-opened store), prints the size/count prompt, reads stdin, and returns ok/refuse per the design; treat a missing file as immediate ok (no prompt)
- [x] 1.3 Rewire `main()`: when `initDB && interactive`, call `confirmInitDB` and `os.Exit(code)` on refusal; keep the unconditional `if initDB { resetDatabase(...) }` for the non-interactive path

## 2. TTY detection and size formatting

- [x] 2.1 Detect non-TTY stdin via `os.Stdin.Stat()` mode bits (`os.ModeCharDevice`); on non-TTY print a refusal to stderr and return exit 1
- [x] 2.2 Add a local human-readable size formatter in `cmd/yerss` (B/KB/MB/GB, one decimal) with unit tests; do not export UI internals

## 3. Usage trim

- [x] 3.1 In `pflag.Usage`, drop the per-file descriptions from the `Files:` block, keeping the label + resolved path lines and the `[data]` relocation note

## 4. Tests and verification

- [x] 4.1 Add `main_test.go` coverage: `-i` parses; `-i` with `-j`/`-a`/`-e`/`-c` is a no-op
- [x] 4.2 Add table-driven `confirmInitDB` tests (temp DB, stubbed stdin): missing file → ok no prompt; TTY `y` → ok; TTY `n`/empty/garbage → refuse (exit 0) and DB untouched; non-TTY → refuse (exit 1) and DB untouched
- [x] 4.3 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [x] 4.4 Run `openspec validate interactive-init-db`

# Tasks: json-dump-mode

## 1. Flag + early exit

- [ ] 1.1 In `cmd/yerss/main.go`, add `pflag.BoolVarP(&jsonOut, "json", "j", false, "dump articles as JSON and exit")`
- [ ] 1.2 After `store.Open` and after `runEditFlags`, call `os.Exit(runJSONDump(cfg, st))` when `jsonOut` is set, before `runStartupGate` and the TUI

## 2. Dump implementation

- [ ] 2.1 Create `cmd/yerss/json.go` with `runJSONDump(cfg *config.Config, st *store.Store) int`
- [ ] 2.2 Define `articleJSON` with snake_case JSON tags for all stored fields; convert `store.Article` → `articleJSON` (`published_at`/`fetched_at` as RFC3339, `categories` initialized to `[]string{}`)
- [ ] 2.3 Bucket `store.ListArticles("")` into `dateBucket` groups keyed by `t.Local().Format("20060102")`; fall back to `FetchedAt` when `PublishedAt` is zero
- [ ] 2.4 Add a `MarshalJSON` on the buckets type that writes keys in slice (descending) order so dates are reverse chronological
- [ ] 2.5 Marshal with `json.Marshal` and write to stdout; return 0 (empty store yields `{}`)

## 3. Tests

- [ ] 3.1 Dump with articles on three days → JSON object with three keys in descending date order; each date's array reverse chronological
- [ ] 3.2 All stored fields present and correctly typed; `categories` is `[]` for an article with no categories
- [ ] 3.3 Local timezone keying: an article at midnight UTC keys to the local calendar day
- [ ] 3.4 Empty store → `{}` and exit 0
- [ ] 3.5 `published_at` and `fetched_at` are RFC3339 strings

## 4. Verification

- [ ] 4.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [ ] 4.2 Run `yerss -h` and confirm `-j`/`--json` is listed
- [ ] 4.3 Run `yerss -j | jq 'keys'` and confirm dates are descending

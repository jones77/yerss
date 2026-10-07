## 1. Reproduction

- [x] 1.1 Reproduce the space-before-comma from the local SQLite DB (article id 10431) at the reported 106x32 compact-mode render, driving the real derivation pipeline
- [x] 1.2 Pin the layer: `StripLinkHrefs` leaves the comma stranded at line start and `ReflowShortLines` inserts the join space; add a focused failing test on that paragraph

## 2. Fix

- [x] 2.1 Add a matcher for lines beginning with closing punctuation and suppress the re-flow join space for them (`reflowParagraph`)
- [x] 2.2 Keep the existing double-space and missing-space scenarios green

## 3. Verification

- [x] 3.1 Add a regression test asserting `$195,869.52, for a grand total` (no space before the comma) renders in compact mode
- [x] 3.2 Run `go build ./...` and `go test ./...` to confirm the fix compiles and passes

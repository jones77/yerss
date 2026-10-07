## 1. Reproduction

- [x] 1.1 Reproduce the `claimed`/`searching` fusion from the local SQLite DB (article id 10431) at a 106x31 compact-mode render, driving the real derivation pipeline
- [x] 1.2 Pin whether the space is lost in `StripLinkHrefs`/`groupLinkUnits` (the href unit's leading space) or earlier in the markdown→render step, and add a focused failing test on that paragraph

## 2. Fix

- [x] 2.1 Restore the separator space at the point it is dropped so the re-wrap yields exactly one space between the link name and following text
- [x] 2.2 Keep the existing double-space scenarios green while fixing the missing-space case

## 3. Verification

- [x] 3.1 Add/update tests asserting `claimed searching` (single space) renders in compact mode, with no doubled and no missing space
- [x] 3.2 Run `go build ./...` and `go test ./...` to confirm the fix compiles and passes

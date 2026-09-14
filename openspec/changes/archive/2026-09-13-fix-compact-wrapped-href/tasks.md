## 1. Regression tests

- [x] 1.1 Add a compact-mode test where a link's URL is longer than the content width (so the URL wraps across lines) and assert the body shows only the link name and no part of the URL text
- [x] 1.2 Add a compact-mode test where a link name is long enough to wrap and assert the full name renders while the URL stays hidden
- [x] 1.3 Add a compact-mode test with two adjacent links to the same URL and assert both names render and both URLs are hidden

## 2. Fix StripLinkHrefs

- [x] 2.1 Rework `compose.StripLinkHrefs` to group same-target OSC 8 spans into logical units (a span joins the current unit when the text between the previous span's close and this span's open contains no space; a space ends the unit) and alternate keep/drop per unit, dropping each href unit's leading space plus all its spans

## 3. Verification

- [x] 3.1 Run `go test ./internal/ui/...` and confirm the link tests pass
- [x] 3.2 Run `go vet ./...` and confirm no new findings
- [x] 3.3 Run `openspec validate --all --strict` and confirm the change validates

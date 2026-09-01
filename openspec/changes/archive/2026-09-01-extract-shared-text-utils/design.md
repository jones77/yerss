## Context

The helpers are behaviorally identical but drifted. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:** one home for each helper; identical output to today.

**Non-Goals:** no new functionality, no API guarantees beyond this repo.

## Decisions

### D1 — `internal/textutil` package

Holds `FormatSize(int64) string`, `EscapeMarkdown(string) string`, and `MaxLineWidth([]string) int`. Rationale: these are cross-cutting text utilities with no dependency on config, store, or the TUI.

### D2 — `convert` moves to `internal/convert`

Package name stays `convert`; only the import path changes. Rationale: it is an implementation detail of the reader, not a public library.

### D3 — Keep the `formatSize` prompt copy until the shared one is imported

`cmd/yerss/initdb.go` calls `textutil.FormatSize`; `internal/ui/size.go` keeps a thin re-export or is removed. Rationale: the prompt and the status bar stay byte-for-byte identical.

## Risks / Trade-offs

- [Wide mechanical edit] → Small, compile-driven change; `go test ./...` catches missed imports.

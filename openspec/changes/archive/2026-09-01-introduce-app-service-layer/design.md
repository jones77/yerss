## Context

`Model` holds `cfg`, `store`, five image caches/renderers, refresh state, selection, and status. See proposal.md.

## Goals / Non-Goals

**Goals:** the TUI depends on one application object; orchestration moves down.

**Non-Goals:** no package split of `internal/ui` (that is a later, optional change); no new features.

## Decisions

### D1 — `app.Session`

`New(cfg, store) *Session` and methods mirroring today's model logic: `ListArticles`, `ArticleCount`, `DBSize`, `SetRead`, `GetArticle`, `SaveSelection`, `LoadSelection`, `FetchFeeds`, plus the image cache set. Rationale: mechanical extraction, preserving call sites.

### D2 — Incremental, not big-bang

The first pass is dependency inversion only (Model holds `*app.Session`); logic migration follows in later tasks. Rationale: keeps the change reviewable and compile-green throughout.

## Risks / Trade-offs

- [Large diff] → Split into tasks; `go test ./...` after each.
- [Session grows god-like] → Acceptable interim; later changes can split by concern.

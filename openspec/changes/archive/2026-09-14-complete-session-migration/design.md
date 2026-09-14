## Context

`Session` already owns the store, config, and image caches and exposes store delegates (`ListArticles`, `SetRead`, `GetArticle`, `SaveSelection`, `LoadSelection`, `LastRefreshedAt`, `VerifiedFeedURLs`, `ListTags`, `FetchFeeds`). What remains in the view layer is the refresh *policy*: the refresh gate, the feeds-file load, and the unfetched-feeds check.

The refresh gate uses the in-memory `lastRefreshedAt` (not the store's persisted value), because a refresh completed earlier in the session must not be gated out by a stale persisted timestamp. Preserving that semantics requires the gate methods to receive `last time.Time` from the caller rather than re-reading the store.

## Goals / Non-Goals

**Goals:** the UI stops importing `feed` for orchestration; refresh policy lives behind `Session`; behavior byte-identical.

**Non-Goals:** no new refresh features; no change to read-state policy (mark-read on open/scroll stays in the view — it is presentation-coupled); no removal of the `feed` result-type import in this change.

## Decisions

### D1 — Gate methods take `last time.Time` as a parameter

`NeedsStartupRefresh(last, now)` and `ManualRefreshAllowed(last, now)` accept the caller's cached last-refresh time instead of reading the store, matching today's `m.lastRefreshedAt` semantics exactly. Rationale: avoids a subtle gating bug after an in-session refresh.

### D2 — `Refresh()` returns `(*feed.FetchResult, error)`

`Refresh` folds `feed.LoadFeeds` + `FetchFeeds` into one call and surfaces the load error separately from the fetch result, mirroring the current `refreshFinishedMsg{result, err}` shape so `model.go`'s handler is unchanged.

### D3 — Result types stay `feed.FetchResult`/`feed.FeedOutcome` for now

The UI and `cmd/yerss` diagnostics already consume these types. Introducing `app.RefreshResult` wrappers is a separate cleanup; this change keeps the import only for the result passthrough, not for orchestration.

## Risks / Trade-offs

- [Session gains refresh policy, moving toward the "god-like" smell] → The gate logic is small and cohesive; if it grows, extract a `feed.Refresher` later (noted, not in scope).
- [Load-error vs fetch-error split is easy to get wrong] → `Refresh` returns them separately and the `refreshFinishedMsg` shape is preserved, so the existing error paths are unchanged.

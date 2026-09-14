## Why

The `introduce-app-service-layer` change did dependency inversion only: `Model` holds `*app.Session`, but the refresh *orchestration* never moved down. `model.go` still reaches into `feed` directly — constructing `feed.Gate`, calling `feed.LoadFeeds`, and `feed.CanonicalURL` — so the application layer is a thin pass-through facade over the store while the view layer owns the refresh policy. Finishing the migration gives the UI a single dependency and a real application boundary.

## What Changes

- Add `Session.Refresh() (*feed.FetchResult, error)`, which loads the feeds file and runs the fetch, so the UI no longer calls `feed.LoadFeeds`.
- Add `Session.NeedsStartupRefresh(last, now time.Time) bool`, encapsulating the "unfetched feeds" check plus the startup-gate decision.
- Add `Session.ManualRefreshAllowed(last, now time.Time) (bool, time.Duration)`, encapsulating the manual-refresh gate.
- `model.go`'s `refreshCmd`, `hasUnfetchedFeeds`, `refreshManual`, and `Init` delegate to these instead of importing `feed` for orchestration. The UI keeps the `feed.FetchResult`/`feed.FeedOutcome` types only as result passthrough (exit diagnostics).
- No behavior change: gate timing, error surfacing, and outcome reporting stay identical.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/app/session.go`: three new orchestration methods; `Refresh` becomes the sole entry point for a refresh pass.
- `internal/ui/model.go`: `hasUnfetchedFeeds`, `refreshCmd`, `refreshManual`, `Init` slimmed; `feed` imports reduced to result types.
- `internal/app/*_test.go`: new unit tests for the gate policy and refresh orchestration.

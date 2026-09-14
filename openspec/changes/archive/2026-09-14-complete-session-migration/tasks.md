## 1. Add orchestration to Session

- [x] 1.1 Add `Session.Refresh() (*feed.FetchResult, error)` — load feeds file, then `FetchFeeds`
- [x] 1.2 Add `Session.NeedsStartupRefresh(last, now time.Time) bool` — unfetched-feeds check + startup gate
- [x] 1.3 Add `Session.ManualRefreshAllowed(last, now time.Time) (bool, time.Duration)` — manual gate

## 2. Delegate from the UI

- [x] 2.1 Rewrite `model.go` `refreshCmd` to call `m.sess.Refresh()`; keep `refreshFinishedMsg{result, err, fetchedAt}` intact
- [x] 2.2 Rewrite `hasUnfetchedFeeds` + the `Init` startup branch to use `m.sess.NeedsStartupRefresh`
- [x] 2.3 Rewrite `refreshManual` to use `m.sess.ManualRefreshAllowed`
- [x] 2.4 Remove `feed.LoadFeeds`/`feed.Gate`/`feed.CanonicalURL` orchestration usage from `model.go` (keep `feed.FetchResult`/`feed.FeedOutcome` types)

## 3. Tests

- [x] 3.1 Add `internal/app` unit tests for `Refresh`, `NeedsStartupRefresh`, and `ManualRefreshAllowed` (gate timing, unfetched-feeds bypass, load-error surfacing)
- [x] 3.2 Keep/adjust `init_test.go` and `refresh_outcomes_test.go` to cover the same observable behavior

## 4. Validation

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test ./...`

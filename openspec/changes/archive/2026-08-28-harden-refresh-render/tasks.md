## 1. Bounded feed fetch

- [x] 1.1 Add refresh-path timeout constants (per-feed 15s, overall 60s) and a concurrency-bound constant (`maxConcurrentFetches`, start at 8) in `internal/feed/fetch.go`
- [x] 1.2 Add a semaphore (buffered `chan struct{}`) to `fetchFeeds`, acquired before launching each feed goroutine and released on return, capping in-flight requests at the bound
- [x] 1.3 Make `fetchFeeds` require a non-nil context; update `FetchFeeds` to build a per-feed `http.Client{Timeout: perFeed}` plus a `context.WithTimeout(background, overall)` and pass them in, mirroring `VerifyFeeds` exactly
- [x] 1.4 Confirm `VerifyFeeds` still routes through the same bounded core with its existing 10s/30s constants and unchanged behavior

## 2. Popup overlay correctness

- [x] 2.1 Reimplement `overlay` in `internal/ui/model.go` using lipgloss placement (`lipgloss.Place`/join helpers) so centering is computed from display width with ANSI escapes excluded
- [x] 2.2 Verify base view content outside the popup region is preserved after composition

## 3. Article border height clamp

- [x] 3.1 In `renderArticleBorder`, compute an effective `padY' = min(padY, max(0, (h-3)/2))` and floor `viewportH` at 1 so the assembled frame never emits more lines than `h`; pad with blank framed rows (or truncate defensively) to exactly `h` on degenerate sizes

## 4. Tests

- [x] 4.1 T2: add a `FetchFeeds` test against an httptest server that accepts the connection but never responds; assert the pass returns within a bounded time and records that feed as an error in the result
- [x] 4.2 T3: add a `VerifyFeeds` overall-deadline test where the server sleeps beyond the overall deadline; assert in-flight fetches are cancelled and the pass completes
- [x] 4.3 T4: add an `overlay` positioning test with a known base and a distinct marker popup; assert the popup lands at the centered column and the base content outside the popup is intact
- [x] 4.4 T7: add a `renderArticleBorder` height test with `h=3, padY=2`; assert `len(lines) <= h`

## 5. Verification

- [x] 5.1 Run `go build ./...`, `go vet ./...`, `go test ./... -race`; resolve any fallout from the `FetchFeeds`/`overlay`/border changes
- [x] 5.2 Confirm `cmd/yerss` gate tests still pass and the startup-gate behavior (exit codes 2/3, offline trust) is unchanged

## Why

The fetch-to-render pipeline has two correctness defects that survive the existing
tests. `FetchFeeds` (the TUI's manual and startup refresh path) has no per-feed or
overall timeout, so a single hung RSS server hangs the entire refresh forever with
no way to cancel. `overlay`, which composes the tag/help popups over the list or
article view, measures string width by rune count over **styled** strings, counting
ANSI escape runes as visible columns — so popups render glued to the left edge and
inject raw escape codes into the base view rather than centering. A related render
gap lets the article border emit more lines than the terminal height on tiny
terminals with non-trivial padding.

## What Changes

- **Bounded `FetchFeeds`:** `FetchFeeds` SHALL apply a per-feed HTTP timeout and an
  overall deadline, mirroring the existing `VerifyFeeds` pattern. The refresh path
  SHALL never block indefinitely; a hung feed is reported as an error in the
  `FetchResult` rather than stalling the pass.
- **Unified fetch core:** The bounded-timeout behavior SHALL be shared between
  `FetchFeeds` and `VerifyFeeds` via one internal `fetchFeeds` path that requires
  timeouts, so the gate and the TUI cannot diverge again.
- **Concurrency bound:** `fetchFeeds` SHALL cap the number of in-flight feed
  requests with a semaphore instead of spawning one unbounded goroutine per feed.
- **Popup overlay correctness:** `overlay` SHALL measure display width with ANSI
  escapes stripped (and using rune width, not rune count), so popups are centered on
  the visible base and the base content outside the popup is preserved.
- **Border height invariant:** `renderArticleBorder` SHALL never emit more lines
  than the available terminal height, including when the viewport height is floored
  at 1 under tiny terminals with padding.
- **Tests:** Add a `FetchFeeds` test against a hung server (bounded return), a
  `VerifyFeeds` overall-deadline cancellation test, an `overlay` positioning test
  (popup lands centered; base preserved outside the popup), and a border-height
  test (degenerate `h`/`padY` yields `≤ h` lines).

## Capabilities

### New Capabilities
<!-- None — both affected capabilities already exist. -->

### Modified Capabilities
- `feed-pipeline`: `FetchFeeds` gains mandatory per-feed and overall timeouts and a
  concurrency bound; the bounded fetch core is unified with `VerifyFeeds`.
- `reader-ui`: the popup overlay gains a display-width centering invariant
  (ANSI-stripped, rune-width aware), and the article border gains a
  never-exceed-terminal-height invariant.

## Impact

- **Code:** `internal/feed/fetch.go` (timeouts, semaphore, unified bounded core),
  `internal/ui/model.go` (`overlay` width measurement), `internal/ui/border.go`
  (line clamping).
- **APIs:** No public API changes. Internal: `FetchFeeds` may gain timeout
  parameters or reuse `VerifyFeeds`'s signature shape; the unbounded one-goroutine
  behavior is removed.
- **Dependencies:** None added. Reuses `net/http`, `context`, `sync`, and
  `github.com/mattn/go-runewidth` (already in use elsewhere in `ui`).
- **Behavior:** A hung feed no longer freezes the TUI's refresh; popups are
  visually centered instead of left-aligned with corrupted base coloring; tiny
  terminals no longer overflow the article frame. No first-run or offline
  behavior changes.

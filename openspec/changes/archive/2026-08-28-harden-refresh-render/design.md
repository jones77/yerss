## Context

See `proposal.md` for motivation. Today the feed layer has two fetch entry points:
`VerifyFeeds` (used by the startup gate) is correctly bounded — it sets a per-feed
`http.Client.Timeout` and a `context.WithTimeout` overall deadline. `FetchFeeds`
(used by the TUI's startup and manual refresh via `Model.refreshCmd`) is not: it
calls `gofeed.NewParser()` whose client has no timeout, and it spawns one goroutine
per feed with no concurrency bound. Both eventually call a shared `fetchFeeds(st,
urls, parser, ctx)`, but `ctx` is `nil` on the `FetchFeeds` path, so neither
deadline nor cancellation applies.

Separately, `overlay` (in `internal/ui/model.go`) composes the tag/help popups over
the base view by measuring line width with `len([]rune(l))` on lipgloss-styled
strings. ANSI escape runes are counted as visible columns, inflating the popup
width and forcing the computed left offset negative (clamped to 0), so popups render
at the left edge and the popup's escape bytes get copied into the base lines.
`renderArticleBorder` computes its line count as `2 + 2*padY + viewportH` where
`viewportH` is floored at 1; when `h` is tiny and `padY > 0`, that total exceeds `h`.

## Goals / Non-Goals

**Goals:**
- Make every feed fetch bounded by per-feed and overall deadlines, with a
  concurrency cap, via one shared implementation used by both the gate and refresh.
- Make popup overlay composition center popups by visible display width and preserve
  the base outside the popup.
- Make the article border never emit more lines than the terminal height.

**Non-Goals:**
- Config-driven refresh timeout values. Use package-level constants for now; a
  future change can add `[refresh]` options (which would bump the config schema).
- Reducing the per-article query count (the existence check that classifies New vs
  Updated stays as-is). That is a separate optimization.
- Pagination of `ListArticles`. Out of scope; tracked separately.
- Changing the popup content, sizing rules, or keybindings.

## Decisions

**D1 — Unify on one bounded `fetchFeeds`.** Keep the single internal `fetchFeeds`
core, but make timeouts *required*, not optional. Both `FetchFeeds` and `VerifyFeeds`
construct a per-feed `http.Client{Timeout: perFeed}` and a `context.WithTimeout(
background, overall)`, then call the shared core with a non-nil context. The gate
keeps its existing 10s/30s constants; the refresh path uses package constants
(per-feed 15s, overall 60s) chosen to be generous for a background TUI refresh
without letting a hung server freeze the UI. *Alternative:* make `FetchFeeds` take
timeout params from callers — rejected because `refreshCmd` would have to pick
values, and we want one obvious default. *Alternative:* reuse the gate's 10s/30s for
refresh — rejected as too aggressive for large feed sets in the background.

**D2 — Concurrency cap via semaphore.** Acquire a buffered channel slot
(`make(chan struct{}, N)` with N=8) before launching each feed goroutine, release on
return. This bounds in-flight HTTP requests and SQLite write contention regardless
of how many feeds are configured. *Alternative:* a fixed worker pool reading a jobs
channel — equivalent but more code for no gain here.

**D3 — Overall-deadline cancellation already propagates.** `fetchFeeds` already
breaks the spawn loop when `ctx.Err() != nil` and uses `ParseURLWithContext`; with
D1 the refresh path now passes a real context, so in-flight fetches cancel at the
overall deadline. No new cancellation wiring needed.

**D4 — Reimplement `overlay` with lipgloss placement.** Replace the manual
rune-index copy with `lipgloss.Place` (or `JoinVertical`+`JoinHorizontal`), which
measures display width with ANSI escapes excluded and centers the popup block over
the base block correctly. This also makes overlaying rune-width aware for free.
*Alternative:* strip ANSI with a regex before measuring and keep the manual copy —
rejected because re-applying stripped styles to the copied region is fragile, and
lipgloss already solves this exactly.

**D5 — Clamp article border to terminal height.** Compute an effective
`padY' = min(padY, max(0, (h-3)/2))` so that `2 + 2*padY' + max(1, h-2*padY'-2) <= h`,
floor `viewportH` at 1, and if the assembled line count is still short of `h`, pad
with blank framed rows; if (defensively) over `h`, truncate to `h`. Normal-sized
terminals are unchanged because the clamp only activates when `h - 2*padY - 2 < 1`.
*Alternative:* drop padding entirely on tiny terminals — rejected, the clamp
preserves padding when it fits.

## Risks / Trade-offs

- **[Refresh now fails faster on slow networks]** → A 15s per-feed / 60s overall
  ceiling may cut off a legitimately slow feed. Mitigation: values are generous for
  background refresh and can be made config-driven later (non-goal here); failures
  are reported in the status bar, not silent.
- **[overlay output shape changes]** → lipgloss.Place may pad differently than the
  old manual copy, so popup positioning shifts for everyone, not only the broken
  case. Mitigation: T4 asserts the popup lands centered and the base is preserved;
  visual review on real terminals.
- **[Semaphore choice of 8]** → Too low starves throughput; too high reintroduces
  burst contention. Mitigation: 8 is a conservative default for a personal
  reader; constant is centralized for easy tuning.
- **[Border clamp changes tiny-terminal rendering only]** → Low risk; covered by T7.

## Migration Plan

No data or config migration. The change is internal library behavior. Rollback is a
plain revert; no persisted state depends on the new timing or rendering.

## Open Questions

- Should the refresh per-feed/overall timeouts be exposed as `[refresh]` config
  options (schema bump)? Deferred — answer later without changing these specs; the
  current constants suffice.

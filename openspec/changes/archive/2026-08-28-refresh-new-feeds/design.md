# Design: refresh-new-feeds

## Root cause

`Model.Init` decides whether to refresh solely from elapsed time:

```go
gate := feed.Gate{...}
if gate.NeedsStartupRefresh(m.lastRefreshedAt, time.Now()) {
    return m.refreshCmd()
}
return nil
```

`NeedsStartupRefresh` returns true only when `now - last > MinInterval`. A newly
added feed URL is invisible to this check, and the startup verification gate
already returned 0 via `HasVerifiedFeed`, so nothing fetches it.

## Fix

Add a "new feed detected" trigger that bypasses the time gate, using data the
`feeds` table already keeps.

### Store

Add `VerifiedFeedURLs() (map[string]bool, error)` returning the set of feed URLs
whose `last_fetched_at` is non-NULL:

```sql
SELECT url FROM feeds WHERE last_fetched_at IS NOT NULL
```

### Model.Init

Load the configured URLs once, compute the unfetched set, and force a refresh
when it is non-empty:

```go
urls, err := feed.LoadFeeds(m.cfg.FeedsFile())
// on error: fall through to the time gate (refreshCmd surfaces the error)
verified, _ := m.store.VerifiedFeedURLs()
unfetched := 0
for _, u := range urls {
    if !verified[u] { unfetched++ }
}
if unfetched > 0 {
    return m.refreshCmd()          // bypass the 15-minute gate
}
if gate.NeedsStartupRefresh(last, time.Now()) {
    return m.refreshCmd()
}
return nil
```

`refreshCmd` already re-reads `feeds.txt` and calls `FetchFeeds(urls)`, so the
new feed is fetched alongside the rest. No change to the refresh command itself.

### Why not the verification gate

The verification gate's job is "refuse to start if nothing works," not "fetch
every feed." Forcing a network check there would break the offline-safe shortcut
and could block startup behind a dead new feed. The refresh path is the right
place: it fetches all feeds and reports per-feed errors in the status bar
without refusing to start.

## Edge cases

- **feeds.txt unreadable**: `LoadFeeds` returns an error; we skip the new-feed
  check and fall through to the time gate. `refreshCmd` will surface the error
  via `refreshFinishedMsg.err`.
- **New feed is dead (404/timeout):** it is recorded as an error in the
  `FetchResult` and shown in the status bar; the other feeds still refresh. It
  does not get a verified row, so the next startup retries it.
- **URL re-added after being removed:** treated as new (no verified row), so it
  refreshes. Correct.

## Tests

- `Store.VerifiedFeedURLs`: returns only URLs with non-NULL `last_fetched_at`;
  stub rows (NULL) are excluded.
- `Model.Init`: a configured URL with no verified row forces `refreshCmd` even
  when the last refresh was moments ago; all-verified + recent does not.
- Existing startup-gate tests remain green.

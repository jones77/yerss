# Refresh when new feeds are added to feeds.txt

## Why

Adding feed URLs to `feeds.txt` does not cause a refresh. The startup
verification gate short-circuits on any previously-verified feed
(`HasVerifiedFeed`), and the startup refresh gate skips fetching when the last
refresh was under 15 minutes ago. So a newly added URL sits unfetched until the
user manually presses `R` past the 60-second cooldown — and even the `-e`
edit-feeds path, which continues into startup specifically so "freshly-edited
feeds are re-read," never fetches them. The refresh command itself already
re-reads `feeds.txt` each run, so the only missing piece is the trigger.

## What

Force a startup refresh when the configured feeds file contains a URL that has
never been successfully fetched — i.e. no `feeds` row for that URL with a
non-NULL `last_fetched_at`. This bypasses the 15-minute startup gate for the
specific case of new feeds, while leaving the time-based gate intact for
already-known feeds. It covers manual edits to `feeds.txt` and the `-e` path
uniformly, with no new persisted state (the `feeds` table already records what
has been fetched).

## Non-goals

- Changing the 60-second manual-refresh cooldown.
- Changing the startup verification gate's offline-safe shortcut (adding a feed
  must not block startup on a network check).
- Watching `feeds.txt` for live changes while the TUI is running.

# Add -j/--json mode to dump articles as JSON

## Why

There is no machine-readable export of the article store. A JSON dump lets
scripts, pipelines, and other tools consume yerss's articles without driving the
TUI, and gives a quick `yerss -j | jq` inspection of what is stored.

## What

- Add a `-j` / `--json` command-line flag. Instead of launching the TUI, it
  dumps all stored articles as a JSON object to stdout and exits 0.
- The object is keyed by publication date as `YYYYMMDD` in the user's local
  timezone, with dates ordered reverse chronologically (most recent first).
- Each date's value is an array of article objects sorted reverse
  chronologically (most recent first).
- Each article object includes all stored fields: `id`, `feed_url`, `guid`,
  `title`, `link`, `author`, `published_at`, `content`, `description`, `read`,
  `fetched_at`, and `categories`.
- No network fetch or startup verification gate runs in this mode; it reads
  straight from the SQLite store. An empty database emits `{}` and exits 0.

## Non-goals

- Persisting or emitting gofeed fields the store does not currently capture
  (enclosures, extension/custom fields, updated dates, images).
- Re-fetching feeds before dumping.
- Streaming/pretty-printing options (compact JSON to stdout for now).

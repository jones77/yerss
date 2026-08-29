# Group the article list by collapsible day headers

## Why

The list view is a flat reverse-chronological run of titles with a red `yerss`
banner on top. It is hard to scan where one day ends and the next begins, there
is no per-article time, the status bar shows only the filtered count (not the
total), and the banner is visual noise — the most recent day header is the
information that actually belongs at the top.

## What

- **Group articles under a collapsible header per day**, most recent day first.
  Each header shows the long local date, e.g. `Saturday 20th August, 2025`
  (weekday, day with ordinal, month, comma, year).
- **Per-article timestamp column**: each article row shows `HH:MM:SS` on the
  right, in the user's local timezone.
- **`<n>/<total>` status bar**: `n` = articles currently shown (after any tag
  filter), `total` = all articles in the database.
- **Remove the `yerss` title banner.** The most recent day header becomes the
  first content at the top of the list (i.e. "the last date for which we got
  news").
- **Fold keys** (context-sensitive on day-header rows): `h` collapses, `l`
  expands, and `Enter` / Space / Tab toggle. From an article row, `Tab` toggles
  the article's own day group. Existing article-row bindings (`h` clears the
  filter, `l`/Enter open, Space half-pages) are unchanged.

## Non-goals

- Persisting collapse state across restarts (in-memory only for now).
- Changing the article reader view, the tag popup, or refresh behavior.
- Re-fetching or re-sorting in the store; grouping is a render/model concern
  over the already reverse-chronological `ListArticles` result.

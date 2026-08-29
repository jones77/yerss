## Why

There is no lightweight, keyboard-driven, local-first terminal RSS reader that
stores every article ever fetched in a local SQLite database, auto-tags
articles from feed categories, renders a reading-progress border, is fully
configurable via TOML, and is MIT-licensed. `yerss` fills that gap using
`gofeed` for parsing and `bubbletea`/`lipgloss` for the TUI.

## What Changes

- New Go application `yerss`: a terminal RSS reader with three views (article
  list, article reader, tag popup) and vi + arrow keyboard navigation.
- Article list: unread articles in bold, read in plain; status bar with article
  count and last-refreshed date; wrap-around selection.
- Article reader: thin-line border with date + title inline in the top border,
  double-line right border showing scroll-progress fill, percent-scrolled
  indicator in the bottom border, 2-space L/R and 1-space T/B padding.
- Tag browsing: `T` opens a popup (80% screen on list, 40% on article) listing
  categories sorted by popularity with unread/total counts (unread in bold);
  selecting a tag filters the list; `Esc`/`h`/`←` clears the filter.
- Read-status: opening an article that fits the viewport marks it read
  immediately; a partial view stays unread until any downward scroll movement
  (↓/j/PgDn/^f/^d/Space), which marks it read instantly.
- Refresh model: auto-refresh only at startup if >15 min since last fetch;
  manual `R`/`Ctrl-R`/`F5` on the list view bypasses the 15-min window but has a
  60-second cooldown with a "next allowed in Xs" status message.
- Storage: every downloaded article persisted in SQLite at
  `$XDG_DATA_HOME/yerss/yerss.sqlite` (fallback `~/.local/share/yerss/`); app
  refuses to start if the DB cannot be created/written; usage message
  references the path.
- Feeds: plain-text file at `~/.config/yerss/feeds.txt` (one URL per line,
  configurable).
- Configuration: TOML at `~/.config/yerss/config.toml` — keys, article padding,
  theme (light/dark/auto), paths, ASCII border fallback.
- HTML rendering: `html2text` with link URLs preserved; images rendered as
  `[alt]` for now.
- MIT License.

## Capabilities

### New Capabilities

- `reader-ui`: The three TUI views (article list, article reader, tag popup),
  their rendering, keyboard navigation, read-status rules, scroll-progress
  border, and tag filtering.
- `feed-pipeline`: Feed URL parsing, gofeed fetching, the refresh model
  (startup 15-min gate + manual 60-second cooldown), article deduplication, and
  SQLite storage (schema, XDG path resolution, startup validation).
- `configuration`: TOML config loading with defaults, configurable keybindings,
  light/dark/auto themes, ASCII border fallback, and path overrides.

### Modified Capabilities

<!-- None — this is a greenfield project. -->

## Impact

- **New codebase**: Go module `yerss`, all new code.
- **Dependencies**: `charmbracelet/bubbletea`, `charmbracelet/bubbles`,
  `charmbracelet/lipgloss`, `mmcdole/gofeed`, `glebarez/go-sqlite`,
  `pelletier/go-toml/v2`, `adrg/xdg`, `jaytaylor/html2text`.
- **Filesystem**: creates `$XDG_DATA_HOME/yerss/yerss.sqlite` and reads
  `~/.config/yerss/feeds.txt` + `~/.config/yerss/config.toml`.
- **No existing code or APIs affected** (greenfield).

## 1. Project setup

- [x] 1.1 Initialize Go module `yerss` and create directory structure (cmd/, internal/)
- [x] 1.2 Add dependencies: bubbletea, bubbles, lipgloss, gofeed, glebarez/go-sqlite, go-toml/v2, adrg/xdg, html2text
- [x] 1.3 Add MIT LICENSE file
- [x] 1.4 Create skeleton main.go that launches an empty bubbletea program and exits cleanly

## 2. Configuration

- [x] 2.1 Define config struct with defaults (data dir, feeds file, refresh interval, cooldown, theme, ascii, padding, keybindings)
- [x] 2.2 Implement TOML config loader with XDG path resolution and fallback to defaults when file is missing
- [x] 2.3 Implement keybinding parser: map action names to key bindings, validate no duplicate keys, error on conflict
- [x] 2.4 Add command-line flag for config path override

## 3. Storage

- [x] 3.1 Implement SQLite open/create with XDG_DATA_HOME resolution and fallback to ~/.local/share/yerss/
- [x] 3.2 Implement startup validation: refuse to start with usage message referencing DB path if cannot create/write
- [x] 3.3 Create schema migrations (feeds, articles, categories, article_categories with UNIQUE(feed_url, guid))
- [x] 3.4 Implement article upsert (dedupe by feed_url+guid, update content/metadata if changed)
- [x] 3.5 Implement read-status persistence (get/set read flag by article ID)
- [x] 3.6 Implement category/tag extraction and storage (insert categories, link via article_categories)
- [x] 3.7 Implement query functions: list articles (optionally filtered by tag), get article by ID, list tags with unread/read/total counts sorted by popularity

## 4. Feed pipeline

- [x] 4.1 Implement feeds.txt parser (one URL per line, skip blanks and comments)
- [x] 4.2 Implement feed fetcher using gofeed (concurrent fetch via tea.Cmd, return tea.Msg on completion)
- [x] 4.3 Implement startup refresh: check last_refreshed_at, fetch only if >15 min elapsed, otherwise load from DB
- [x] 4.4 Implement manual refresh (R/Ctrl-R/F5): bypass 15-min gate, enforce 60-second cooldown, show "next allowed in Xs" transient status message
- [x] 4.5 Wire fetched articles through upsert and category extraction

## 5. Article list view

- [x] 5.1 Implement list model: articles slice, cursor index, wrap-around on up/down
- [x] 5.2 Render articles: bold for unread, plain for read, cursor indicator
- [x] 5.3 Render status bar: article count, last-refreshed date, active filter tag name
- [x] 5.4 Implement tag filter state: filter articles by selected tag, clear on Esc/h/left-arrow
- [x] 5.5 Implement full/half page navigation (PgUp/PgDn/Ctrl-F/Ctrl-B, Ctrl-D/Ctrl-U/Space, g/G/Ctrl-Up/Ctrl-Down)

## 6. Article reader view

- [x] 6.1 Implement custom border renderer: top border with inline date+title, bottom border with inline percent-scrolled
- [x] 6.2 Implement double-line right border with scroll-progress fill (accent top N%, dim rest)
- [x] 6.3 Implement content area: viewport with configurable padding (2 L/R, 1 T/B default), width = terminal - 2*padX - 2
- [x] 6.4 Implement HTML-to-text conversion via html2text with link URL preservation; images as [alt]/[image]
- [x] 6.5 Implement scroll handling: line up/down, page/half-page, top/bottom navigation
- [x] 6.6 Implement read-status rules: mark read if fully visible on open; mark read on first downward scroll if partial
- [x] 6.7 Implement percent-scrolled calculation from viewport position
- [x] 6.8 Implement ASCII border fallback (Unicode glyphs → ASCII equivalents)

## 7. Tag popup

- [x] 7.1 Implement 80%-screen popup for list view: all tags sorted by popularity, wrap-around navigation
- [x] 7.2 Implement 40%-screen popup for article view: only current article's tags, same navigation
- [x] 7.3 Implement tag count display: (unread/total) with bold unread; collapse to (total) when read==0 or unread==0
- [x] 7.4 Implement tag selection → filter list view and navigate back
- [x] 7.5 Implement popup dismiss on Esc

## 8. View transitions and keymap

- [x] 8.1 Implement state machine: list ↔ article transitions, popup overlay/dismiss
- [x] 8.2 Wire full keymap from config: quit, refresh, open_article, back, move_up/down, page_up/down, half_page_up/down, top, bottom, tag_popup, toggle_read, mark_all_read, open_url, copy_url, help
- [x] 8.3 Implement article reselection on return from article view to list
- [x] 8.4 Implement `m` toggle read/unread, `a` mark all in view as read
- [x] 8.5 Implement `o` open URL in system browser (xdg-open/open/cmd), `c` copy URL to clipboard (pbcopy/xclip/clip)

## 9. Themes and polish

- [x] 9.1 Define light and dark color palettes (accent, dim, bold, border, status bar)
- [x] 9.2 Implement auto theme detection from terminal
- [x] 9.3 Implement `?` help popup showing current keymap
- [x] 9.4 Implement ASCII fallback auto-detection for terminals without Unicode box-drawing support

## 10. Testing

- [x] 10.1 Test article upsert deduplication (same GUID not duplicated, content updated)
- [x] 10.2 Test read-status persistence across simulated restarts
- [x] 10.3 Test refresh gate logic (startup 15-min, manual 60-second cooldown, bypass behavior)
- [x] 10.4 Test tag count formatting (all-unread collapse, all-read collapse, mixed, bold styling)
- [x] 10.5 Test read-status rules (short article on open, long article on first scroll)
- [x] 10.6 Test keybinding validation (conflict detection, missing config defaults)
- [x] 10.7 Test tag filtering (filter applied, filter cleared, list reverts)
- [x] 10.8 Test HTML-to-text rendering (links preserved, images as [alt])

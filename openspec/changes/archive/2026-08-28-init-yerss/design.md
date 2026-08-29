## Context

Greenfield Go project. No existing code, no existing specs. The application is
a terminal RSS reader built on `bubbletea` (Elm architecture: Model/Update/View),
`gofeed` (RSS/Atom parsing), and `glebarez/go-sqlite` (pure-Go SQLite, no CGO
toolchain required for end users). See proposal.md for motivation and scope.

## Goals / Non-Goals

**Goals:**
- Define the TUI state machine and rendering approach for all three views.
- Specify the SQLite schema and storage strategy.
- Document the refresh lifecycle and cooldown logic.
- Resolve the custom article-border rendering approach.
- Lock down the dependency set and configuration shape.

**Non-Goals:**
- Inline image rendering or braille image fallback (documented as future work).
- In-app feed management (feeds are edited in the plain-text file).
- Search functionality.
- Network sync or multi-device support.

## Decisions

### D1: Bubbletea Elm architecture with a single state enum

The application uses a single `Model` struct holding all state, with a `viewState`
enum (`listView`, `articleView`, `tagPopup`) and an optional `popupMode`
(`listTags80`, `articleTags40`). `Update` dispatches on `viewState` first, then
on key. `View` renders based on `viewState`, with popups drawn as overlays.

**Alternatives considered:** Separate models per view with a stack. Rejected —
Bubbletea's single-model pattern is simpler and the state space is small (3
views, 2 popup modes).

### D2: Custom article border rendering (not lipgloss borders)

lipgloss cannot inline text into a border line or render a double-line right
edge with per-cell coloring. The article view border is hand-built:

- **Top line**: `┌` + `─ <date> · <title> ` + `─` repeated to fill width + `╖`
- **Bottom line**: `└` + `─` repeated + ` <N>% scrolled ` + `─` repeated + `╜`
- **Left edge**: `│` (single vertical)
- **Right edge**: `║` (double vertical) — rendered column-by-column: top
  `scrollPercent%` cells in accent color, rest in dim color
- **Corners**: `┌` `╖` `└` `╜` (single top/bottom, double right)
- **Content**: `bubbles/viewport` with width = `terminalWidth - 2*paddingX - 2`

ASCII fallback: `─│┌└` → `-|++`, `╖╜` → `++`, `║` → `#` (filled) / `:` (unfilled).

**Alternatives considered:** `bubbles/viewport` with lipgloss border style.
Rejected — cannot inline title/date or render the progress-fill right edge.

### D3: Pure-Go SQLite (glebarez/go-sqlite)

`glebarez/go-sqlite` is a pure-Go driver backed by `modernc.org/sqlite`. No CGO
toolchain needed on the user's machine, which simplifies `go install`
distribution. The cost is a slightly larger binary (~2-4 MB), acceptable for a
desktop TUI application.

**Alternatives considered:** `mattn/go-sqlite3` (CGO-based). Rejected — requires
a C compiler at build time, complicating installation for end users.

### D4: Read-status as "intent to read," not scroll completion

The read flag is an "intent to read" signal: opening a short article (fully
visible) or making any downward movement on a long article marks it read
instantly. The progress bar displays `% scrolled` (positional), not `% read`
(semantic), to avoid implying the user must scroll to 100% to mark read.

### D5: Refresh gate architecture

A single `lastRefreshedAt` timestamp gates both startup and manual refresh:

- **Startup**: if `now - lastRefreshedAt > 15m`, fetch all feeds; otherwise
  load from DB only.
- **Manual** (`R`/`^R`/`F5` on list view): bypasses the 15-min gate but checks
  a 60-second cooldown. If `now - lastRefreshedAt < 60s`, show transient
  "next allowed in Xs" status message; otherwise fetch and update
  `lastRefreshedAt`.

The cooldown message replaces the normal status bar content for ~3 seconds,
then reverts.

### D6: Tag count display rules

```
read == 0 (all unread):   tagname (N)        ← N in BOLD
unread == 0 (all read):   tagname (N)        ← N in plain
mixed:                    tagname (U/N)      ← U in BOLD, N in plain
  where U = unread count, N = total count
```

Since `read = total - unread` is implied, only `unread/total` is shown.

### D7: HTML-to-text with link preservation

`jaytaylor/html2text` converts article HTML to plain text, preserving paragraph
breaks and rendering link URLs inline. Images are replaced with `[alt]` (or
`[image]` if no alt text). The resulting text is re-wrapped to the viewport
width.

**Future work (documented, not implemented):**
- Inline image display via `blacktop/go-termimg` (auto-detects Kitty/Sixel/
  iTerm2 protocols; has Bubbletea integration).
- Braille image fallback via `glibsm/dots` (2×4 dot matrix per cell).
- Detection method: check `$TERM_PROGRAM`, `$TERM`, `$COLORTERM` env vars;
  send `CSI > c` / `CSI c` queries for Sixel support.

### D8: SQLite schema

```
feeds(
  url              TEXT PRIMARY KEY,
  title            TEXT,
  last_fetched_at  DATETIME
)

articles(
  id              INTEGER PRIMARY KEY,
  feed_url        TEXT NOT NULL REFERENCES feeds(url),
  guid            TEXT NOT NULL,
  title           TEXT,
  link            TEXT,
  author          TEXT,
  published_at    DATETIME,
  content         TEXT,
  description     TEXT,
  read            BOOLEAN DEFAULT 0,
  fetched_at      DATETIME,
  UNIQUE(feed_url, guid)
)

categories(
  id              INTEGER PRIMARY KEY,
  name            TEXT UNIQUE
)

article_categories(
  article_id      INTEGER REFERENCES articles(id),
  category_id     INTEGER REFERENCES categories(id),
  PRIMARY KEY(article_id, category_id)
)
```

### D9: Configuration shape

```toml
[data]
dir = ""              # empty = XDG_DATA_HOME/yerss; else override
feeds_file = ""       # empty = XDG_CONFIG_HOME/yerss/feeds.txt

[refresh]
min_interval = "15m"  # startup gate
cooldown = "60s"      # manual refresh cooldown

[display]
theme = "auto"        # auto | dark | light
ascii = false         # ASCII border fallback
padding_x = 2
padding_y = 1

[keybindings]
quit = ["q", "ctrl+c"]
refresh = ["R", "ctrl+r", "f5"]
open_article = ["enter", "l"]
back = ["esc", "enter", "h"]
# ... etc, one entry per action
```

Keybindings use a fixed `Action` enum; users can rebind keys to actions but
cannot define new actions. Validation rejects duplicate key assignments.

## Risks / Trade-offs

- **[Custom border rendering complexity]** → The hand-built article border is
  the most complex rendering code. Mitigation: isolate in a dedicated renderer,
  unit-test glyph placement with fixed-width inputs.
- **[Terminal compatibility variance]** → Box-drawing support, truecolor, and
  key encoding vary across terminals. Mitigation: ASCII fallback for borders;
  use `bubbletea`'s key matching which normalizes terminal differences; test on
  xterm, iTerm2, Kitty, tmux.
- **[Pure-Go SQLite performance]** → `modernc.org/sqlite` is slower than CGO
  SQLite for heavy workloads. Mitigation: yerss is a single-user reader with
  modest data volume; performance is not a concern.
- **[gofeed category inconsistency]** → Different feeds populate categories
  differently (some in `<category>`, some in `<dc:subject>`, some not at all).
  Mitigation: gofeed normalizes this; document that some articles may have no
  tags.
- **[Large binary size]** → Pure-Go SQLite + multiple Charm deps produce a
  larger binary. Mitigation: acceptable for a desktop TUI; can use `-ldflags
  -s -w` and UPX if needed.

## Open Questions

None — all decisions are locked from the exploration phase.

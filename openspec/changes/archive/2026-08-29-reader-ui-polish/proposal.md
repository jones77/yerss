## Why

The reader display is functional but rough: list rows show seconds nobody needs
and hide which publication a story came from; the article frame's left edge
reads as a hard white bar against the grey top/bottom, the scrollbar is a red
double line that clashes with the dim aesthetic, and the bottom border's
"X% scrolled" gives no sense of position within the article. Inline links render
as `text ( url )` with the URL trailing awkwardly, the help popup prints raw
action constants like `open_article`, and tag/refresh keys only work with the
shifted uppercase letter. A coordinated polish pass makes the reader feel
finished and consistent.

## What Changes

### List view rows

- Show publication time as `HH:MM` (drop the seconds).
- Insert a 12-character source-domain identifier between the title and the
  time, derived from the article's link host with the TLD stripped and truncated
  to 12 characters (e.g. `newrepublic.com` → `newrepublic`,
  `reallylongnewspaperdomainname.net` → `reallylongne`).
- Separate title, source, and time with a middle-dot bullet (`·`, ASCII `.`),
  matching the bullet already used in the article top border.
- Render the source identifier and the time in the dim/grey style in all cases;
  only the title changes from bold/white (unread) to dim (read).

### Article view

- Top border date becomes `YYYY-MM-DD HH:MM:SS` (add the publication time in the
  user's local timezone), followed by the bullet and title.
- Bottom border replaces `X% scrolled` with `X% · <bottomLine>/<totalLines>`,
  where `<bottomLine>` is the line number of the last visible viewport line
  (clamped to the total) and `<totalLines>` is the article's total line count
  (e.g. at the very bottom: `100% · 120/120`).
- Add a permanent, left-aligned help hint `o: open in browser` to the bottom
  border, with the percent/line indicator on the right and dashes between.
- Restyle the frame: the left vertical border is rendered in the same grey as
  the top/bottom borders (content text stays as-is); the right edge track becomes
  a grey single line; the scrollbar thumb becomes a bright double line (was a
  red/accent double line).
- ASCII fallback mirrors this: left and right borders are `:` (grey), the
  scrollbar thumb is `|` (white, was `#`).
- Render inline links as markdown `[text](url)`. The link text stays wrapped in
  OSC 8 so Cmd/Ctrl-clicking it opens the browser; the `(url)` is visible plain
  text. URLs that do not fit on the current line break onto a new line, and a
  URL too long for a single line is truncated with an ellipsis (`…`, ASCII
  `...`).

### Help popup

- Render the left-hand action labels with underscores replaced by spaces and in
  Title Case (e.g. `open_article` → `Open Article`, `mark_all_read` → `Mark All
  Read`).

### Keybindings

- Make shifted-letter default bindings reachable via both cases where no
  conflict exists: `T`/`t` both open the tag popup, `R`/`r` both refresh. `G`
  (bottom) stays uppercase only because lowercase `g` is already `top`.
  Keybinding parsing is already case-sensitive; this only widens the defaults.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Article list rendering" requirement is modified for the new
  row layout (source identifier, `HH:MM`, bullet separators, color rules). The
  "Article reader border rendering" requirement is modified for the restyled
  borders (left grey, right grey single-line track, bright double-line thumb),
  the ASCII fallback glyphs, and the bottom-border line indicator plus permanent
  help hint. The "Article top border title truncation" requirement is modified
  to include the publication time in the date prefix. The "Article content
  rendering" and "Clickable links in article content" requirements are modified
  for the markdown link format with OSC 8 clickability and URL wrap/truncation. A
  new "Help popup" requirement is added for the action-label formatting.
- `configuration`: The "Configurable keybindings" requirement is modified to
  reflect the widened default bindings (`t`/`T`, `r`/`R`) and the case-sensitivity
  already in effect.

## Impact

- **List rendering**: `internal/ui/list.go` — row composition (source, bullet,
  `HH:MM`), color rules; domain extraction helper (host parsing, TLD strip,
  truncate to 12).
- **Article border**: `internal/ui/border.go` — left border styling, right
  track/thumb glyphs and colors, ASCII fallback (`:`/`|`), bottom-border layout
  (help hint + line indicator), top-border date format; a `bullet` glyph added
  to `borderGlyphs` with ASCII fallback per the glyph-pair convention in
  `AGENTS.md`.
- **Article view**: `internal/ui/article.go` — pass published time to the border
  renderer; compute bottom-line/total-line indicator from viewport state.
- **HTML/link rendering**: `internal/ui/html.go` — produce `[text](url)` markdown
  (disabling html2text's own `( url )` append via `OmitLinks`), wrap the link
  text in OSC 8, and apply URL break/truncate-with-ellipsis logic in `wrapText`.
- **Help popup**: `internal/ui/popup.go` — format action labels (underscore →
  space, Title Case).
- **Keybindings**: `internal/config/keys.go` — add `t` and `r` to the relevant
  default bindings.
- **Dependencies**: none new — `charmbracelet/x/ansi` (OSC 8, width, truncate)
  and `net/url` (host parsing) are already available.
- **Tests**: list row layout/colors and domain extraction; border glyphs and
  bottom-border indicator; markdown link rendering with OSC 8 and URL
  truncation; help label formatting; widened default keybindings.

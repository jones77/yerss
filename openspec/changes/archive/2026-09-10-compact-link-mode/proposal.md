## Why

Article body links render as `text url` — the visible URL wastes columns and
reads noisily, and the URL text is redundant for clickability (OSC 8 carries
the target regardless). Glamour has no config switch to hide the URL, so we
need our own compact/expanded mode. **Compact mode is the default**: readers
see only the clickable link name unless they opt to reveal the full URLs.

## What Changes

- Body hyperlinks render **compact by default**: the styled link name only, in
  the status-bar blue role, still wrapped in OSC 8 so both Cmd/Ctrl-click and
  the app's left-click open the URL.
- The `x` key binding in the article view toggles **expanded** mode, showing
  the link name followed by the full URL (`text url`), restoring the current
  behavior for readers who want to inspect targets. This is a default keybinding
  from the action catalog, available with zero configuration — exactly like the
  list view's expand toggle, with no explicit config file line required.
- The article header URL line is fixed to wrap at the content width (it
  currently overflows the viewport because it bypasses glamour's word wrap) and
  renders in the status-bar blue role, matching the status bar and title.
- Considered and deferred: a **hover tooltip** revealing the link's URL when the
  mouse hovers a link in compact mode. Feasible — the app already enables
  `tea.WithMouseCellMotion()` and the article view tracks `MouseActionMotion`
  for drag selection, so a transient overlay in `View` keyed on the hovered
  cell can show the URL. Not in scope for this change; re-open if wanted.

## Capabilities

### New Capabilities

- `compact-link-mode`: the article reader's link display modes — compact
  (name only) and expanded (name + URL) — their default state, the `x` toggle,
  and the clickability guarantees that hold in both modes.

### Modified Capabilities

- `reader-ui`: the "Clickable links in article content" and "Clickable article
  URL in reader header" requirements change — body links default to compact
  name-only rendering in the blue role with an expanded mode, and the header
  URL wraps within the content width in the blue role.

## Impact

- `internal/ui/compose/markdown.go`: glamour style config — set `Link` and
  `LinkText` to the status-bar blue role; the compact/expanded toggle drives
  whether the rendered output is post-processed to drop the URL href span.
- `internal/ui/article.go`: header URL rendering — wrap the stripped URL at the
  content width; header line color from the existing status style.
- `internal/config/keys.go`: new article-view action (e.g. `LinkExpandToggle`)
  bound to `x` in `ViewArticle`, added to the action catalog so the help popup
  and seeded config pick it up.
- `internal/ui/article.go` / `articleState`: the toggle state joins the
  derivation cache key so pressing `x` recomposes with the cached body.
- `internal/ui/article_selection.go`: `parseLinkSpans` / `linkAtContentCell`
  already hit-test OSC 8 spans, so clickability works unchanged in both modes.
- Optional: a small upstream glamour patch exposing `SkipHref` (as
  `HideLinkURL`) to suppress the href span natively instead of post-processing.
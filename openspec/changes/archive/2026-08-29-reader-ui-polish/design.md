## Context

The reader UI renders the article list in `internal/ui/list.go` (rows of title +
`HH:MM:SS`, bold-unread/dim-read) and the article view in `internal/ui/article.go`
+ `internal/ui/border.go`. The article frame (`renderArticleBorder`) draws a
top border with `YYYY-MM-DD · title`, a bottom border with a centered
`X% scrolled`, a left vertical border that is emitted as a raw unstyled glyph
(so it renders in the terminal's default bright white), and a right edge that is
a double-line scrollbar (`║`) with the thumb in the accent (red) style and the
track in the dim style. ASCII fallback uses `v="|"`, `fill="#"`, `unfill=":"`.

Inline links currently render as `text ( url )`: `wrapLinks` in `html.go` keeps
the `<a>` tag and wraps its inner text in OSC 8, and `html2text` appends
`( url )` as display text. Width math in `html.go` already uses the ANSI-aware
`ansi.StringWidth`/`ansi.Truncate` (landed in the mouse-and-clickable-links
change). The help popup (`internal/ui/popup.go`) prints action identifiers
verbatim (e.g. `open_article`). Default keybindings live in
`internal/config/keys.go`; parsing is already case-sensitive
(`normalizeKey` preserves case, `R` ≠ `r`).

The palette (`internal/ui/theme.go`) exposes `Accent`, `Dim`, `Bold`, `Border`,
`StatusBar`. `Border` is the grey used by the top/bottom borders
(`#3b4261` dark / `#a0a0a0` light).

## Goals / Non-Goals

**Goals:**
- Reshape the list row to title · source · `HH:MM` with a stable color rule.
- Restyle the article frame so the left and right track read as grey and the
  scrollbar thumb reads as bright, in both Unicode and ASCII modes.
- Replace the bottom-border percent label with a percent + line-position
  indicator and a permanent open-URL hint.
- Render inline links as markdown `[text](url)` with the text still OSC
  8-clickable, and wrap/truncate URLs sensibly.
- Title-Case the help-popup action labels.
- Widen shifted-letter default bindings to both cases where conflict-free.

**Non-Goals:**
- Changing the read-status marking rules or scroll mechanics.
- Altering the tag popup, popup overlay composition, or quit/utility keys
  beyond the label formatting.
- Public-suffix-list–based registrable-domain extraction (see Open Questions).
- Making the scrollbar thumb theme-aware beyond "bright white" (see Risks).

## Decisions

### D1: Source identifier = host minus TLD, `www.` stripped, truncated to 12

Derive the identifier from the article's link URL host (`net/url`), falling back
to the feed URL host when the article has no link. Strip a leading `www.`,
then take the label before the final dot (the second-level domain), then
truncate to 12 characters. This matches every given example
(`newrepublic.com` → `newrepublic`, `sueddeutsche.de` → `sueddeutsche`,
`reallylongnewspaperdomainname.net` → `reallylongne`) with no new dependency.

**Alternative considered:** registrable-domain extraction via a public suffix
list (eTLD+1) to collapse `feeds.nytimes.com` → `nytimes`. Rejected for this
pass — it adds a dependency and edge cases (`.co.uk`); see Open Questions.

### D2: Row layout and color rule

Compose the row as `title · source · time` using the existing middle-dot bullet,
formalized as a `bullet` glyph in `borderGlyphs` (`·` / ASCII `.`) per the
AGENTS.md glyph-pair convention (the top border already uses `·` literally; this
makes it consistent and gives the row a shared separator). Render `source` and
`time` with the dim (`p.Dim`) style always; render `title` bold+`p.Bold` when
unread and dim when read. The selection background continues to apply across the
whole row. The time format switches from `15:04:05` to `15:04`.

### D3: Left border grey; right track grey single, thumb bright double

Style the left vertical `g.v` with `p.Border` (matching top/bottom) instead of
emitting it raw. Change the right edge: the unfilled track uses a single-line
glyph (`│` Unicode / `:` ASCII) in `p.Border` (grey), and the thumb uses the
double-line glyph (`║` Unicode / `|` ASCII) in a bright white style. This
replaces the current accent/red double-line thumb and dim double-line track. The
thumb-height and positioning math in `scrollState.thumb` is unchanged; only the
glyphs and per-segment styles change.

**Bright thumb color:** render the thumb with a bright white foreground
(`#ffffff`, bold) in both themes, per the explicit "bright"/"white" request. See
Risks for the light-theme contrast caveat.

### D4: Bottom border = left hint + right indicator

Replace the centered `X% scrolled` with a left/right layout: `o: open in
browser` left-aligned, `<percent>% · <bottomLine>/<totalLines>` right-aligned,
horizontal dashes filling between. `<percent>` reuses `scrollState.percent()`.
`<bottomLine>` = `min(YOffset + viewportHeight, totalLines)`; `<totalLines>` =
`viewport.TotalLineCount()`. The bullet between percent and ratio is the shared
`bullet` glyph. The corners (`bl`/`br`) remain.

### D5: Top border date prefix includes local time

Change the top-border date prefix from `2006-01-02` to
`2006-01-02 15:04:05` using `PublishedAt.Local()`, so the article border's date
matches the list view's local-time convention. The bullet and title-truncation
behavior are unchanged; only the prefix width grows, which the existing
truncation math (`titleW = coreW - width(prefix) - 1`) already accounts for.

### D6: Markdown links with OSC 8 on the text, link-aware wrapping

Switch link rendering to markdown. In `wrapLinks`, replace
`<a href="url">inner</a>` with
`ansi.SetHyperlink(url) + "[" + inner + "]" + ansi.ResetHyperlink() + "(url)"`
and set `html2text.Options{OmitLinks: true}` so `html2text` no longer appends its
own `( url )`. The OSC 8 sequences wrap the `[text]` portion (clickable text),
and `(url)` is plain fallback text — satisfying both the markdown format and the
clickable-links requirement. `replaceImages` still runs first, so an image-only
link yields `[` + `[alt]` + `](url)`.

URL wrap/truncation is handled by making `wrapText` link-aware: when a token
containing a markdown link `](url)` does not fit on the current line, break at
the `]` boundary so `[text]` stays (if it fits) and `(url)` starts the next line.
If a `(url)` alone is wider than the line width, truncate the URL with
`ansi.Truncate` + the ellipsis glyph (`…` / ASCII `...`) so it fits one line.
Width math stays ANSI-aware so the OSC 8 bytes do not corrupt the break points.

**Alternative considered:** insert a zero-width break opportunity between `]`
and `(url)` and keep `wrapText` whitespace-based. Rejected — `strings.Fields`
splits on whitespace only, so a zero-width marker would not create a break
point, and a visible space would break the markdown look.

### D7: Help labels via action-name transform

Add a helper that maps an `Action` to its display name: replace `_` with spaces
and apply `strings.Title` (Title Case). The help popup (`renderHelp`) uses this
for the left column instead of printing the raw action string. No change to the
key-strings column.

### D8: Widen shifted defaults only where conflict-free

In `DefaultKeybindings`, add `t` to `TagPopup` and `r` to `Refresh`. Do not add
`g` to `Bottom` (lowercase `g` is `Top`). Parsing is already case-sensitive, so
`t` and `T` mapping to the same action is valid and conflict-free; `Validate`/
`EffectiveKeys` need no changes. The `?` help binding stays uppercase only (it is
a shifted symbol, not a letter, and `/` is not a useful help alias).

## Risks / Trade-offs

- **[Bright-white thumb low contrast on light themes]** → The thumb is rendered
  bright white in both themes per the explicit request. On a light background a
  white thumb may be hard to see. Mitigation: document as a known limitation;
  a follow-up could make the thumb color theme-aware (e.g. `p.Bold`).
- **[Subdomain feeds produce a noisy source identifier]** → `feeds.nytimes.com`
  yields `feeds.nytimes` (truncated) rather than `nytimes`, because D1 strips
  only `www.`. Mitigation: acceptable for the common `www.` and bare-domain
  cases; registrable-domain extraction is deferred (Open Questions).
- **[Link-aware wrapping adds complexity to `wrapText`]** → The break-at-`]`
  rule is a small addition to the existing word-wrap. Mitigation: keep the
  detection to a simple `](url)` pattern and unit-test wrap/truncate cases
  explicitly.
- **[Bottom-border layout change may overflow on very narrow widths]** → The
  hint plus indicator need width; on degenerate widths the existing truncate
  helpers keep it within bounds. Mitigation: reuse `truncate`/`padRight` and
  the degenerate-dimension render-safety tests.

## Open Questions

- Should the source identifier collapse arbitrary subdomains (e.g.
  `feeds.nytimes.com` → `nytimes`) via registrable-domain extraction? Deferred —
  the current SLD approach matches all stated examples; revisit if real feeds
  produce ugly identifiers.
- Should the scrollbar thumb be theme-aware (bright on dark, dark on light)
  instead of fixed bright white? Deferred — follow the explicit "white" request
  for now and revisit if light-theme contrast is reported.

## Why

The article view prints the article's URL twice (glamour renders `[url](url)` as
link text plus href), the link is buried under the title/author, and the lead
photo renders full-width with no credit. Navigation also has gaps: there is no
mouse path back to the list, and no way to act on the links inside an article
body without clicking each one.

## What Changes

- **Header composition reordered**: the article URL becomes the very first line
  of the reader, rendered once (bare/autolink form, still OSC 8 clickable),
  followed by a blank line, the bold title, and the author. The duplicated
  `text (url)` rendering of the header link is removed; body links keep the
  `text url` form.
- **Centered lead photo with attribution**: the halfblocks image block moves
  below the header (instead of above everything) and is horizontally centered.
  A one-line attribution is centered directly beneath the photo (no gap): the
  photo's own credit extracted from the article HTML when present, otherwise
  `photo: <org>` derived with the same source-identifier rules as the list
  view. A blank line follows the attribution. The photo + attribution scroll
  atomically as one block.
- **Right-click returns to the list**: a right-button press in the article view
  navigates back to the list view (same behavior as Esc/Enter/h/b).
- **Links popup**: pressing `l` or right-arrow in the article view opens a
  popup listing every link in the article (link text with the URL dimmed and
  truncated). `j`/`k`/arrows navigate, Enter opens the selected link in the
  system browser, Esc closes. Enter continues to close the article view
  (unchanged).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `reader-ui`: reader header composition and single-rendered header URL; lead
  image block placement (below header), centering, and attribution line;
  atomic image scroll range includes the attribution; right-click navigation
  from article to list; new article links popup requirement.

## Impact

- `internal/ui/article.go` — header markdown order, compose-time link
  harvesting, right-click handling.
- `internal/ui/markdown.go` — header link becomes a bare autolink;
  `markdownLink` helper becomes unused and is deleted.
- `internal/ui/image.go` — image block composition moves below the header
  lines; centering padding; attribution line inside the atomic scroll range.
- `internal/ui/popup.go` — links popup mode, state, rendering, and navigation.
- `internal/ui/model.go` — popup dispatch for the links popup.
- `internal/config/keys.go` — new `link_popup` action bound to `l` and `right`
  in the article view; help popup and seeded config template derive from the
  catalog automatically.
- `convert` package (or a small ui helper) — attribution extraction from
  stored article HTML (figcaption / credit class), reusing the existing HTML
  parser.
- No schema changes: content HTML is already stored, attribution is re-derived
  at compose time.
- Tests: `osc8_test`, `article_image_test`, `mouse_test`, `keys_test` updated;
  new tests for attribution extraction, centering, and the links popup.

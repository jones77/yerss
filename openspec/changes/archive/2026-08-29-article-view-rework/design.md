## Context

The article view composes three pieces in `newArticleState` (internal/ui/article.go):
the halfblocks image block, then the header markdown (title, author,
`[url](url)` link), then the converted body. Two constraints shape this design:

- glamour v2's `ansi.LinkElement` always renders the href after the link text
  (`text url`); autolinks are the exception (`SkipText: true` — URL only).
  This is why the header's `[url](url)` prints the URL twice.
- The image block participates in atomic scroll snapping via
  `articleState.imgStart/imgEnd`, currently hardcoded to `0..len(imgBlock)-1`.

The lead image URL comes from feed enclosures only; the stored article HTML is
available at render time and is already re-converted on every compose, so
attribution can be derived from it without schema changes.

## Goals / Non-Goals

**Goals:**

- Header reads top-down: URL (once, OSC 8), blank, bold title, author.
- Photo centered under the header with a centered one-line attribution that
  never separates from the photo while scrolling.
- Right-button press in the article view returns to the list.
- `l`/right-arrow opens a links popup; Enter opens the selected URL in the
  browser; Enter (popup closed) keeps closing the article.

**Non-Goals:**

- No change to body link rendering (`text url` stays — glamour behavior).
- No schema migration, no persisted attribution field.
- No change to the top/bottom border chrome (date/title in the top border stay).
- No image resizing/zoom interactions.

## Decisions

**Header URL as bare autolink, not `[text](url)`.**
Emit the URL as the line content (goldmark detects it as an autolink; wrap in
`<...>` only when it contains spaces/parens). Glamour renders autolinks with
`SkipText: true`, printing the URL exactly once, still OSC 8-hyperlinked.
Alternative considered: a custom glamour style suppressing the href part — not
supported by `StyleConfig` (`SkipHref` is element-level, set only for
table-footer links). Consequence: `markdownLink` in markdown.go becomes dead
code and is deleted along with its tests. Title and author are rendered as
separate glamour documents joined without a blank line: glamour folds line
breaks inside a paragraph into spaces, so a single document would render
"Title by Jane" on one line, and separate paragraphs would insert a blank
between them.

**Composition order via header-lines count.**
`newArticleState` renders the header markdown first, counts its lines, then
asks `articleImageBlock` for the image block and inserts it after the header
with one blank line on each side: `header · blank · [image lines · attribution
line] · blank · body`. `imgStart = headerLines + 1`, `imgEnd = imgStart +
len(block) - 1` so the snap range covers image plus attribution as one unit.
The height cap becomes `maxH = max(1, vpH - headerLines - 4)` — the two blank
lines, the attribution line, and one line of body text are reserved, so the
full block plus at least one body line is visible on open (the spec scenario;
the earlier `-3` figure dropped the body-line reserve). Scroll-position
preservation on late image load (`onImageLoaded`) is unchanged — it already
works on line counts.

**Centering by padding, not renderer changes.**
Render the image at full content width exactly as today (aspect-fit, capped),
then measure the widest rendered line (`imgW`) with `ansi.StringWidth` and
prefix every image line with `(contentW - imgW)/2` spaces so the photo is
centered as a unit. The halfblocks renderer is untouched, so resize/re-render
behavior and the cache stay as-is. The attribution line is truncated to the
photo's own width and centered within the photo's span — its left padding is
`photoPad + (imgW - attrW)/2` — so it stays bounded by the photo even when the
height cap scales the photo narrower than the content area. Edge case:
degenerate widths where a computed pad is negative clamp to 0.

**Attribution extracted at compose time from stored HTML.**
A small helper (in the `convert` package, which already parses HTML) scans the
article HTML for the `figure` whose `img` matches the lead image URL (or the
first `figure` when none matches) and returns its `figcaption` text; if absent,
any element with `credit` in its class is used. Text is collapsed to one line
and truncated. Fallback at render time: `photo: ` + the list view's
`sourceID(article.Link, article.FeedURL)`. Alternatives: persisting credit at
fetch time (schema churn, and the enclosure image often is not in the content
HTML, so fetch-time matching would miss) or deriving everything from the org
(loses real credits).

**Links harvested from the markdown source, not rendered lines.**
At compose time, scan `renderArticleMarkdown(a)` output for `\[...\](...)`
link destinations (the same document the renderer sees), preserving first-
occurrence order and deduplicating by URL; store `{text, url}` pairs on
`articleState`. Harvesting from rendered lines instead would split wrapped
links into duplicate partials. The regex tolerates `<...>`-wrapped
destinations. The header URL is naturally included — acceptable, it is a link
in the article and the popup is the discoverable path to it.

**Links popup reuses the popup machinery.**
New `popupLinks` mode; `popupState` gains a `links []articleLink` field with
its own cursor (the tags slice is unused in this mode). Rendering mirrors
`renderTagPopup`: rounded border, accent title "Links", rows `text · url` with
the URL dim and truncated to the box width. Keys: MoveUp/MoveDown (+wrap) via
the existing popup view mapping, Enter opens the selected URL with
`openURLCmd`, Esc/Back closes. A new `link_popup` action (`l`, `right`) is
added to the catalog for ViewArticle only — `l` is currently unbound in the
article view and `right` is unbound everywhere, so no conflicts; help popup
and seeded config template derive from the catalog automatically.

**Right-click mapping.**
In `updateArticleMouse`, an unmodified right-button press performs the same
transition as Back (view = list, clear selection, `loadList`). Motion/release
events for the right button are ignored. Modifier-carrying presses fall through
the existing modifier guard.

## Risks / Trade-offs

- [Attribution regex/HTML heuristics vary by publisher] → Extraction is best-
  effort by design; the org fallback guarantees a sane line. figcaption is
  matched structurally (not by class guesses) first.
- [Bare-URL autolink detection is terminal-width dependent] → goldmark's
  autolink parsing is textual, not width-dependent; only URLs with spaces or
  parens need the `<...>` form, which the writer already handles.
- [Popup rows with very long URLs] → URL is truncated to the box width; the
  full URL is what gets opened, so nothing is lost functionally.
- [Header reflow shifts `imgStart` bookkeeping] → The snap range and late-load
  offset math both derive from the same compose-time counts; tests cover the
  new offsets (`article_image_test`).
- [Right-click in terminals reporting button-release quirks] → Only Press
  events act; release is ignored, matching how wheel events are already
  handled.

## Migration Plan

No data migration. Rollback is a plain revert; no stored state changes shape
(attribution is re-derived per render, links are per-session).

## Open Questions

(none — interaction details settled during exploration)

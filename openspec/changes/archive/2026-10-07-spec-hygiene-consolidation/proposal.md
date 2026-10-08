## Why

The `openspec/specs/` tree has drifted from the implemented behavior. An audit found
three outright contradictions (two specs assert opposite facts about the same behavior),
five cases of the same requirement stated in two-to-four places, and four smaller
documentation errors — all of it traceable to the image and link behavior being split
across nine specs plus large sections of the 1,991-line `reader-ui` spec. Every drift
item was verified against the Go source: the code is the source of truth, and in each
contradiction one spec already matches it while the other is stale. No behavior is
changing; this change makes the specs truthful and single-sourced.

## What Changes

- **Resolve three contradictions** so each behavior has one canonical, code-accurate statement:
  - Lead-image attribution: drop the "document-global credit fallback" that `reader-ui`
    still claims (`inline-image-rendering` and `internal/convert/credit.go` both say there
    is none).
  - Image storage schema: drop `feed-pipeline`'s stale "add `image_url` + `image_data`
    columns" migration and `image_data`-on-article persistence (`image-block-storage`'s
    unified `article_images` table is current).
  - ASCII border glyphs: `configuration` lists `|` as a border character; the vertical
    border is actually `:` and `|` is the scrollbar thumb (per `render/border.go`).
- **Fix smaller drift / documentation errors**:
  - `reader-ui`: `photo: intercept` example → `photo: theintercept` (source-identifier rule).
  - `configuration`: document the `link_expand_toggle` action (bound to `x` in the article
    view), the full refresh key set (`R`, `r`, `ctrl+r`, `f5`), and the correct ASCII glyphs.
  - `copy-debug-info`: name the default binding (`D`, article view).
- **Consolidate** the nine fragmented image/link specs into two canonical specs
  (`image-rendering` and `image-loading`), slimming `reader-ui` down to a UI shell that
  references them and making `compact-link-mode` the single home for link-display behavior.
- **No implementation code changes.** This is a pure specification hygiene change; the
  affected behavior is already implemented and tested.

## Capabilities

### New Capabilities

- `image-rendering`: What images look like and how they are identified and credited —
  lead and inline image discovery, unified attribution (figcaption / alt / link-text
  fallback chain with **no** document-global credit fallback), lead-image dedup, the
  2048px decode-width cap, caption-width fitting, and halfblock + native (iTerm2/kitty)
  rendering with block placeholders and placement cleanup.
- `image-loading`: How images get into view asynchronously — the unified `article_images`
  storage schema and legacy-column purge, photo-byte persistence, viewport-driven load
  frontier with bounded concurrency, lazy native render, the decode-once cache, the
  derivation cache and recomposition bounds, and atomic scroll-snap invariants.

### Modified Capabilities

- `reader-ui`: remove the lead-image, inline-image, attribution, link-display,
  header-URL, async-image-lifecycle, and scroll-snap requirements (they move to
  `image-rendering`, `image-loading`, and `compact-link-mode`); keep list, borders,
  scrollbar, title truncation, keyboard/mouse navigation, read-status, tag/help/links
  popups, content rendering, restore state, render safety, selection, and color roles.
- `feed-pipeline`: remove the stale "additive image-column migration" and
  `image_data`-on-article persistence requirements (image storage is owned by
  `image-loading`); keep feed lifecycle, dedup, tagging, and verification.
- `configuration`: add `link_expand_toggle` to the keybinding catalog, reconcile the
  refresh key set to `R`/`r`/`ctrl+r`/`f5`, and correct the ASCII border glyph list
  (vertical border is `:`; `|` is the scrollbar thumb).
- `compact-link-mode`: become the single canonical home for link-display behavior,
  absorbing the expanded header-URL and in-body hyperlink requirements from `reader-ui`.
- `copy-debug-info`: document the default binding `D` (article view).

### Removed Capabilities

Merged into `image-rendering`:

- `native-image-rendering`
- `inline-image-rendering`
- `lead-image-dedup`
- `image-decode-cap`

Merged into `image-loading`:

- `image-block-storage`
- `image-load-scheduling`
- `article-recomposition`

## Impact

- **Specs only.** No Go source, API, dependency, or database-schema changes.
- `openspec/specs/` tree is restructured: 13 capabilities become 8
  (`reader-ui`, `feed-pipeline`, `configuration`, `compact-link-mode`,
  `image-rendering`, `image-loading`, `article-json-export`, `copy-debug-info`).
- Downstream consumers of the specs (agents following `openspec/specs/` for behavior
  requirements) will find image/link behavior in one place instead of nine.

## Why

RSS articles carry a publisher-attached lead image (`<enclosure>` /
`media:thumbnail`) that yerss currently discards, rendering only `[alt]`
placeholders for inline body images. Modern terminals (Ghostty, Kitty, iTerm2)
can display real images, and `blacktop/go-termimg` provides multi-protocol
rendering with a universal Unicode halfblocks fallback. Surfacing the lead
image makes the reader view match the web-article mental model where a photo
scrolls with the text, with no privacy exposure beyond what fetching the feed
already incurs (enclosure images are publisher-hosted, same origin).

## What Changes

- Capture an article-level image URL from `<enclosure>` / `media:thumbnail`
  during feed parsing and persist it in a new nullable `articles.image_url`
  column. Inline content `<img>` URLs are **not** captured (publisher-hosted
  only).
- Render the captured image **inline** at the top of the article body so it
  scrolls with the content, via `blacktop/go-termimg`'s halfblocks renderer
  (Unicode half-block characters + ANSI color — text-grid content that the
  existing `bubbles/viewport` handles natively).
- Investigate go-termimg's Kitty Unicode-placeholder mode as an upgrade to
  native-protocol (photographic) inline rendering; integrate it only if a
  gated spike passes, otherwise ship halfblocks-only and defer native.
- Scroll atomically past the image block: the image is either fully visible or
  fully scrolled off, never partially clipped. A downward scroll that would
  partially clip the image snaps past it in one keystroke.
- Add a `Display.Images` config option (`auto` / `on` / `off`, default `auto`)
  controlling image rendering. The existing `-a` / `--ascii` flag forces
  images off (halfblocks are Unicode, incompatible with ASCII-only mode).
- Fetch images asynchronously as `tea.Cmd` messages (timeout, max-bytes cap,
  `image/*` content-type allowlist). Persist the fetched image bytes in a new
  nullable `articles.image_data` BLOB column so re-opens, resizes, and
  application restarts never re-fetch (cache hierarchy: in-memory decoded →
  DB bytes → network; the image host is contacted exactly once per article).
  On fetch failure, no image block is rendered and no bytes are persisted
  (graceful fallback).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `reader-ui`: The "Article content rendering" requirement is modified so
  publisher-attached (enclosure) images render inline as actual images when
  enabled, while inline content `<img>` elements remain `[alt]`/`[image]`
  placeholders. New requirements cover inline image rendering, atomic scroll
  behavior, and the async load/resize lifecycle.
- `feed-pipeline`: The "Article persistence across restarts" requirement is
  extended to include the new nullable `image_url` and `image_data` columns
  and their guarded additive migration. Feed parsing captures
  enclosure/media-thumbnail URLs; fetched image bytes are persisted and
  restored across restarts.
- `configuration`: The "ASCII border fallback" requirement is extended so
  `--ascii` also forces image rendering off. A new requirement covers the
  `Display.Images` option and its interaction with terminal detection.

## Impact

- **Dependencies**: Adds `github.com/blacktop/go-termimg` (MIT, actively
  maintained, Bubbletea-aware widgets). Pulls in `charmbracelet/x/mosaic` for
  halfblocks rendering.
- **Data layer**: `internal/store` — new `Article.ImageURL` field and image-
  byte storage, new `articles.image_url TEXT` and `articles.image_data BLOB`
  columns with a guarded additive migration (the project's first; uses
  `PRAGMA table_info` since SQLite lacks `ADD COLUMN IF NOT EXISTS`). New
  `image_url`/`image_data` handling in insert/upsert and image read/write
  query paths.
- **Feed parsing**: `internal/feed/fetch.go` — `itemToArticle` extracts image
  URL from `item.Image` / image-typed `item.Enclosures`.
- **New package**: `internal/image` — async fetch `tea.Cmd`, cache hierarchy
  (in-memory decoded → DB bytes → network), guard logic
  (timeout/max-bytes/content-type), `imageLoadedMsg` / `imageFailedMsg`
  messages.
- **UI layer**: `internal/ui/article.go` — compose image block into article
  content; handle load/resize/failure messages; preserve scroll position on
  async insert. `internal/ui/render` path — `snapYOffset` atomic-scroll logic
  applied after each scroll mutation. Renderer behind an interface seam for
  testability.
- **Config**: `internal/config` — new `Display.Images` field and TOML option.
- **Tests**: URL extraction, cache hit/miss, config/ascii gating, composition
  ordering, `snapYOffset` pure-function table tests, scroll preservation,
  fetch-failure fallback, resize re-render.

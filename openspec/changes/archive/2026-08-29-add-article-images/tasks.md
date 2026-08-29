## 1. Data layer

- [x] 1.1 Add `ImageURL string` field to `store.Article` struct (`internal/store/store.go`)
- [x] 1.2 Add guarded additive migration: query `PRAGMA table_info(articles)`, add `image_url TEXT` and `image_data BLOB` columns only when each is absent (`migrate()` in `internal/store/store.go`)
- [x] 1.3 Update article insert/upsert query paths to persist `image_url`
- [x] 1.4 Update `GetArticle`/list scan paths to read `image_url` into the struct
- [x] 1.5 Add `SetImageData(id int64, data []byte)` and `GetImageData(id int64) ([]byte, error)` store methods to write/read the `image_data` BLOB
- [x] 1.6 Test: columns added on first migration, idempotent on re-run, existing rows retain data with null image URL/data; `SetImageData`/`GetImageData` round-trip

## 2. Feed parsing

- [x] 2.1 Extract lead image URL in `itemToArticle` from first image-typed `item.Enclosures` (MIME `image/*`), falling back to `media:thumbnail` extension
- [x] 2.2 Confirm inline content `<img>` URLs are NOT captured (only enclosure/thumbnail metadata)
- [x] 2.3 Test: enclosure image URL captured, media thumbnail captured, non-image enclosure ignored, inline `<img>` not captured

## 3. Configuration

- [x] 3.1 Add `Images string` field to `config.Display` struct with default `"auto"` (`internal/config/config.go`)
- [x] 3.2 Add `images` TOML option under `[display]` in `displayOpts` and load logic
- [x] 3.3 Make `-a`/`--ascii` flag force image rendering off, taking precedence over config `auto`/`on`
- [x] 3.4 Treat unrecognized `images` values as `auto`
- [x] 3.5 Test: default auto, explicit off disables, ascii flag overrides on, invalid value falls back to auto

## 4. Image fetch and cache

- [x] 4.1 Create `internal/image` package with an in-memory `map[string]image.Image` decoded cache keyed by URL
- [x] 4.2 Implement load `tea.Cmd` with cache hierarchy: in-memory → DB (`store.GetImageData`) → network fetch; on network success persist raw bytes via `store.SetImageData` before emitting load message
- [x] 4.3 Network fetch guards: HTTP timeout, max-bytes cap, `image/*` content-type allowlist, `image.Decode` validation
- [x] 4.4 Define `imageLoadedMsg{key string, img image.Image}` and `imageFailedMsg{key string}` messages
- [x] 4.5 Test: in-memory hit, DB hit (decode + cache), network miss (fetch + persist), timeout/non-image/oversized produce failure messages

## 5. Article composition

- [x] 5.1 Define a renderer interface seam (returns deterministic strings in a fake) for testability
- [x] 5.2 Render image block to `contentW` via halfblocks (go-termimg `StatefulImageWidget` or `mosaic`), aspect-preserved, height-capped to `viewportH - headerLines - 1`
- [x] 5.3 Splice the pre-rendered image-block lines above the glamour-rendered content in `newArticleState` (the block is raw halfblock lines and is never passed through the markdown renderer)
- [x] 5.4 Confirm glamour wraps the body text only: the image block is pre-sized to `contentW` and never re-wrapped
- [x] 5.5 Test: composition ordering (image, title, author, link, body), image not re-wrapped, height cap enforced

## 6. Update loop and lifecycle

- [x] 6.1 Fire load `tea.Cmd` in `openArticle` when `ImageURL` is set and images are enabled (load resolves via the cache hierarchy)
- [x] 6.2 Handle `imageLoadedMsg`: store decoded image in cache, re-compose content with image block, re-render
- [x] 6.3 Handle `imageFailedMsg`: re-compose without image block (graceful fallback)
- [x] 6.4 On `WindowSizeMsg`: re-render image block at new `contentW` from cached image, re-compose (no re-fetch)
- [x] 6.5 Scroll-position preservation on late load: when `YOffset > imgStart`, bump `YOffset` by inserted line count before re-compose
- [x] 6.6 Test: async open, load-and-appear, failure fallback, resize re-render, scroll preservation, re-open uses in-memory cache, restart uses DB bytes

## 7. Atomic scroll

- [x] 7.1 Implement `snapYOffset(yOffset, vpH, imgStart, imgEnd, direction int) int` as a pure function
- [x] 7.2 Apply `snapYOffset` after each scroll mutation in `updateArticle` (ScrollDown/Up, PageDown/Up, HalfPage*, GotoBottom), then `viewport.SetYOffset`
- [x] 7.3 Test: downward skip-past, upward reveal, page-down skip, no-snap when fully out of view, boundary cases

## 8. Halfblocks path end-to-end

- [x] 8.1 Wire halfblocks renderer as the default render path when images are enabled
- [x] 8.2 Verify image displays inline, scrolls with content, and exits cleanly (no ghost) in a Unicode-capable terminal
- [x] 8.3 Verify `--ascii` mode renders no image block and keeps `[alt]`/`[image]` placeholders

## 9. Native-protocol spike

- [x] 9.1 Assess Kitty Unicode-placeholder transmission/placement separation in go-termimg's API against spike criteria 1-6
- [ ] 9.2 If all criteria pass: wire native (Kitty) as preferred protocol with halfblocks fallback, detect protocol once via `termimg.DetectProtocol()` and cache on Model — **not applicable: spike did not pass (criterion 6, see design.md)**
- [x] 9.3 Document findings in design.md and ship halfblocks-only; native-inline becomes future work

## 10. Validation

- [x] 10.1 Run full test suite (`go test ./...`) and linter; fix any failures
- [x] 10.2 `openspec validate --change add-article-images --strict` passes
- [ ] 10.3 Manual smoke test (requires a human in a real terminal: Ghostty inline render + scroll + atomic snap + resize): open an article with an enclosure image in Ghostty, verify inline render + scroll + atomic snap + resize

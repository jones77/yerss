## 1. Slim the list-load query

- [x] 1.1 Add a lighter list-load path: `ListArticlesLite` (store + session) selects only the columns the list renders (`id, feed_url, guid, title, link, author, published_at, read, fetched_at, image_url` — without `content`/`description`) and populates the `Article` rows with empty body fields. `ListArticles` stays full for the JSON export.
- [x] 1.2 Point `loadList` at `ListArticlesLite`; confirm no other list-path caller reads `Content`/`Description` (the JSON dump keeps using the full `ListArticles`); no store test change needed since `ListArticles` is unchanged.
- [x] 1.3 Run `go test ./internal/store/... ./internal/ui/...` and confirm the slimmer query renders the list identically.

## 2. Derive-once article composition cache

- [x] 2.1 Split `newArticleState` into a derivation stage (`computeDerivation`: ConvertImages, inline discovery/dedupe, link harvesting, per-segment glamour rendering, header) and a composition stage (`composeArticle`/`spliceBody`: image blocks spliced into cached segments, viewport build).
- [x] 2.2 Add a single-entry derivation cache on `Model` keyed by `(articleID, contentW, vpH)` plus content, ascii mode, palette, and glamour theme style (so renderer/ascii/palette changes invalidate it).
- [x] 2.3 Add a regression test that composes a fixture with lead + inline + duplicate-sentinel cases and asserts the cached-derivation output is byte-identical to the pre-change `newArticleState` output.

## 3. Recompose once per image message with the cached derivation

- [x] 3.1 In `Update`, set a "recompose pending" flag for `BlockMsg`, `PhotoMsg`, `NativeMsg`, and `FailedMsg` instead of recomposing inline; after the message switch, run `recomposeArticle` once if the flag is set and the view is still the article view, then clear it (per-message, at most once; the recompose reuses the cached derivation).
- [x] 3.2 Ensure image caches are still updated per message (existing behavior) even when the article view was left mid-batch, so re-opening renders the block.
- [x] 3.3 Add a test that delivers multiple image messages and asserts exactly one recompose per message (recomposeCount), with the derivation cache reused and the reading position preserved.

## 4. Verification

- [x] 4.1 Run the full test suite (`go test ./...`) and the linter; confirm no regressions in scroll/snap, native-clear, or read-state tests.
- [x] 4.2 Manual check against a stored image-heavy article (the ProPublica Helene piece in the local DB): open, confirm the UI stays responsive while images load, and confirm back-out returns to the list promptly. (Pre-existing snap-asymmetry and gallery-caption behaviors observed during the check are out of scope and deferred.)
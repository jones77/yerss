## 1. Header composition

- [x] 1.1 Rewrite `articleHeaderMarkdown` to emit the bare URL (autolink, `<...>`-wrapped only when it contains spaces/parens) as the first line, then a blank line, the bold escaped title, and the `by <author>` line when present
- [x] 1.2 Delete the now-unused `markdownLink` helper and adjust `markdown_test` accordingly
- [x] 1.3 Update `TestRenderArticleHeaderURLIsOSC8` to assert the URL appears exactly once and remains OSC 8-wrapped; add header-order scenario tests (URL first, blank, bold title, author; no-author variant)

## 2. Attribution extraction

- [x] 2.1 Add a credit-extraction helper to the `convert` package: find the `figure` containing an `img` matching the lead image URL (else the first `figure`), return its `figcaption` text; fall back to an element with `credit` in its class; collapse whitespace and cap to one line
- [x] 2.2 Add unit tests for the extractor: figcaption match, first-figure fallback, credit-class fallback, no-credit case

## 3. Image block placement and centering

- [x] 3.1 Move image block composition below the header lines in `newArticleState` with one blank line on each side; set `imgStart`/`imgEnd` to cover image plus attribution line
- [x] 3.2 Center the rendered halfblock lines by padding `(contentW - lineWidth)/2` spaces (ANSI-aware width, clamp at 0) and add the centered, dim attribution line directly beneath
- [x] 3.3 Update the height cap to reserve the attribution line and both blank lines (`maxH = max(1, vpH - headerLines - 4)`, the extra line being the body-text reserve required by the spec scenario)
- [x] 3.4 Update `article_image_test` expectations (block position, offsets, cap) and add tests: centered padding, attribution under image with no gap, attribution fallback `photo: <sourceID>`, blank line after attribution, snap range covering attribution

## 4. Right-click navigation

- [x] 4.1 In `updateArticleMouse`, handle an unmodified right-button press as Back: return to the list view, clear selection, reload the list; ignore right-button motion/release and modifier-carrying presses
- [x] 4.2 Add tests: right-click returns to list with the article selected, no selection anchored

## 5. Links popup

- [x] 5.1 Add the `link_popup` action to the config catalog bound to `l` and `right` in ViewArticle only; confirm no keymap conflicts and that the help popup/seeded template reflect it
- [x] 5.2 Harvest links at compose time from `renderArticleMarkdown` output (link text + URL, document order, dedup by URL) into `articleState`
- [x] 5.3 Add `popupLinks` mode with its own state (links + cursor), rendering text plus dim/truncated URL in the rounded-border box, MoveUp/MoveDown wrap navigation, Enter opens via `openURLCmd`, Esc/Back closes; empty list shows an empty popup and Enter is a no-op
- [x] 5.4 Add tests: open via `l` and `right`, row content, dedup of wrapped links, wrap-around navigation, Enter issues an open command, Esc closes without navigating, empty-article behavior

## 6. Verification

- [x] 6.1 Run `go build ./...` and `go test ./...` — all green
- [x] 6.2 Run `go vet ./...` and the project linter; fix findings
- [x] 6.3 Manual smoke test in a real terminal: header order, centered photo + attribution, right-click, links popup with browser open, ASCII mode (`-a`) glyphs unaffected

## 7. Refinements after smoke test

- [x] 7.1 Remove the blank line between the title and the author: render the header as separate glamour documents (`renderHeader`) joined with the author directly beneath the title; update header-order tests
- [x] 7.2 Constrain the photo attribution to the photo's own width: center the photo as a unit (`photoPad`), truncate the attribution to the photo width, and center it within the photo's span; update centering/attribution tests and sync the delta spec and design

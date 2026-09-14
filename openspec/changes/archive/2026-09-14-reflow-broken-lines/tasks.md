## 1. Blockquote closing punctuation fix

- [x] 1.1 Extend `isOrphanLine` in `internal/ui/compose/markdown.go` so a line whose trimmed visible text is punctuation-only is treated as an orphan when the previous line is a quote-bar line, regardless of the successor; adjust guards (keep the `i == 0` early-return, move the end-of-lines guard after the punctuation check).
- [x] 1.2 Add a compose unit test covering a blockquote whose final word ends in a closing curly quote wrapping at the content width, asserting the closing quote renders on a quote-bar line (no bare punctuation-only line).

## 2. Compact-mode paragraph re-flow

- [x] 2.1 Add `ReflowShortLines(rendered string, contentW int) string` (and the styled trailing-whitespace trim helper) in `internal/ui/compose/`; group consecutive non-blank lines into paragraphs, skip blockquote/list/indented groups, and re-flow only paragraphs with a raw sub-width line using `charm.land/lipgloss/v2` `Wrap` at the content width; glue punctuation-only orphan lines without a space.
- [x] 2.2 Wire `ReflowShortLines` into `computeDerivation` in `internal/ui/article.go` after `StripLinkHrefs`, applied only when `!m.linkExpand`.
- [x] 2.3 Add a compose unit test reproducing the wrapped-URL strip artifact (a paragraph line left below the content width) and asserting the re-flowed paragraph fills the width with the link name preserved and no short line.

## 3. Integration and verification

- [x] 3.1 Verify `go mod tidy` records `charm.land/lipgloss/v2` as a direct dependency; build the module.
- [x] 3.2 Add an article-level test (compact and expanded modes) using the pullquote and wrapped-elclip-link paragraph shapes, asserting no bare punctuation line and no sub-width line in compact mode.
- [x] 3.3 Run `go test ./internal/ui/...` and `go vet ./...`; confirm all existing tests (including compact-link-mode and reader-ui specs' tests) still pass.
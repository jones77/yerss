## 1. Toggle anchoring (blank/URL lines and bottom pinning)

- [x] 1.1 Add a helper that, given the pre-toggle lines and `oldOffset`, returns the first anchorable line's leading text plus the number of leading blank lines skipped (URL-token detection via `m.article.links` per design decision 1)
- [x] 1.2 Rewire `toggleLinkExpand` to use the anchorable-line scan instead of `TrimSpace(Strip(lines[oldOffset]))`, preserving the photo-block anchor path unchanged
- [x] 1.3 Position the post-recompose viewport at `findAnchorLine(anchor) - gap` (clamped), and glue to bottom (`GotoBottom`) when the scan finds no anchorable word
- [x] 1.4 Add tests: blank-first-line toggle keeps the separator on top; expanded URL-first-line toggle near the end stays glued to the bottom; reflowed text re-anchors so the new first line contains the pre-toggle first word

## 2. Compact reflow spacing

- [x] 2.1 Extend the styled-whitespace trimmer in `compose/reflow.go` to strip leading whitespace (and adjacent style/OSC spans) symmetrically with the existing trailing trim
- [x] 2.2 Apply the leading trim to every line in `reflowParagraph`
- [x] 2.3 Add tests: a hidden URL followed by text yields exactly one space (`claimed was`, no `claimed  was`); link name ending in punctuation yields no extra space

## 3. Selection highlight across links

- [x] 3.1 Check x/ansi for an existing re-emit-aware truncate; use it if present, otherwise implement a segment-aware slice on `compose.ScanANSI` that re-opens the active SGR/OSC-8 state at each boundary
- [x] 3.2 Rewrite `highlightSelection` to build `before`/`middle`/`after` with balanced escapes and re-assert `\x1b[7m` after interior full resets, keeping OSC-8 links clickable
- [x] 3.3 Add tests: selecting a region spanning a link keeps the whole region inverted with no early reset at the link; selection that starts or ends mid-link stays balanced

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go test ./internal/ui/...` and fix regressions
- [x] 4.2 Run `openspec validate --strict` on the change and resolve any delta errors
- [x] 4.3 Confirm `git diff` is scoped to the change (no unrelated files)

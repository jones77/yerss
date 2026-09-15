## 1. Expanded link-mode URL inline reflow

- [x] 1.1 Add `compose.InlineLinkURLs` re-flow for expanded link mode
- [x] 1.2 Wire the expanded-link branch of `computeDerivation` to `InlineLinkURLs`
- [x] 1.3 Test long-URL re-flow and per-line clickability

## 2. Compact-mode wrapped-href residual fix

- [x] 2.1 Split `collectLinkSpans`/`groupLinkUnits` and extend the wrapped-href drop through trailing fill + wrap newline
- [x] 2.2 Test that hiding a wrapped URL leaves no whitespace-only residual line

## 3. Content-anchored selection restore

- [x] 3.1 Add `AnchorAt`/`RestoreAt` to `articleview.State`
- [x] 3.2 Rewrite `RecomposeInPlace` insertion accounting (below-fold blocks do not shift the reader)
- [x] 3.3 Record the anchor in `persistSelection` and restore it in `restoreSelection`
- [x] 3.4 Re-apply the anchor on the first real `WindowSizeMsg` via `pendingRestore`
- [x] 3.5 Test anchor restore with image blocks and across a startup resize

## 4. Verification and cleanup

- [x] 4.1 Remove the debug probe test files
- [x] 4.2 Run build, vet, full test suite, and lint; confirm no new issues
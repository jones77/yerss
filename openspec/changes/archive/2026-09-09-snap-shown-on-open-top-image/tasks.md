## 1. Down-scroll snap for the shown-on-open top image

- [x] 1.1 Derive `leadShown` in `composeArticle` from `blocks[0].ImgStart == headerLines+1` instead of only the genuine-lead branch, so a suppressed lead whose URL opens the body as its first inline image is treated as shown on open
- [x] 1.2 Update the `articleState.leadShown` and `SnapYOffset` doc comments to describe the geometry-derived semantics
- [x] 1.3 Add a regression test asserting a suppressed-lead-at-top block rises flush to the viewport top on down-scroll then skips past its caption, and the mid-article promo-banner reuse still snaps like an inline image
- [x] 1.4 Update the `reader-ui` main spec lead paragraph to distinguish the shown-on-open top image from the mid-article promo-banner case

## 2. Up-scroll snap for the shown-on-open top image

- [x] 2.1 Remove the `(i > 0 || !lead)` gate on the upward top-snap so the whole photo-and-caption block snaps flush to the viewport top as soon as its bottom enters the window, for every block including the shown-on-open top image
- [x] 2.2 Update the snap unit tests that asserted the old line-by-line caption reveal to assert the flush reveal
- [x] 2.3 Add the upward snap assertion to the shown-on-open regression test
- [x] 2.4 Add the upward top-snap scenario to the `reader-ui` spec delta and align the main spec scenario wording

## 3. Verification

- [x] 3.1 Reproduce the ProPublica Helene article's block geometry from the local DB and confirm both down and up scrolls snap correctly at 124x37
- [x] 3.2 Run the full test suite, `go vet`, and `golangci-lint` with no new issues
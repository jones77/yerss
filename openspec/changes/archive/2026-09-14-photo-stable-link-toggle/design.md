# Design: photo-stable-link-toggle

## Context

See proposal.md — Why. Two defects surface when the user toggles compact/
expanded link display with `x` while a photo is on screen:

1. **Photo changes size.** The image-block height fit (`fitBlock`, image.go:331)
   caps an image against `base := vpH - headerLines - 3` where `headerLines` is
   the line count of the *current-mode* rendered header. Expanded mode prepends
   the URL line (plus a blank), so the cap differs between modes; a block that
   has to re-fit on recompose renders a different number of photo rows per mode.
   Empirically (Helene article 93, halfblock, contentW 118): 32 rows compact vs
   30 rows expanded, and the width-guard in `composeImageBlock` (image.go:303)
   forces the re-fit whenever the cached block's width ≠ current content width.
2. **Photo drifts off screen.** `toggleLinkExpand` (image.go:596) anchors on the
   first visible line's *visible text*. A photo row strips to blank/escape
   noise, so `anchor == ""` and the toggle falls back to `recomposeArticle`'s
   total line-count delta — which is only correct for top-of-article
   insertions. Anchoring a caption line can also push the photo (above the
   caption) off the top of the viewport.

Header height is a red herring for the reader: photos should not care what the
text is doing.

## Goals / Non-Goals

- Goal: photo blocks render at the same size regardless of link display mode.
- Goal: toggling link mode keeps an on-screen photo on screen (photo rows and
  attribution rows treated as an anchored unit).
- Goal: keep clicking, OSC8 clickability, caption rendering, and the snap
  invariants untouched.
- Non-Goal: changing the image-load recompose path or the `atBottom` /
  `GotoBottom` semantics of `recomposeArticle`.
- Non-Goal: making every block width-matched; the width-guard serve path stays
  as-is (its output is already mode-independent).

## Decisions

### D1. Mode-independent header fit budget

Make the height fit consume a fixed per-article header budget, not the
mode-dependent rendered header. Add `fitHeaderLines` to `articleDerivation`
(and mirror it into `articleState`): the maximum line count of the header in
either link mode (compact and expanded), computed once in `computeDerivation`
by rendering the header for both modes. `fitBlock` and `nativeRenderCmd` switch
from `headerLines`/`m.article.headerLines` to this budget; the unchanged
`headerLines` field keeps serving everything that is genuinely about the actual
rendered header (line content, `leadShown`, header-relative anchors).

Rationale: reserving the tallest possible header guarantees the photo never
re-fits differently on a toggle. The compact-mode read loses nothing visible —
the reserved rows were going to be claimed by the expanded header anyway.
Because `renderHeader` is cheap (pure text), rendering it twice in the
derivation is negligible next to the markdown conversion the derivation already
caches.

Alternatives rejected: (a) fit against a literal fixed height — loses the
header-awareness that keeps header+photo+one body line inside the viewport;
(b) re-fit on every toggle using the other mode's header — exactly the resize
we are removing.

### D2. Photo-block anchor for the toggle

In `toggleLinkExpand`, if the first visible line (`lines[oldOffset]`) lies
inside a composed image block's span — `[ImgStart, ImgEnd]` in
`m.article.imageBlocks`, covering both photo rows and attribution/caption rows
— capture `(url, rowRel)` where `rowRel = oldOffset - ImgStart`, flip the mode,
recompose, then locate the same URL's block in the recomposed article and
`SetYOffset(ImgStart' + rowRel)` (clamped to `len(lines') - vpH`). If the URL is
not found after recompose, fall back to today's text-anchor behavior (no worse
than current). `atBottom` short-circuits first, so bottom-pinning is untouched.

Rationale: the photo block is the natural unit of continuity; a photo row has no
text for the existing anchor, and anchoring a caption text line mispositions the
photo above it. Anchoring by the block's own `ImgStart` also absorbs the
header's own mode-dependent height change, because block positions already
account for the actual rendered header.

### D3. Keep the width-matched serve path

A block already composed at the current content width (the `w == contentW`
guard in `composeImageBlock`) is served as-is; that path never re-fits, so its
photo size is already mode-independent. Only the re-fit path needed D1, and D1
keeps both paths agreeing after a toggle.

## Risks / Trade-offs

- [A photo that barely fits under the compact header is now capped by the
  expanded budget, so it may render a row or two shorter in compact mode than
  today] → Mitigation: accept the trade; it is the whole point of D1, and the
  reserved rows are exactly what the expanded header would consume next toggle.
- [`fitHeaderLines` duplicates `renderHeader` work in the derivation] →
  Mitigation: one extra string render per derivation (cheap) vs. the markdown
  conversion already cached.
- [Photo block not found after recompose (e.g. image dropped by dedup) →
  fallback re-find] → Mitigation: text-anchor and total-delta fallbacks remain,
  matching today's outcome.
- [Block span test (`ImgStart..ImgEnd`) mislabels a body line coincidentally
  inside the span] → Mitigation: spans are exact block rows produced by
  `composeArticle`; a body line can only land there if the block is composed.

## Migration Plan

Single-step code change; no data migration. Existing stored blocks, photo
bytes, and caches are untouched. Rollback is reverting the composition-budget
and toggle-anchor edits.

## Open Questions

None.
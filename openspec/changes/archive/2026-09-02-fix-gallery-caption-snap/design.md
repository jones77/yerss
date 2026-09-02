## Context

Reproduced on the ProPublica Helene article (DB id 93) at 100×40 and 80×24
with the real native height fit. The three-photo block is a WordPress
`wp-block-gallery`: an outer `<figure class="wp-block-gallery">` wrapping three
nested `<figure>`s, each with one `<img>` and its own `<figcaption>`, plus an
outer `<figcaption>` credit ("Courtesy of Joe Rogers").

Two independent defects surface there:

- **Caption attribution / duplication.** `convert.ConvertImages` emits each
  `figcaption` as a body paragraph, so every gallery figcaption renders as a
  left-aligned text block under the photos. Separately, `figureCaption`
  (`internal/convert/credit.go`) walks `findAll(figure)` in document order; the
  outer gallery figure matches first for each image and, because a single
  figcaption belongs to a figure's last `<img>`, returns `("", true)` for the
  first two images — so their `ResolveImageAttribution` returns "" and they
  compose with **no** centered caption. The third (last) image resolves its alt
  text as the caption. Net: photos 1–2 show only the left-aligned figcaption
  paragraph; photo 3 shows a centered alt caption plus the figcaption
  paragraphs.

- **Downward snap asymmetry.** `SnapYOffset` (`internal/ui/compose/image.go`)
  only bottom-snaps an inline image that is *partially* visible (top inside the
  window, last line below the fold) or whose offset lands in its photo range.
  A short image (block shorter than the viewport) entered from above is fully
  visible the moment its top crosses the fold, so no snap fires: pressing j
  from the paragraph above scrolls through it line-by-line (222→…→226 then a
  skip to the caption at 254), while the upward path's top-snap/sink/off-reveal
  stages snap on nearly every k press.

## Goals / Non-Goals

**Goals:**

- Every inline image whose figure carries its own figcaption composes that
  figcaption as the centered attribution beneath the photo, and the article body
  does not repeat it as a separate paragraph.
- A short inline image entered from above snaps flush to the viewport top on a
  single downward scroll, mirroring the upward snap, so photos align to the
  screen instead of scrolling through line-by-line.
- Preserve native-image placement cleanup, re-transmission, the 
  partial-visibility blank-strip invariants, and read-state behavior.

**Non-Goals:**

- Changing image transmission, caching, or the derivation/recomposition pipeline.
- Altering paging, half-page, or goto-bottom scroll behavior (multi-line moves
  still land mid-photo).
- Re-architecting the figcaption/credit extraction; the change is confined to
  attribution matching and converter output.

## Decisions

### D1: Attribute each image its own figure's caption

`figureCaption` SHALL match the figure that **directly** contains the image
(innermost figure), so a gallery photo in its own nested figure resolves that
figure's figcaption instead of the outer gallery's. The photo-grid rule is
preserved at the innermost level: a single figcaption over multiple `<img>`s in
one figure still belongs to that figure's last image (so a shared
gallery/photo-grid caption keeps its current behavior). `ResolveImageAttribution`
precedence changes from alt-first to the directly-containing figure's caption
first, then alt, then link text, then `photo: <source>`.

**Alternatives considered:**
- *Keep the outer gallery's "last image" attribution.* Leaves photos 1–2
  captionless — the reported bug.
- *Alt-first precedence.* Leaves the editorial figcaptions ("Sandra Rogers
  devoted…", "A peach tree blossoms…") out of the composed captions even after
  D1, since those images also carry alt text; the reader would still not get the
  real captions.

### D2: Do not duplicate a single-image figure's caption as body text

`ConvertImages` SHALL suppress the `figcaption` of a figure that directly
contains a single `<img>` (emitting only the image sentinel), because D1 composes
that caption centered beneath the photo. A `figcaption` over multiple images
(the gallery's shared credit, a photo-grid label) is not a single image's
caption, so it stays as body text. The attribution path reads the HTML directly
(`ImageCaption`), so suppressing the markdown copy loses no caption source.

**Alternatives considered:**
- *Suppress all figcaptions from body text.* Loses standalone/shared credits
  (e.g. the gallery's "Courtesy of Joe Rogers") that belong to no single image.
- *Leave figcaption body text and only fix D1.* Duplicates every caption under
  its photo (centered caption plus identical left-aligned paragraph).

### D3: Downward snap for short inline images entered from above

Extend the downward branch of `SnapYOffset`: when a single-line down move brings
an inline image's top into the window from below and the block is shorter than
the viewport (it cannot be partially clipped), snap to the image's top boundary
(its first line at the viewport top), mirroring the upward top-snap. The image
and its caption are one snap unit: the next downward move skips the whole block
onto the line immediately after its wrapped caption (never resting on a caption
line), and an upward move that would show only the caption's last line snaps the
whole block into view instead. The existing partial-visibility bottom snap, the
photo-range skip, and the `before`-gating (no snap may return the pre-move
offset) stay intact.

**Alternatives considered:**
- *Extend the partial-visibility bottom snap to short images.* Their last line
  is already above the fold when the top enters, so there is no distinct bottom
  stage to snap to; the top-boundary snap is the only unambiguous alignment.
- *Aggregate the gallery photos into one snap unit.* Fragile: photos are
  separated by body text and their indices shift; the top-boundary snap works
  with the existing per-block model.

## Risks / Trade-offs

- [Caption precedence flip changes existing articles] → A captioned image with
  both alt and figcaption now shows the figcaption (editorial) instead of the
  alt. This is the intended correction; the alt remains the fallback for
  images with no figure caption. Byte-identical renderings for uncaptioned
  images are unaffected.
- [Suppressing single-image figcaption body text could hide a caption the
  reader used] → The same text is composed centered beneath the photo by D1, so
  nothing is lost; a regression test asserts the caption appears exactly once,
  centered.
- [Down snap changes scrolling feel for short images] → It only fires when a
  short image's top enters from below; normal line scrolling above and below
  the image is unchanged, and every snap is `before`-gated so it cannot loop.
- [Innermost-figure matching is ambiguous for deeply nested markup] → The
  directly-containing figure of the image is well-defined in practice (the
  gallery nests one level); the matching prefers the nearest ancestor figure
  that carries a figcaption.

## Migration Plan

- Implementation order: D1+D2 (caption attribution + converter suppression)
  together, then D3 (down snap) with its scroll tests. Each lands with tests
  green; no schema or data migration.
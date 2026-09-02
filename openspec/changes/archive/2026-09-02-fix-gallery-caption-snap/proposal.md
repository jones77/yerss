## Why

A photo-dense article (the ProPublica Hurricane Helene piece) exposes two
rendering bugs in the reader. A WordPress photo gallery — three nested figures,
each with its own figcaption — renders those figcaptions as left-aligned body
paragraphs instead of centered captions, and its first two photos get no
centered caption at all. Separately, downward single-line scrolls do not snap
short inline images to their boundaries the way upward scrolls do: pressing j
from just above a photo scrolls through it line-by-line, while pressing k snaps
cleanly.

## What Changes

- **Gallery captions become real captions.** Each gallery photo's own figure
  caption renders as the centered attribution directly beneath its photo,
  instead of the caption text also appearing as a left-aligned body paragraph.
  The attribution resolution no longer misattributes a shared gallery figure's
  credit to the last image only, so every photo in a multi-photo figure with
  per-photo figcaptions gets its own centered caption. No caption text is
  duplicated as body content.
- **Downward scroll snaps mirror upward scroll.** A single-line downward scroll
  that brings an inline image into view from above snaps the image to its snap
  boundaries (flush to the viewport top, or its wrapped caption's last line at
  the viewport bottom), matching the upward scroll's snapping so photos align to
  the screen instead of scrolling through line-by-line. The partial-visibility
  blank-strip invariants are unchanged.
- No change to native-image placement cleanup, re-transmission, or read-state
  behavior.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `inline-image-rendering`: the inline-image attribution requirement changes —
  an inline image in a figure that carries its own figcaption renders that
  caption centered beneath it, and the caption text is not also emitted as body
  text.
- `reader-ui`: the atomic image scroll behavior requirement changes — downward
  single-line scrolls snap an inline image entered from above to its snap
  boundaries, mirroring the upward scroll transitions.

## Impact

- `internal/convert/convert.go` (and `credit.go`) — figcaption handling so the
  caption text is not duplicated as body paragraphs.
- `internal/ui/compose/image.go` — `ResolveImageAttribution` / `figureCaption`
  attribution for multi-photo figures with per-photo figcaptions.
- `internal/ui/compose/snap.go` (`SnapYOffset`) — downward snap transitions for
  inline images entered from above.
- Tests in `internal/ui/*` (attribution, scroll/snap) and `internal/convert/*`
  (figcaption conversion).
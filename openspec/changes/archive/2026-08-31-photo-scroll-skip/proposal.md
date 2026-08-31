## Why

When the full lead photo (and its caption) is already on screen, scrolling away
from it is awkward in both directions. Scrolling down past a fully visible photo
requires several keys because the height-fitted block is always fully visible;
scrolling up through the caption is impossible because the upward snap range
covers the caption too and jumps straight to the photo. On a native terminal
each of those intermediate scrolls re-transmits the whole photo, which is slow.
The reader should skip a fully visible photo in one keypress, scroll freely
through the caption, and jump to the top when leaving a fully visible photo.

## What Changes

- Extend the existing downward snap: when a downward scroll (line, page,
  half-page, or goto-bottom) either lands within the photo rows or leaves the
  full photo block on screen (image top at or above the window top and image
  bottom at or within the viewport bottom), the offset SHALL snap to the first
  non-photo line — the attribution line, or the line past the block when there
  is no attribution.
- Narrow the upward snap from the whole block to the photo rows only, so an
  upward scroll landing within the caption advances the caption line by line
  instead of snapping to the photo.
- On a full-image-capable terminal, an upward scroll while the full photo is on
  screen snaps the offset to the article top (0) in one step, avoiding the
  repeated native re-transmits of nudging through the header.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `reader-ui`: "Atomic image scroll behavior" — a downward scroll skips a fully
  visible photo; the upward snap is narrowed to the photo rows so the caption
  scrolls line-by-line upward; and a native-terminal upward scroll from a fully
  visible photo jumps to the article top.

## Impact

- `internal/ui/image.go`: `snapYOffset` gains a `native bool` parameter; the
  down branch (already implemented) is unchanged; the up branch's reveal range
  narrows from `[imgStart, imgEnd]` to `[imgStart, capStart)`; a new native-only
  up rule returns 0 when the photo is fully visible and `yOffset > 0`.
  `scrollArticle` passes `m.article.nativeImg`.
- Tests: update `TestSnapYOffset` (the "up from caption" cases now expect
  unchanged offsets; add native skip-to-top cases) and extend the UI-level
  `TestScrollSkipsFullyVisiblePhoto` with an upward caption-scroll and a
  native skip-to-top check.
- No new external dependencies.

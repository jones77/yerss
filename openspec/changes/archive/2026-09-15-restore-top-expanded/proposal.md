## Why

In expanded link mode, quitting at the very top of an article and restarting
jumps the viewport to the article's photo, scrolling the URL header and title
off screen. The saved content anchor for a top-of-article position is the title
row minus the gap for the expanded header's blank separator, so the restore
lands at offset 1 rather than 0. When the photo load then lands, the
recompose's snap-forward rule treats the photo's top materializing inside the
window as an inline image the reader scrolled past, and snaps the viewport to
it — even though the reader was sitting in the header. The same weakness makes
`x`-toggling near the top jump to a photo whose top materializes mid-window.

## What Changes

- A saved selection whose article offset is zero restores to offset 0, keeping
  the header (including the expanded URL line) visible through later image
  loads, instead of landing one line lower via the anchor.
- The recompose snap-forward rule is skipped while the reader's first visible
  line is inside the header region: a photo below the header is the normal open
  state, not a scrolled-past image, so an image load must not scroll the header
  off.
- `FindAnchorLine` gains a fallback for an anchor text split across two
  consecutive lines by a reflow (its first word pulled to the end of the
  previous line), so the `x`-toggle re-anchors on the right line instead of
  falling through to the snap-forward offset.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `article-recomposition`: the reading-position-preservation rules gain a
  header-region exception to the snap-forward behavior and a top-of-article
  restore rule that keeps the header visible through image loads.

## Impact

- **Code:** `internal/ui/articleview/method.go` (`RestoreAt`,
  `RecomposeInPlace`), `internal/ui/compose/anchor.go` (`FindAnchorLine`), and
  tests in `internal/ui/selection_test.go` and
  `internal/ui/compose/anchor_test.go`.
- **Behavior:** restarting at the top of an expanded-mode article stays at the
  top once the photo loads; an image load landing below the header never scrolls
  the header off; toggling link mode near the top no longer yanks the viewport
  to a photo.
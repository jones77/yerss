## Context

The reader persists a content anchor on quit and re-finds it on restart
(`RestoreAt` in `articleview/method.go`; `AnchorAt`/`FindAnchorLine` in
`compose/anchor.go`). For a top-of-article position the anchor is the title's
first wrapped line with a gap of 1 — the expanded header's blank separator
between the URL line and the title. The restore therefore lands at offset 1, not
0. A photo load landing just afterwards runs `RecomposeInPlace`, whose
snap-forward rule (`compose.SnapAfterRecompose`) snaps the viewport to any image
block whose top materialized inside the window below the offset. The photo's top
is inside the window (it sits below the header), so the viewport jumps to the
photo and the URL header and title scroll off — the reported restart bug.

`SnapAfterRecompose`'s lead exemption (`(i>0 || !lead)`) only covers a genuine
lead block directly under the header. When the lead URL also opens the body, the
block is a suppressed lead that materializes after some intro text
(`leadShown == false`), so it is snapped to like any inline image. The same
weakness makes `x`-toggling near the top jump to the photo, and a reflowed
anchor line whose first word moved to the previous line misses `FindAnchorLine`
entirely, falling through to the snapped offset.

## Goals / Non-Goals

**Goals:**
- A restart at the very top of an expanded-mode article stays at the top once
  the photo load lands, with the URL header and title visible.
- An image load landing below the header never scrolls the header off while the
  reader is in the header region.
- A link-mode toggle near the top re-anchors on the right line when the anchor
  text reflowed across a line break.

**Non-Goals:**
- Changing the snap-forward behavior for a reader genuinely scrolled into the
  body (past the header): a block materializing mid-window still snaps to its
  top, as specified and tested.
- Re-anchoring a line whose leading text is a wrapped URL fragment that vanishes
  on collapse (the `off=12` real-DB edge); its anchor text does not survive the
  mode change and remains a known limitation.
- Changing the placeholder/derivation or block caches.

## Decisions

### 1. A zero article offset restores to offset 0

`RestoreAt` returns offset 0 before consulting the block or text anchor when the
saved `ArticleOffset` is 0. The top is unambiguous across reflows and modes, so
the anchor adds nothing there; landing at 0 also matches compact mode and the
`x`-toggle's existing `oldOffset == 0` protection, and it means the later image
load sees `oldOffset == 0` and takes `RecomposeInPlace`'s no-snap path.

Alternative considered: special-casing the expanded-mode gap (subtract 1). That
is brittle and only fixes the exact reported geometry.

### 2. Skip the snap-forward while the reader is in the header region

`RecomposeInPlace` applies `SnapAfterRecompose` only when `oldOffset >=
s.headerLines`; inside the header it bumps the offset by the inserted lines
without snapping. A photo below the header is the article's normal open state,
not a scrolled-past image. The guard lives in `RecomposeInPlace`, not in
`SnapAfterRecompose`, so the snap's own unit tests and the past-the-header
behavior are untouched.

### 3. `FindAnchorLine` falls back to a split-across-lines pair

After the exact, prefix, and longest-common-run passes fail, `FindAnchorLine`
scans consecutive line pairs and returns the earlier line when the pair's joined
text contains the anchor text. This reconstructs an anchor whose first word a
reflow pulled to the end of the previous line, so the toggle re-anchors on the
line holding that first word instead of leaving the snapped offset.

## Risks

- The split-pair pass could match a coincidental reconstruction. The anchor
  text is a full rendered line (tens of characters), so a substring match across
  two lines is a strong signal; a unit test covers the negative case.
- `RecomposeInPlace`'s header guard uses the post-recompose header height. For a
  mode toggle the header height differs slightly between modes, but the region
  is a few rows in either mode, so the guard is correct at the granularity it
  needs.
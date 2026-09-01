## Context

See proposal.md — Why. The current `snapYOffset` (internal/ui/image.go) already
implements the two boundaries implicitly, but as ~8 overlapping conditionals
with three separate `best`-selections, a `before`-offset gate, and dedicated
branches for the fills-viewport collapse, the lead block, and adjacent-block
collisions. The conditions key on `capStart` (the caption's first line) where
the intended boundary is `imgEnd` (the wrapped caption's last line). Per-block
geometry lives in `imageBlock{imgStart, capStart, imgEnd}` (article.go) and the
snap contract is specified in reader-ui "Atomic image scroll behavior".

The user's model (confirmed in exploration): every block has two snap
boundaries — the image's first line at the viewport top, and the last line of
the wrapped caption at the viewport bottom — and scrolling snaps between them.

## Goals / Non-Goals

**Goals:**
- Make the two boundaries first-class: `TOP(b) = b.imgStart`,
  `BOTTOM(b) = b.imgEnd - vpH + 1`.
- Key the bottom-snap and scroll-off conditions on the block's last line
  (`imgEnd`) instead of `capStart`, so a wrapped caption whose tail hangs below
  the fold snaps correctly.
- Restructure the snap rules so the four canonical positions per block
  (enter-at-bottom, top, exit, scroll-off) drive the transitions, with the
  special cases as named, localized exceptions.
- Preserve every behavior already locked in by the existing snap tests.

**Non-Goals:**
- No change to the lead block's skip behavior, the native top-skip
  optimization, or the recompose rules (`snapAfterRecompose`).
- No change to `nativeImageClear` / `suppressClippedNativeTransmits` /
  `previewClippedImage`: the containment check stays keyed on the photo range
  (`capStart`), because only the photo rows are a native placement — the
  caption is text and never paints over the border.
- No change to page/half-page/goto-bottom behavior (they opt out of snapping).

## Decisions

**1. Four canonical positions per block, computed once per snap call.**

For each block and the current `vpH`:

```
BOTTOM(b) = b.imgEnd - vpH + 1   // caption's last line at viewport's last row
TOP(b)    = b.imgStart           // image's first line at viewport's first row
EXIT(b)   = b.capStart           // photo scrolled out, caption at viewport top (down-exit)
OFF(b)    = b.imgStart - vpH     // image's top at the fold, block fully below (up-exit)
```

The down transitions are `BOTTOM → TOP → EXIT`; up mirrors `TOP → BOTTOM → OFF`.
When `TOP == BOTTOM` (block fills the viewport) the middle stage collapses and
the sequence degrades to `BOTTOM → EXIT` (down) and `TOP → OFF` (up) — which the
existing loop-guard plus collapse branch already express; the boundary
definition makes the collapse a derivation, not a patch.

*Alternative considered:* explicit per-block stage state (a struct field or map
keyed by block). Rejected after tracing: every snap path lands exactly on a
boundary offset, so the `before == boundary` inference fires correctly in all
the paths we could construct (top-snap entry, reveal-previous fallback,
recompose snap). Explicit state would add bookkeeping for no behavioral gain and
would need resetting on recompose.

**2. Bottom-snap and scroll-off conditions key on `imgEnd`, not `capStart`.**

Current conditions (`b.capStart > yOffset+vpH`) test the caption's *first* line
against the fold, leaving a band of offsets where a wrapped caption's tail hangs
below the fold unsnapped. Replace with `b.imgEnd > yOffset+vpH` — the block's
last line. For one-line captions (capStart == imgEnd) this is identical; for
wrapped captions it widens the snap zone to exactly the "last line below the
fold" state the user described.

*Alternative considered:* keep `capStart` and add a separate caption-tail check.
Rejected — two conditions for the same boundary invites the same confusion the
change removes.

**3. Restructure the rule order around the positions.**

Down:
1. Landed in `[imgStart, capStart)` (photo range): lead → `EXIT`; inline with
   `before < imgStart` → `TOP` (reveal); else → `EXIT`.
2. Lead fully visible → `EXIT` (pre-consumed stages).
3. `before == BOTTOM(b)` and still ≤ `TOP(b)` → `TOP` (rise stage).
4. Bottom-clipped (`imgStart` in window, `imgEnd` below fold) → `BOTTOM` of the
   lowest such block (`best` = largest `imgStart`, the current selection).

Up mirrors: photo-range → `TOP` (reveal); `before == TOP(b)` → `BOTTOM` (sink
stage); entry from above (`imgEnd` crossed into the window) → `TOP`; scroll-off
(bottom-clipped) → `OFF` with the reveal-previous fallback when `OFF` lands in a
preceding block's photo range.

The `before`-gate on the rise/sink stages stays — it is the two-stage sequence
encoded as a transition, and the guard "never return the pre-move offset"
(`snap` helper) remains the loop protection.

**4. Named exceptions stay, expressed against the boundaries.**

- Lead block (`i == 0`): stages pre-consumed — photo-range landings and the
  fully-visible skip go straight to `EXIT`; native top-skip unchanged.
- Adjacent-block collisions: keep the `best` (lowest) selection for multiple
  bottom-clipped candidates and the reveal-previous fallback in the scroll-off
  branch — these are the tie-breaks when blocks are spaced closer than the
  viewport height.

## Risks / Trade-offs

- **Widened snap zone for wrapped captions** (Decision 2) is a genuine behavior
  change: a block whose caption tail was left cut after certain moves now snaps
  down to the bottom boundary. → Covered by the new "Wrapped caption with its
  last line below the fold" scenario; verify against the existing
  scroll-through tests to confirm no mid-band free-scroll regression.
- **Refactor regression in the degenerate cases** (fills-viewport, adjacent
  tall images). → The existing loop-protection tests
  (`TestSnapYOffsetFullViewportBlockDoesNotLoop`,
  `TestScrollUpThroughFullViewportImageDoesNotLoop`,
  `TestScrollDownNeverLeavesInlineImageBlank`) are the safety net; they must
  pass unmodified.
- **The `before`-keyed stages are positional proxies for state.** → Analysis
  found every snap path lands exactly on a boundary, so the proxy holds; if a
  future path breaks it, the fix is localized to the rise/sink conditions.

## Migration Plan

Pure refactor + condition change within `internal/ui/image.go` and the snap
tests. No schema, storage, or CLI changes; no rollout steps. Rollback is the
previous commit. The reader-ui spec is updated in the same change and archived
with it.

## Open Questions

None — the exit target (`capStart`), the bottom boundary (`imgEnd`), and the
stage detection (keep `before`-keying) were resolved during exploration.
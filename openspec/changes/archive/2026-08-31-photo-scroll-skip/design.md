## Open Questions

Resolved. The up direction is now also refined (previously the change was
down-only): the up-snap is narrowed to the photo rows, and a native-only
skip-to-top rule is added. Halfblock terminals keep line-by-line upward
scrolling because re-rendering halfblock art is cheap.

## Options

- Down from a fully visible photo snaps to the caption. Chosen.
- Narrow the up-snap to the photo rows (caption scrolls upward line by line).
  Chosen, matching the existing down-direction caption behavior.
- Native-only skip-to-top on up. Chosen; halfblock is fast so it scrolls the
  header line by line.
- Skip-to-top always (even halfblock). Rejected: pointless on halfblock and
  would hide the header without the performance motivation.

## Problems

- The skip-to-top must not fire when there is nothing above to reveal
  (`yOffset == 0`), and must not fire on halfblock. Both are guarded in the
  rule below.
- The reveal snap and the skip-to-top both apply to an upward move; ordering
  matters. The skip-to-top is evaluated first (it only matches `yOffset <=
  imgStart`), then the reveal (`yOffset >= imgStart`), so at `yOffset ==
  imgStart` the skip-to-top wins and the offset jumps to 0.

## Remaining Decisions

None.

## Details

### Downward skip of a fully visible photo (already implemented)

The photo is height-fitted so its block always contains at most `vpH - header`
rows; the photo is fully on screen exactly when the window top is at or above
the image start and the window bottom reaches or passes the image end:

    yOffset <= imgStart && yOffset+vpH >= imgEnd

When that holds after a downward move — or the move landed within the photo
rows (`imgStart <= yOffset < capStart`) — the offset snaps to `capStart`, the
first non-photo line.

### Upward refinements (new)

`snapYOffset` gains a `native bool` parameter. Its full shape:

```go
func snapYOffset(yOffset, vpH, imgStart, imgEnd, capStart, direction int, native bool) int {
    if imgStart < 0 || imgEnd < imgStart {
        return yOffset
    }
    if direction > 0 {
        if yOffset >= imgStart && yOffset < capStart {
            return capStart   // down into photo rows -> caption
        }
        if yOffset <= imgStart && yOffset+vpH >= imgEnd {
            return capStart   // down while photo fully visible -> caption
        }
    }
    if direction < 0 {
        if native && yOffset > 0 && yOffset <= imgStart && yOffset+vpH >= imgEnd {
            return 0          // up while native photo fully visible -> article top
        }
        if yOffset >= imgStart && yOffset < capStart {
            return imgStart   // up into photo rows -> reveal
        }
    }
    return yOffset
}
```

Key changes from the previous shape:

1. The up reveal range narrows from `[imgStart, imgEnd]` to `[imgStart,
   capStart)` so a move landing on a caption line is returned unchanged and the
   caption scrolls upward line by line.
2. A native-only rule returns 0 for an upward move while the photo is fully
   visible and `yOffset > 0`, jumping to the article top instead of nudging
   through the header (which re-transmits the native photo each step).

`scrollArticle` passes the flag:

```go
snapped := snapYOffset(after, m.article.viewport.Height,
    m.article.imgStart, m.article.imgEnd, m.article.capStart, dir,
    m.article.nativeImg)
```

### Test changes

`TestSnapYOffset` adds a `native` column to its case struct. Existing
non-native up cases change:

- "up from caption first line" `{8,...,-1}` now wants `8` (was `4`).
- "up from caption last line" `{9,...,-1}` now wants `9` (was `4`).
- "up from photo interior" `{6,...,-1}` and "up onto first line" `{4,...,-1}`
  still want `4` (photo rows).

New cases: native up from a fully visible photo (`{2,20,4,9,8,-1,native}` →
`0`), and a halfblock guard (`{2,20,4,9,8,-1,not-native}` → `2`, unchanged).

The UI-level `TestScrollSkipsFullyVisiblePhoto` adds: after the existing down
skip, set the offset below the block and scroll up to confirm it lands on the
caption line (`imgEnd`) rather than the photo, then continue up through the
caption, then reveal; and set `m.article.nativeImg = true`, place the offset at
the photo top, scroll up, and confirm the offset becomes 0.

### Edge: page-down already past the block

A page-down usually lands with the window top `>= imgEnd` (beyond the block), so
neither down condition triggers and the full-page scroll is preserved. The only
new snaps fire when the photo is still at least partly on screen after the move.

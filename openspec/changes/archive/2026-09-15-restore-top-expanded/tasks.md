## 1. Restore-at-top lands at the very top in expanded mode

- [x] 1.1 `RestoreAt` returns a top-of-article restore to offset 0 before the
      block/anchor logic (the anchor would land a line lower, and a photo load
      landing just afterwards would snap forward past the header)
- [x] 1.2 Regression test `TestRestoreSelectionTopExpandedStaysAtTop` mirroring
      the reported restart (expanded, suppressed-lead photo, resize before the
      photo load lands); confirmed it failed at offset 12 (the photo) before the fix

## 2. Header-region guard on the recompose snap

- [x] 2.1 `RecomposeInPlace` skips `SnapAfterRecompose` when the old offset is
      inside the header/top region, so a photo below the header is treated as
      the normal open state and the header is never scrolled off by a load

## 3. Reflow split-anchor fallback

- [x] 3.1 `FindAnchorLine` adds a split-pair pass for an anchor text
      reconstructed across two consecutive lines (a reflow pulls its first word
      to the end of the previous line)
- [x] 3.2 Unit test `TestFindAnchorLineSplitAcrossReflow`
- [x] 3.3 Verified the real-DB toggle sweep: offsets 7/8 no longer yank to the
      photo (off=12 on a URL-fragment-leading line remains a pre-existing edge
      whose anchor text vanishes on collapse)

## 4. Verification and cleanup

- [x] 4.1 Removed the temporary repro/debug test files and worktrees
- [x] 4.2 Ran build, vet, and the full test suite; all green
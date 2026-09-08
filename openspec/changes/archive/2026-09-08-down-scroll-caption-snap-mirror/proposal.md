## Why

Scrolling up through the ProPublica Helene article's three consecutive Sandra
Rogers photos shows each photo's caption at the viewport bottom (the two-stage
snap: top, then caption at the bottom, then off), but scrolling down through
them snaps each photo straight to its top and the next press skips the whole
block off — the caption-at-bottom stage never appears. The down direction also
wasted a press on the blank separator line below each block: after a photo
scrolled off, the next photo rested one line below the viewport top needing an
extra scroll to snap flush. The two directions should be mirrors of each other,
and the space below an image should scroll off with the image as one unit.

## What Changes

- The lead photo now rises flush to the viewport top on the first downward
  press (its entry stages are pre-consumed — it was shown on open) and the
  following press skips the whole photo-and-caption block past the blank line
  below it onto the first line of the next image or paragraph; it is never
  skipped past unseen in a single press.
- A downward scroll that skips an image off the top of the viewport now lands
  on the first line of the next image or paragraph (past the block's blank
  separator), so the space below an image scrolls off with it instead of
  becoming a resting position.
- For two adjacent images (one blank line apart), the down-skip lands on the
  next image's BOTTOM boundary when that image is short enough to have a
  distinct bottom stage — its caption at the viewport bottom, mirroring the
  upward sink so the caption snapping works on the way down exactly as it does
  on the way up — and the next press rises it to its top. The rise from a
  bottom boundary takes precedence over the photo-range skip so the adjacent
  bottom stage never loops.
- A non-snapping move (goto-bottom, page, recompose) that leaves a short inline
  block fully visible mid-window is settled to its bottom boundary by the next
  upward scroll, so scrolling up from the article's end never walks the last
  photo past the fold line by line.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `reader-ui`: the atomic image scroll behavior requirement changes how a
  downward scroll skips an image off the viewport (landing on the first line of
  the next image or paragraph, past the blank separator), how the lead photo
  enters (rise flush to the top before the whole block skips off), how adjacent
  photos present on the way down (the next photo's bottom boundary, mirroring
  the up direction), and how the article end settles the last short photo.

## Impact

- `internal/ui/compose/image.go`: `SnapYOffset` — the down-exit target is the
  first line of the next image or paragraph (one past the block's blank
  separator); an adjacent short image's skip lands on its BOTTOM boundary; the
  rise stage runs before the photo-range skip; the fully visible lead photo
  rises flush to the top instead of skipping; and the up direction settles a
  short block fully visible mid-window to its bottom boundary.
- `internal/ui/article_image_test.go`: the snap tests assert the new exit
  (first line past the block), the lead rise-then-skip flow, the adjacent
  photo's bottom-boundary reveal on the way down, and the up-from-article-end
  settle.
- `openspec/specs/reader-ui/spec.md`: the "Atomic image scroll behavior"
  requirement's down-exit, lead-entry, adjacent-photo, and article-end
  sentences and the related scenarios are revised.
## Why

After backing out of an article, the article's last-visible photo sometimes
remains drawn over the list view until the app quits. The exit frame emits a
kitty delete-by-id sequence using the lowercase `d=i` action, which per the
kitty graphics protocol only deletes the placement and deliberately keeps the
terminal's cached image data — so the placement survives on terminals whose
delete handling is partial (Ghostty, where the kitty-graphics implementation is
incomplete).

## What Changes

- Change the kitty delete-by-id sequence to the uppercase `d=I` action, which
  deletes the placement AND frees the terminal's cached image data, matching
  the documented intent in the cleanup spec ("freeing its cached image data").
- On Ghostty specifically — where delete-by-id is unreliable even with the
  correct action — emit a belt-and-suspenders delete-all (`d=a`) on the frame
  that leaves the article view, so a surviving placement is cleared rather than
  floating over the list.
- Keep the existing per-id tracking and per-scroll-out deletes unchanged; the
  change is confined to the exact escape sequences emitted and to which
  terminals get the fallback.
- No change to OSC 1337 (iTerm2/WezTerm) handling, which correctly emits no
  delete sequences.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `native-image-rendering`: the placement-cleanup requirement's delete-by-id
  sequence must free the cached image data (`d=I`), and leaving the article
  view on Ghostty must also emit a delete-all clear.

## Impact

- `internal/image/native.go` — `DeleteByID` sequence (lowercase → uppercase `d`).
- `internal/ui/image.go` — `computeNativeClear` exit path; add the Ghostty
  delete-all fallback when leaving the article view.
- `internal/image/protocol.go` — possibly expose a Ghostty-detection helper or
  carry the terminal identity on `NativeRenderer`.
- Tests in `internal/ui/article_image_test.go` asserting the exit-frame
  sequences (`TestNativeKittyExitDeletesEveryRecordedID` and friends).
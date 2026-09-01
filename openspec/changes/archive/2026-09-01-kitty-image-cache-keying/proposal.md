## Why

The kitty image id is derived from the image URL alone (`stableID(url)`), but
the transmitted payload is re-encoded at the terminal's current geometry, so a
resize sends a different-sized PNG under the same id. Whether a terminal
replaces or accumulates an image when the same id is re-transmitted with
different data is not guaranteed, and Ghostty has been observed growing
unboundedly during heavy image testing. The id scheme must make each distinct
render independently addressable and deletable so the terminal's cache can be
reclaimed deterministically.

## What Changes

- Derive the kitty image id from the URL and the render's pixel dimensions, so
  each distinct render size is a distinct terminal image.
- Record, per native image block, the id it was rendered under.
- When a render is replaced by a re-render at a new geometry, delete the prior
  id before transmitting the new one.
- Delete every recorded id on article exit (building on the exit-delete already
  proposed in `image-cache-eviction-and-native-fixes`).

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `native-image-rendering`: The kitty image id SHALL be unique per (image,
  render size) rather than per image, a re-render at a new size SHALL delete
  the prior size's image, and cleanup SHALL delete every id the article
  transmitted.

## Impact

- Code: `internal/image/native.go` (`stableID` → per-render id, `escape`,
  `DeletePlacement`), `internal/ui/image.go` (track the id per block, delete
  the prior id on re-render), `internal/ui/article.go` (`imageBlock` gains the
  id).
- Tests: id stability and delete-by-id escape emission across resizes.
- No new dependencies, no schema or config changes.
- Builds on `image-cache-eviction-and-native-fixes` (exit delete by id); that
  change introduces the exit delete, this one makes the id itself per-render.

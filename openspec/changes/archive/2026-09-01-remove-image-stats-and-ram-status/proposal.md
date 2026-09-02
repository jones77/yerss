## Why

The `--image-stats` diagnostic and the `RAM <n>MB` status-bar readout were added
to measure image footprint. That investigation is done, and the readouts are no
longer wanted: the flag is a one-off measurement tool with no ongoing value, and
the RAM figure reports an estimate (`width × height × 4`) that overstates the
true resident memory of decoded JPEGs, which the status bar was never a good
place to surface.

## What Changes

- Remove the `--image-stats` command-line flag and the code that backs it: the
  `runImageStats`/`printImageStats`/`percentile` command, the
  `Store.ImageStats()` query and its `ImageStats`/`ImageSize` result types, and
  their tests.
- Remove the `RAM <n>MB` element from the list-view status bar, including the
  now-unused decoded-byte accounting on the image cache (`TotalBytes`,
  `estBytes`, and the `totalBytes` field).
- The `DB <n>MB` database-size element and the `formatMB` helper remain.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `image-stats`: the `--image-stats` diagnostic is removed — this is the
  capability's only requirement, so the capability goes away.
- `reader-ui`: the list view status bar no longer reports the in-memory image
  footprint; the right side becomes `?: help · DB <n>MB · <percent>% ·
  <n>/<total>`.

## Impact

- `cmd/yerss`: remove the `image-stats` flag registration and early-exit path;
  delete `imagestats.go`.
- `internal/store`: remove `ImageStats`/`ImageSize` and `Store.ImageStats()`;
  drop the now-unused `bytes`, `image`, and image-format blank imports.
- `internal/image`: remove `Cache.TotalBytes()`, `estBytes()`, and the
  `totalBytes` field, and simplify `Cache.Set` to stop tracking byte totals.
- `internal/ui`: remove the RAM element from `renderStatusBar` and update the
  status-bar tests.
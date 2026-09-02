## Context

The `--image-stats` flag (`cmd/yerss/imagestats.go`) prints a read-only image
footprint report backed by `Store.ImageStats()` and `ImageStats`/`ImageSize`
types in `internal/store`. The list-view status bar (`renderStatusBar` in
`internal/ui/list.go`) reads `m.sess.ImgCache.TotalBytes()` for its `RAM <n>MB`
element, which is the only consumer of the decoded-byte accounting on
`image.Cache` (`totalBytes`, `estBytes`, `TotalBytes`). See proposal.md for
motivation.

## Goals / Non-Goals

**Goals:**
- Delete the `--image-stats` flag and all code that exists only to serve it.
- Delete the RAM element from the status bar and the cache accounting that
  existed only to feed it.

**Non-Goals:**
- Keep the `DB <n>MB` element, `Store.DBSize()`, and the `formatMB` helper.
- No change to image fetching, rendering, or storage behavior.

## Decisions

### D1 — Remove the image-stats path entirely, including the store query

Delete `cmd/yerss/imagestats.go`; drop the `imageStats` option field, flag
registration, and the `if imageStats` early exit in `main.go`. Remove
`ImageStats`/`ImageSize` and `Store.ImageStats()` from `internal/store`,
including the `bytes`, `image`, and image-format blank imports that existed
only for `DecodeConfig`. Remove `pngBytes`, `TestImageStatsEmpty`, and
`TestImageStatsAggregates` from `store_test.go`.

### D2 — Drop the RAM element and the cache byte accounting together

Remove the `RAM <n>MB` line from `renderStatusBar` and update the surrounding
comment. Because the RAM readout is the sole caller of `Cache.TotalBytes()`,
also remove `TotalBytes()`, `estBytes()`, and the `totalBytes` field, and
simplify `Cache.Set` to a plain upsert. This avoids leaving dead accounting
code whose only purpose was a readout that no longer exists.

### D3 — Keep DB reporting intact

`Store.DBSize()`, the `dbSize` model field, and the `DB <n>MB` element all
remain, since the user only asked to remove the RAM number and the
`--image-stats` flag.

## Risks / Trade-offs

- [Status bar layout] → The right side shrinks by one element; the existing
  truncation logic in `renderStatusBar` already handles narrow widths, and the
  tests pin the new `?: help · DB <n>MB · …` ordering.
- [Stale references] → Any test or comment mentioning `RAM` or `image-stats`
  must be updated in the same change; a grep for those tokens at the end is the
  verification.

## Migration Plan

None. No configuration, data, or schema changes; the flag and readout simply
disappear. Rollback is reverting the change.
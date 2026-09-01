## Context

`renderStatusBar` builds the right side from `formatSize(m.dbSize)`, the percent, and the position. The decoded-image byte total comes from the image cache counter added in `collapse-image-caches`.

## Goals / Non-Goals

**Goals:**
- Both RAM and DB visible, RAM first, whole megabytes, no space before `MB`.

**Non-Goals:**
- No process-RSS measurement; RAM is the decoded-image cache estimate.
- No change to the left side or the tag-filter display.

## Decisions

### D1 — `formatMB(bytes int64) string`

Returns `fmt.Sprintf("%dMB", bytes>>20)` (floored). Lives beside the existing `formatSize` in `internal/ui/size.go`. Rationale: distinct from `formatSize`, which the `--init-db` prompt still uses.

### D2 — RAM value sourced from the image cache

The model reads `m.imgCache.TotalBytes()` each render. The value is an estimate (decoded width*height*4), refreshed on cache writes.

### D3 — Order RAM before DB

Matches the decision that memory is the primary growth signal.

## Risks / Trade-offs

- [Estimate not allocation] → Acceptable; it tracks the unbounded term.
- [Depends on collapse-image-caches] → Execute that change first.

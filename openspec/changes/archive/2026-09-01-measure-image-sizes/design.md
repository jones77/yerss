## Context

`store.Store` persists photo bytes and rendered blocks in `article_images`. The decoded-size estimate needs each photo's pixel dimensions, obtainable from the stored bytes via `image.DecodeConfig` without a full decode.

## Goals / Non-Goals

**Goals:**
- One flag, one report, read-only, offline.

**Non-Goals:**
- No LRU or cache bounding; this only produces the numbers for that future decision.
- No JSON output mode, no histogram buckets beyond min/median/p90/max.

## Decisions

### D1 — `--image-stats` long-only flag

No short alias, to avoid colliding with existing short flags. The flag is registered alongside the others and handled in `main` after the store opens and before the startup gate.

### D2 — Store-level aggregation

Add `Store.ImageStats()` returning counts and a `[]ImageSize{URL string; PhotoBytes, BlockBytes int; Width, Height int; Decodable bool}`. SQL aggregates count/sum/min/max; the median and p90 are computed in Go after loading the lengths. Rationale: keeps SQL simple and percentiles testable.

### D3 — Decoded-size estimate is width*height*4

An RGBA estimate per photo. `image.DecodeConfig` gives width/height; non-decodable blobs are reported as unparseable and excluded.

## Risks / Trade-offs

- [DecodeConfig reads each blob] → Bounded by image count; acceptable for a diagnostic invoked on demand.
- [Percentile choice] → p90 is a heuristic; median and max are exact.

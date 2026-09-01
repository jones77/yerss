## Context

`internal/image/image.go` defines four caches that differ only in their value type. The specs describe bounded caches that do not exist. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One cache implementation, four instantiations, no eviction.
- A byte total on the decoded-image cache (surfaced by the status-bar change).
- Specs that match the code.

**Non-Goals:**
- No LRU, no byte budget, no decoded-resolution cap.
- No change to cache keying semantics beyond the native spec reconciliation.

## Decisions

### D1 — Generic map cache replaces the four bespoke types

One `cache[V]` type (`mu sync.Mutex; m map[string]V; Get/Set`) in `internal/image`, with `NewCache[V]()`. `Cache`/`Blocks`/`Photos`/`Natives` become distinct named instantiations so call sites keep their types. Rationale: removes ~100 lines of duplicated map+mutex wrappers while preserving the existing API.

### D2 — Decoded-image byte total

The decoded-image cache tracks `totalBytes`, incremented on `Set` (estimate `Dx*Dy*4`) and decremented when an entry is overwritten. `TotalBytes() int64` exposes it. No eviction. Rationale: the status-bar `RAM` readout needs a number; it is an estimate, not allocation accounting.

### D3 — Native render cache keyed by render size

Keep the existing key `url\x00width\x00maxHeight` (width + viewport-height cap). A viewport-height-only resize re-renders. Rationale: the height cap changes the fitted block, so the cache key must include it.

### D4 — Decode once per native render

`renderNativeFitted` decodes the photo bytes once, then scales/re-encodes the shared decoded image for each candidate height. Rationale: matches the native spec and avoids up to 3x redundant decode.

### D5 — Stored-photo fallback stays

`PhotoCmd` keeps re-reading stored database bytes when the in-memory photo is absent (cold start). Rationale: unchanged behavior; only the spec wording loses the "eviction" framing.

## Risks / Trade-offs

- [Unbounded decoded-image memory in a long session] → Accepted; the image-stats diagnostic (separate change) provides data to revisit with a library-backed LRU later.
- [Byte total is an estimate] → It is width*height*4, a heuristic, and only shown in a status readout.

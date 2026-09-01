## Why

The image pipeline ships four bespoke session caches (decoded images, halfblock blocks, raw photo bytes, native renders) that are all plain map-plus-mutex wrappers. The `image-cache-eviction` spec promised byte-budgeted LRU eviction and a decoded-source resolution cap, but that machinery was never implemented, so the spec describes behavior the code does not have. The reader is a single-user terminal app; unbounded session memory is an accepted trade-off, and the database is the real retention layer.

## What Changes

- Collapse `Cache`, `Blocks`, `Photos`, and `Natives` into one generic map-based cache instantiated four times.
- Keep a decoded-image byte total on the cache so the status bar can later report `RAM <n>MB`.
- Remove the `image-cache-eviction` capability (byte-budget LRU eviction and the decoded-source resolution cap) from the spec tree.
- Reconcile `native-image-rendering` with the unbounded design: the native render cache is keyed by render size (content width and viewport-height cap), a viewport-height-only resize re-renders, the source is decoded once per render, and a missing in-memory photo is re-read from stored database bytes.
- Make the native height-fit loop decode the photo once and reuse it across candidate heights (the spec already requires this; the code currently decodes up to three times).

## Capabilities

### New Capabilities

### Modified Capabilities

- `image-cache-eviction`: capability removed (bounded caches and the decoded cap are no longer required).
- `native-image-rendering`: unbounded cache design, render-size cache keying, single decode per render, and stored-photo fallback replace the bounded-cache clauses.

## Impact

- `internal/image`: replace four bespoke cache types with one generic map cache plus a decoded-byte counter.
- `internal/ui`: no change here; the later status-bar change reads the counter.
- `openspec/specs/image-cache-eviction/`: removed at archive time.
- `openspec/specs/native-image-rendering/`: updated.

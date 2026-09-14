## Context

After `extract-image-view-state`, the image concern is already an `imageView` type inside `internal/ui`; the list, article, and popup concerns remain as `listState`/`articleState`/`popupState` plus their methods in the same package. The package is still too big and `Model` still mixes orchestration with rendering.

Two things make a package split here harder than the existing `render`/`compose` extraction: the view types are **stateful** (they read/write `Model` state and hold `tea.Cmd`-returning methods), and they **depend on each other** (article composition calls image block composition). `render` and `compose` worked because they are pure.

## Goals / Non-Goals

**Goals:** `internal/ui` becomes a thin `Model` + four focused view packages; dependency direction is acyclic (views don't import `Model`); behavior byte-identical.

**Non-Goals:** no behavior or invariant changes; no further subdivision of the views in this change; no public API stability guarantee.

## Decisions

### D1 — Dependency direction (acyclic)

```
render (pure)   ◀── compose (pure)
   ▲                   ▲
   │                   │
listview   popup   imageview   articleview
   ▲        ▲          ▲           ▲
   └────────┴──────────┴───────────┘
                  Model (wiring)
```

`articleview` depends on `imageview` (for block composition) and on `render`/`compose`. `imageview` depends on `render`/`compose`. `listview`/`popup` depend on `render`. Nothing imports `Model`. `Model` constructs all views in `New` and routes messages to them in `Update`.

### D2 — Replace `imageView`'s `*Model` back-reference with explicit dependencies

The back-reference introduced in `extract-image-view-state` cannot survive a package boundary (cycle). During the `imageview` move, its methods change to take the article state and geometry they need as parameters and return the signals `Model` consumes (e.g. `needsRecompose bool`, a `tea.Cmd`), instead of reaching into `Model`. This is the "replace interim smell" step `extract-image-view-state` flagged.

### D3 — Order the extraction by fan-out, not fan-in

`listview` first (depends only on `render`), then `popup`, then `imageview`, then `articleview` (depends on `imageview`). This keeps each step compile-green and reviewable, and each package's imports point only at already-extracted packages.

### D4 — Export the view types once, at their final home

`listState`/`articleState`/`popupState`/`imageView` become exported (`State`, `ImageView`, etc.) only when they land in their package; no intermediate aliases. `Model` keeps private references to the concrete types.

### D5 — Pure helpers go to `render`/`compose` first

`overlay`, `spliceStyled`, `formatMB`, and the glyph/status-bar pure parts move to `render` (or `compose`) as a standalone first phase, shrinking `Model` before any stateful move.

## Risks / Trade-offs

- [Very large diff] → Strict phase ordering + `go build ./...` and `go test ./...` after each phase; keep the whole change as a sequence of green commits.
- [Cycle risk during moves] → Follow D1/D2; if a cycle surfaces, stop and re-derive the direction before continuing.
- [Exported-type churn] → Mechanical; tests move with their subject so churn is self-correcting.

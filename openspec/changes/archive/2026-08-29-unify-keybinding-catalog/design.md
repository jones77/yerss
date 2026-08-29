## Context

See `proposal.md` - Why. Current sources of truth in `internal/config`:

- `keys.go`: `Action`/`View` consts, `DefaultKeybindings()` (runtime defaults),
  `allActions` (valid set), `actionsForView(v)` (per-view validity).
- `popup.go` (`internal/ui`): hardcoded help order.
- `releases.go`: `ConfigReleases` (v1 keybinding rows, `Doc: ""`).
- `bootstrap.go`: `seededConfig()` (commented keybinding lines).

`Action` is a `string` type; ordering matters for both the seeded template and
the help screen, so the catalog must preserve a deterministic order (a slice,
not a map).

## Goals / Non-Goals

**Goals:**

- One ordered table is the sole definition of every action's identifier,
  default keys, and valid views.
- The seeded template and self-update registry are *generated* from that table,
  eliminating the current drift.
- The help popup lists exactly the catalog's actions.

**Non-Goals:**

- No new actions, no changed default bindings, no new views.
- No routing of structural fold keys (`h`/`l`/Enter/Space/Tab on day headers)
  through the keymap — those remain contextual overrides.
- No change to `normalizeKey` / `ParseKeybindings` / `EffectiveKeys` logic.

## Decisions

**D1 — A `catalog []actionSpec` slice with `{action Action; keys []string; views []View}`.**

Lives in `keys.go`. Order is the canonical display/seed order (mirror the current
`DefaultKeybindings` order). Rationale: a slice keeps ordering deterministic for
the seeded template and help screen, and one struct per action makes the "views"
dimension explicit where today `actionsForView` re-lists it by hand.

**D2 — Derive `DefaultKeybindings`, `allActions`, `actionsForView` from `catalog`.**

`DefaultKeybindings()` maps `catalog` → `Keymap`. `allActions` becomes
`map[Action]bool` built from `catalog`. `actionsForView(v)` filters `catalog` by
`spec.views`. The behavior (key → action resolution) is unchanged because the
catalog encodes today's values.

**D3 — Generate the `keybindings` section of `seededConfig` and the keybinding
rows of `ConfigReleases` from `catalog`.**

- `ConfigReleases` keeps its explicit non-keybinding rows (`data`, `refresh`,
  `display`); a helper (e.g. `keybindingOptions()`) returns the keybinding rows
  by rendering each catalog entry as `{Version: 1, Section: "keybindings",
  Key: string(action), Default: tomlArray(keys), Doc: ""}`.
- `seededConfig()` renders its `[keybindings]` block by iterating `catalog`,
  emitting `# <action> = ["k1", ...]`.
- This makes the seeded template and the registry derive from the same source as
  runtime defaults, closing the drift at the root rather than patching the two
  stale strings.

- *Alternative considered*: just fix the stale strings in `bootstrap.go` and
  `releases.go`. Rejected — a one-off fix leaves the split-brain intact and the
  next action added will re-introduce the drift.

**D4 — Export `config.AllActions() []Action` (or `Catalog()`) for the help popup.**

`popup.go`'s `renderHelp` iterates the exported list instead of its hardcoded
slice, keeping display order from the catalog. The `actionLabel` helper is
unchanged.

**D5 — Extend the drift test to close the loop.**

The existing test compares seeded template ↔ `ConfigReleases`. Add assertions
that: `DefaultKeybindings()` equals the catalog's key map; every action in
`allActions` appears in the catalog and vice versa; and the seeded `[keybindings]`
block contains every catalog action with its exact default keys.

## Risks / Trade-offs

- **Ordering sensitivity** of the seeded template and help screen → Mitigation:
  a slice with explicit canonical order, and the drift test pins it.
- **`ConfigReleases` generation must stay a pure append** so self-update is still
  idempotent → Mitigation: keybinding rows keep `Version: 1` and the same
  `Section`/`Key`/`Default` shape; existing migrate tests cover idempotency.
- **Help popup output change** (it now always matches the catalog) → Mitigation:
  this is intended; the reader-ui spec already requires the help to list "each
  action and its bound keys", and the catalog is the canonical action set.

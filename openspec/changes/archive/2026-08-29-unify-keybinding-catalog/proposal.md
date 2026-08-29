## Why

The set of actions and their default keybindings is enumerated in five places —
`DefaultKeybindings`, `allActions`, and `actionsForView` in `keys.go`, the
hardcoded help list in `popup.go`, the keybinding entries in `ConfigReleases`
(`releases.go`), and the seeded template in `seededConfig` (`bootstrap.go`).
Only two of those five are protected by a drift test, and the catalog has
already drifted: the seeded template and `ConfigReleases` are missing the `b`
alias for `back`, the `o` alias for `open_article`, the `t`/`r` case aliases,
and the entire `copy_article_text` action — all of which the configuration spec
and `DefaultKeybindings` already define. A new user reading the seeded template
gets stale defaults, and the next schema bump would self-update from a stale
registry.

## What Changes

- Introduce a single action catalog in `internal/config` — one ordered table
  holding, per action: its identifier, default keys, and applicable views — and
  derive `DefaultKeybindings`, `allActions`, and `actionsForView` from it.
- Generate the `keybindings` section of the seeded template and the keybinding
  entries of `ConfigReleases` from the same catalog, so the defaults can no
  longer drift from the runtime behavior.
- Derive the help popup's action list (`popup.go`) from the catalog instead of a
  hand-written slice, so the help screen always lists every bindable action.
- As a direct result, the seeded template and self-update registry now include
  `b`, `o`, `t`, `r`, and `copy_article_text`, matching the spec and runtime.
- Keep the existing drift test (seeded template ↔ `ConfigReleases`) and extend it
  to assert the catalog ↔ `DefaultKeybindings` chain is consistent.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

_None._ Behavior (which keys map to which actions) is already specified and
unchanged; this is a de-duplication and a conformance fix for the seeded
template, so this change declares `skip_specs: true`.

## Impact

- **Config layer**: `internal/config` — `keys.go` (catalog + derivations),
  `releases.go` (generated keybinding entries), `bootstrap.go` (generated
  keybinding template), plus the existing drift test.
- **UI layer**: `internal/ui/popup.go` — help list derived from the catalog via
  a small exported accessor (e.g. `config.Catalog()` or `config.AllActions()`).
- **Tests**: extend the config drift test; help popup still lists every action
  in the same order; keybinding parse/validate/conflict tests unchanged.
- **Out of scope**: routing the structural day-header fold keys (`h`/`l`/
  Enter/Space/Tab) through the keymap — those are contextual overrides of
  already-bound keys and stay as-is.

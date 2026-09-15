## Why

The article reader's compact/expanded link display mode (toggled with `x`) is
currently session-scoped: it survives switching between articles during one run,
but a restart always falls back to compact. A reader who keeps expanded mode on
to inspect link targets (and the expanded header, which shows the article URL
above the title) has to press `x` again at the start of every session. The mode
should be remembered across restarts like the saved reader selection.

## What Changes

- The `x` toggle's compact/expanded choice is persisted to the SQLite `state`
  table (the same mechanism that stores `last_selection`) when the app quits,
  and restored at startup before the first article composes.
- Expanded mode renders the article URL as the top line of the reader, a blank
  line, then the title; compact mode omits the URL line so the title is first.
  This rendering already exists and is covered by tests; the change pins it as
  part of the persisted-mode contract and adds coverage tying the persisted
  mode to the rendered header.
- The persisted value is a runtime state, not a config option: `state` table
  row keyed independently of the config file, so toggling never rewrites
  `config.yaml`.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `compact-link-mode`: the mode SHALL persist across application restarts via
  the DB (not just across article switches within a session), and the expanded
  header layout (URL line above the title) is specified as the behavior the
  restored mode drives.

## Impact

- `internal/store/store.go`: add a `state`-table-backed save/load pair for the
  link-expand mode, mirroring `SaveLastSelection`/`LoadLastSelection`.
- `internal/app/session.go`: delegate the new store calls to the UI layer.
- `internal/ui/selection.go`: persist the mode alongside `persistSelection` on
  quit.
- `internal/ui/model.go`: load the persisted mode in `Init` before
  `restoreSelection`, applying it via `SetLinkExpand`.
- `internal/ui/` tests: extend the link-expand suite to cover the DB round-trip
  and that a restored expanded mode renders the URL-above-title header.
- No schema migration needed: `state` is a generic key/value table already
  present.
## Context

The compact/expanded link display mode is currently session-scoped: `Model.linkExpand`
(model.go) is session truth, toggled by `x`, and it survives article switches within
one run but resets to compact on restart. The mode must instead persist across
restarts through the SQLite `state` table — the same generic key/value table that
already holds `last_selection` — mirroring the existing selection persistence
(see proposal.md - What Changes for the motivation).

Persistence seams already exist and are battle-tested:

- `store.SaveLastSelection` / `store.LoadLastSelection` (store.go) upsert/read a
  JSON value under a `state`-table key. A second key for the link-expand mode
  follows the identical pattern with no schema migration.
- `persistSelection` (selection.go) already runs on every quit path (list, article,
  popup). It is the natural single hook to also record the mode.
- `Model.Init` (model.go) calls `restoreSelection` before the first article
  composes; loading the persisted mode just before that, then applying it via
  `SetLinkExpand`, guarantees the saved mode drives the header of the first
  article (spec: "Restored expanded mode renders the URL-above-title header").
- The header rendering contract (spec: URL line, blank line, title) already exists
  in `renderHeaderMode`/`renderHeaderLink` (article.go) and is covered by
  `link_expand_test.go`; this change pins it as part of the persisted-mode
  contract and adds a round-trip test tying the restored mode to the rendered
  header — no header rendering work is expected.

## Goals / Non-Goals

**Goals:**

- Persist the compact/expanded mode to the `state` table on quit and restore it at
  startup, so a restart reopens in the last-used mode.
- Expose the mode through the app session layer (`internal/app/session.go`) so the
  UI layer does not reach into the store directly.
- Cover the DB round-trip and the rendered-header consequence of a restored mode
  with tests.

**Non-Goals:**

- No config-file involvement: the mode stays a runtime state and toggling never
  writes `config.yaml` (spec: "Toggling does not touch the config file").
- No schema migration: the `state` table already exists.
- No new user-facing controls or changes to the `x` toggle behavior itself.

## Decisions

**1. Store the mode under its own `state` key as a tiny JSON envelope.**
A key like `link_expand_mode` holding `{"expanded":true}` — or a plain "1"/"0"
value — keeps the row self-describing and extensible, mirroring how
`last_selection` is a JSON blob. Chosen over reusing the `LastSelection` JSON
struct: the mode is orthogonal to the reader selection (it also applies before an
article is open) and should not ride along in a struct whose semantics are about
where the reader left off. Alternatives considered: reusing `LastSelection` (rejected:
wrong coupling), writing to the config file (rejected: explicit non-goal, would
rewrite user config on every toggle).

**2. `Session` delegates the mode save/load to the store.**
`Session` already wraps `store` for selections (`SaveSelection`/`LoadSelection`).
Adding `SaveLinkExpandMode`/`LoadLinkExpandMode` keeps the UI layer on the session
interface, consistent with how `Model.sess` is used elsewhere. Rationale: the
store is opened in `main.go` and handed to the session; the UI never touches the
store directly.

**3. Persist from the existing `persistSelection` hook on quit.**
`persistSelection` is already invoked by every quit path. Adding the mode write
there means no new quit-path wiring and no risk of a mode not being saved on some
exit path. Rationale: single, already-correct save point. Alternative considered:
a separate save call at each quit site (rejected: duplicates wiring).

**4. Load in `Init` before `restoreSelection`, applied via `SetLinkExpand`.**
`SetLinkExpand` already syncs `Model.linkExpand` with `articleview.State.LinkExpand`
so the first composition picks up the mode. Loading before `restoreSelection`
ensures an article reopened by the selection restore composes in the saved mode
from its very first render. Rationale: zero new plumbing into the composition
path; the existing setter is the contract. Alternative considered: applying after
restore (rejected: would compose the restored article once in the wrong mode).

**5. Missing value = compact.**
`LoadLinkExpandMode` returns compact for an absent row, satisfying the
"no saved mode means compact" scenario without an explicit default write.

## Risks / Trade-offs

- [A future save with a corrupt/unknown value silently falls back to compact]
  → Mitigation: the loader returns compact on unmarshal failure, and the save
  writes only a boolean under a fixed key.
- [Mode write happens only on quit; a crash loses a mid-session toggle]
  → Mitigation: acceptable — the same trade-off already applies to the reader
  selection; the toggle is one keystroke.
- [Restoring expanded mode changes the header height of the first article]
  → Mitigation: the existing header-render path already accounts for the
  expanded mode's extra URL line; no layout work is needed (spec pins the
  rendered behavior, covered by the round-trip test).

## Migration Plan

No data migration: the `state` table is keyed generically and the new key is
created on first save. Rollback is trivial — removing the two new session/store
methods and the Init load reverts to session-scoped behavior with no residual
state-row cleanup required (an orphan row is inert).
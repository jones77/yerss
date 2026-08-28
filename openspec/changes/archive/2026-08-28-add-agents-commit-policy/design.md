## Context

See `proposal.md` for motivation. The repo has no `AGENTS.md`, an empty
`openspec/config.yaml` template, and zero commits despite completed, archived
OpenSpec changes. The OpenSpec `archive` operation moves a change directory into
`openspec/changes/archive/<date>-<name>/` and syncs delta specs into the main
`openspec/specs/` tree; it does not itself commit. Today that leaves the working
tree dirty after every archive, and commits — when they happen — follow no
convention.

## Goals / Non-Goals

**Goals:**
- Give agents a single, discoverable file (`AGENTS.md`) with the repo's working
  rules, including the commit-message standard.
- Make the OpenSpec archive step produce a correctly-formatted commit automatically.
- Surface the commit convention in `openspec/config.yaml` `context` so it is visible
  to agents at artifact-creation time, and in `operations.archive.guidance` so the
  archive step enforces auto-commit.

**Non-Goals:**
- Replacing OpenSpec's archive mechanics (we layer a commit on top, not alter the
  move/sync).
- Enforcing commit conventions via a git hook or CI lint in this change (that
  belongs to the separate `add-build-gate-ci` change's linter scope).
- Mandating conventional-commits prefixes (`feat:`, `fix:`). The Chris Beams guide
  does not require them; we follow the guide as written.

## Decisions

**D1 — Conventions live in `AGENTS.md`; `config.yaml` references them.** The full
seven-rule summary and the canonical link go in `AGENTS.md` (the file agents read
first). `openspec/config.yaml` `context` carries a one-line pointer ("Commits
follow https://cbea.ms/git-commit/") rather than duplicating the rules, so there is
one source of truth. *Alternative:* put the rules only in `config.yaml` context —
rejected, `AGENTS.md` is the conventional discovery point for agent guidelines and
is read even outside OpenSpec workflows.

**D2 — Archive auto-commit is `operations.archive.guidance`, not a script.** OpenSpec
config's `operations.archive.guidance` is the documented hook for "what the archive
step must do." We add guidance lines stating the agent SHALL, at the end of archive,
stage and commit the archived change + synced specs + implementation edits in one
commit following the policy. This keeps the requirement portable and visible in
`openspec context`/`view`. *Alternative:* add a wrapper shell script that runs
`openspec archive` then `git commit` — rejected, it duplicates OpenSpec's own
flow and is fragile across environments; guidance is the intended extension point.

**D3 — One commit per archive, imperative subject scoped to the change.** The
auto-commit message uses the form `Archive <change-name>` as the subject (imperative,
under 50 chars), with a body explaining what was implemented and that specs were
synced. The change name keeps subjects unique and greppable. *Alternative:* split
implementation and archive into two commits — rejected, the archive is the
finalization event; one atomic commit keeps `openspec/specs/` and the archive
directory consistent in history.

**D4 — `AGENTS.md` scope is agent rules, not a project README.** `AGENTS.md` covers
commit policy, the OpenSpec workflow expectations (work within `tasks.md`, update
checkboxes, verify before marking done), and a pointer to the project's specs. It
is not a user-facing README (that is created by the separate `add-build-gate-ci`
change). *Alternative:* merge README and AGENTS.md — rejected, they serve different
audiences (users vs. agents).

## Risks / Trade-offs

- **[Guidance is advisory, not enforced by a hook]** → A guideline cannot
  mechanically prevent a bad commit. Mitigation: the `add-build-gate-ci` change adds
  CI + linters; a future commit-lint (e.g., commitlint) could enforce the seven
  rules mechanically. This change establishes the standard agents follow.
- **[Auto-commit may race with a dirty tree]** → If unrelated edits are staged when
  archive runs, the auto-commit could sweep them in. Mitigation: the archive
  guidance SHALL scope the commit to the change's files (the archived directory, the
  synced `openspec/specs/` paths, and the implementation files touched by the
  change), staging those paths explicitly rather than `git add -A`.
- **[Single archive commit could be large]** → A big implementation + archive in one
  commit is harder to review. Mitigation: implementation commits during apply remain
  separate if the user requests them; the archive commit is the finalization. Accept
  the tradeoff for spec/history atomicity.

## Migration Plan

No data migration. Apply the change by creating `AGENTS.md` and editing
`openspec/config.yaml`. Rollback is deleting `AGENTS.md` and reverting the config
edits.

## Open Questions

- Should the archive commit also push to the remote automatically, or stop at a
  local commit? Deferred — local commit is the safe default; pushing is a separate
  policy decision that can be added to guidance later without changing this design.

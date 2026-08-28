## Why

The project has no `AGENTS.md` and no commit conventions, so AI agents working in
this repo have no shared rules for how to write commits — and the OpenSpec archive
step leaves changes uncommitted, requiring a manual commit that is easy to forget
(the repo currently has zero commits despite completed, archived work). Establishing
a single commit-message standard and making archive auto-commit a requirement gives
agents consistent, reviewable history without manual intervention.

## What Changes

- **Add `AGENTS.md`:** Create a root `AGENTS.md` that defines agent guidelines for
  this repo, including a commit-message policy. Agents SHALL write commit messages
  following the seven rules from <https://cbea.ms/git-commit/> (separate subject
  from body with a blank line; limit subject to 50 characters; capitalize the
  subject; no trailing period; imperative mood in the subject; wrap the body at 72
  characters; use the body to explain what and why, not how). The file SHALL
  summarize the seven rules and link the guide as the source of truth.
- **Auto-commit on archive:** The OpenSpec archive operation SHALL conclude by
  committing the archived change, the synced main specs, and the implementation
  edits in a single commit whose message follows the commit policy above. This is
  recorded as `operations.archive.guidance` in `openspec/config.yaml` so the
  archive step enforces it.
- **Config context:** Populate `openspec/config.yaml` `context` with the tech stack
  and the commit convention reference, so agents creating artifacts see the standard
  up front.

## Capabilities

### New Capabilities
<!-- None — this is a docs/process change with no application behavior changes. -->

### Modified Capabilities
<!-- None — no yerss capability requirements change. -->

This change opts out of specs (`skip_specs: true` in `.openspec.yaml`): it adds
project documentation and development-workflow guidance, not application behavior.

## Impact

- **Code:** None. No application source changes.
- **Files:** New `AGENTS.md` (repo root); `openspec/config.yaml` gains `context`
  and `operations.archive.guidance` entries.
- **APIs:** None.
- **Dependencies:** None added.
- **Behavior:** Agent-authored commits now follow a defined convention, and the
  archive step produces a commit automatically instead of leaving the working tree
  dirty.

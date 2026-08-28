## 1. AGENTS.md

- [x] 1.1 Create `AGENTS.md` at the repo root defining agent guidelines: the OpenSpec workflow expectations (work within `openspec/changes/<change>/tasks.md`, update checkboxes in real time, verify with tests/linter/build before marking a task `[x]`) and a pointer to the project's `openspec/specs/`
- [x] 1.2 Add the commit-message policy to `AGENTS.md`: agents SHALL follow the seven rules from https://cbea.ms/git-commit/ — (1) separate subject from body with a blank line, (2) limit subject to 50 chars, (3) capitalize the subject, (4) no trailing period, (5) imperative mood in subject, (6) wrap body at 72 chars, (7) body explains what and why, not how — with the link as the source of truth

## 2. openspec/config.yaml

- [x] 2.1 Populate the `context` block with the tech stack (Go 1.27, bubbletea/lipgloss TUI, SQLite via glebarez/go-sqlite, gofeed) and a one-line commit-convention pointer to https://cbea.ms/git-commit/
- [x] 2.2 Add `operations.archive.guidance` requiring that the archive step SHALL, as its final action, stage and commit the archived change directory, the synced `openspec/specs/` paths, and the implementation files touched by the change in a single commit whose message follows the AGENTS.md commit policy; stage those paths explicitly (not `git add -A`)

## 3. Verification

- [x] 3.1 Run `openspec validate --all --strict` and confirm the change validates with `skip_specs: true` (no spec files expected)
- [x] 3.2 Confirm `AGENTS.md` is discoverable at the repo root and `openspec/config.yaml` parses (YAML valid)
- [ ] 3.3 (This change's own archive) On archive, follow the new policy: commit `AGENTS.md`, `openspec/config.yaml`, and this change's archive directory in one commit with an imperative, ≤50-char subject following the seven rules

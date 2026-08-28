# AGENTS.md

Guidelines for AI agents working in this repository.

## Project

`yerss` is a terminal RSS reader written in Go: a bubbletea/lipgloss TUI, SQLite
for storage, and gofeed for feed parsing. Behavioral requirements live in
`openspec/specs/`.

## OpenSpec workflow

- Work strictly within `openspec/changes/<change>/tasks.md` for the active change;
  update task checkboxes (`- [ ]` → `- [x]`) in real time as work completes.
- Never mark a task `[x]` without executing tests, the linter, or a build that
  proves the change works.
- Keep changes clean and localized; run `git status` / `git diff` at task
  boundaries to avoid unintended side effects.
- Consult `openspec/changes/<change>/specs/` (delta specs) and the main
  `openspec/specs/` tree for the behavior requirements you are implementing.

## Commit messages

Commit messages SHALL follow the seven rules from
https://cbea.ms/git-commit/ (the source of truth):

1. Separate subject from body with a blank line.
2. Limit the subject line to 50 characters.
3. Capitalize the subject line.
4. Do not end the subject line with a period.
5. Use the imperative mood in the subject line.
6. Wrap the body at 72 characters.
7. Use the body to explain what and why, not how.

Changes are committed at OpenSpec archive time. When archiving, stage the archived
change directory, the synced `openspec/specs/` paths, and the change's
implementation files explicitly (not `git add -A`) and commit them in one message
following the policy above.
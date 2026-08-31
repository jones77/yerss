## Open Questions

None blocking: the remaining work is verification and bookkeeping, not new
behavior. The one deliberate tuning knob is the caption-width fraction (70%
from `standard-caption-width`); a real-terminal smoke test may suggest a
different value, but that would be a follow-up change, not part of this one.

## Options

- Finish the smoke tests and archive `inline-image-rendering` in this change
  (chosen) vs leaving the change active indefinitely. The implementation is
  done and green; keeping it active only defers the spec sync and the commit of
  the entangled working tree.
- Commit the shared implementation files at `inline-image-rendering`'s archive
  (chosen, per AGENTS.md "changes are committed at archive time") vs committing
  them now. Committing now would split the two changes' interleaved diffs
  arbitrarily; archiving first keeps the whole image-rendering implementation
  in one commit.

## Problems

- The working tree's image-rendering implementation files are interleaved
  across two changes (`inline-image-rendering` active, `standard-caption-width`
  archived). There is no clean line-level split, so both are committed together
  when the remaining active change archives. Do not try to stage individual
  hunks.
- The smoke tests are user-run and terminal-dependent; automated tests cannot
  substitute for them. The geometry-sensitive snap/preview logic was verified
  against the real article geometry during development (temporary debug tests
  reading the local SQLite DB); reproduce that technique if a fix is needed.

## Details

### 1. Re-run the pending smoke tests

Rebuild and open the two stored articles, then work the task list in
`openspec/changes/inline-image-rendering/tasks.md`:

- Intercept article "Ecocide in Iran" — tasks 9.1, 10.3, 11.4 (and 6.3/7.5,
  which are the earlier superseded smoke-test rounds; mark them done once 9.1
  passes).
- Hurricane article (`.../hurricane-helene-ice-asheville-nc-g20-trump/`) —
  task 12.3.

If a check fails, fix the code in `internal/ui/image.go` (snapYOffset,
scrollArticle, previewClippedImage) and `internal/ui/article.go`
(inlineBodyParts, nextSentinel), re-run `go test ./...`, then re-test. To
reproduce geometry-sensitive issues, write a temporary test that opens the
store, renders the native blocks with the real fit, and simulates the scrolls
(as done in earlier rounds).

### 2. Archive inline-image-rendering

1. Mark the completed smoke-test tasks `[x]` in its `tasks.md`.
2. Run `openspec validate inline-image-rendering`.
3. Sync its delta specs (`specs/inline-image-rendering/spec.md`,
   `specs/reader-ui/spec.md`) into the main specs
   (`openspec/specs/inline-image-rendering/spec.md` is new;
   `openspec/specs/reader-ui/spec.md` already carries the standard-caption-width
   changes and gains the inline-image/scroll requirements). Verify each delta's
   requirements/scenarios land and nothing is lost.
4. Move it to `openspec/changes/archive/2026-08-31-inline-image-rendering/`.

### 3. Commit the archive

Follow AGENTS.md: stage the archived change dir, the synced `openspec/specs/`
paths, and the change's implementation files explicitly (no `git add -A`). The
implementation files also carry the `standard-caption-width` implementation and
the `AGENTS.md` additions, so the commit captures the whole image-rendering
effort. Commit message follows the seven rules
(https://cbea.ms/git-commit/).

### Tests

- `go test ./...` must stay green throughout; `openspec validate --all` after
  the spec sync.
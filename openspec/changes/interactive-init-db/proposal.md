## Why

`yerss -z` / `--init-db` deletes the SQLite database (and its `-wal`/`-shm`
sidecars) with no confirmation. Now that the database is expected to grow to
hundreds of megabytes of stored images, an accidental `-z` destroys real data
with a single flag. The standard Unix guard for this is `-i` / `--interactive`
(`mv`, `cp`, `rm` all use it): the destructive operation prompts before acting,
and refuses on a non-interactive stdin.

## What Changes

- Add an `-i` / `--interactive` boolean flag. It is a no-op for every flag
  except `-z` / `--init-db`.
- `-z` alone keeps today's behavior: delete immediately, no prompt.
- `-z -i` (and `--init-db --interactive`) prompts before deleting, printing the
  size of the database and the number of articles it holds, then reads a `y/N`
  confirmation from stdin.
- On a non-interactive stdin (no TTY, e.g. a pipe or CI), `-z -i` refuses to
  delete and exits non-zero, leaving the database untouched. `-z` alone still
  deletes (backwards compatible).
- The usage output's data-file block drops its per-file descriptions, leaving
  just the resolved paths (the paths are self-explanatory).

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `configuration`: a new `-i` / `--interactive` flag guards `-z` / `--init-db`
  with a confirmation prompt that reports the database size and article count,
  and refuses on non-interactive stdin.
- `feed-pipeline`: the "Usage output documents the data files" requirement drops
  the one-line descriptions, listing only the resolved paths and the `[data]`
  relocation note.

## Impact

- `cmd/yerss/main.go`: register `-i` / `--interactive`; reorder the `-z` path so
  it can stat the database file and count articles before deleting; add a
  `confirmInitDB` helper that prompts and reads stdin; drop the `Files:` block
  descriptions from `pflag.Usage`.
- `cmd/yerss/main_test.go` and a new stdin prompt test exercise the confirm /
  refuse paths (interactive yes, interactive no, non-tty refuse).
- No new external dependencies (TTY detection uses `os.Stdin.Stat()` mode bits).

**Sequencing note:** `-z` / `--init-db` was introduced by the
`tags-popup-column-grid` change, which is complete but not yet archived; its
`configuration` delta (which documents `-z`) is not yet in `openspec/specs/`.
This change's `-i` requirement references `-z` and is independent of it, so the
two archive cleanly in any order.

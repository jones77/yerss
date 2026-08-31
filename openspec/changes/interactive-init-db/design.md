## Open Questions

None. Decisions locked: prompt prints DB size + article count; non-tty aborts.

## Options

- `-i` opt-in prompt (chosen) vs always-prompt with `-f`/`--force` to skip.
  Chosen `-i` opt-in because it is backwards compatible (existing `-z` scripts
  keep working) and matches `mv`/`cp`/`rm` semantics the user asked for.
- Print article count via `SELECT COUNT(*)` (chosen) vs walking files. Chosen
  the SQL count because it is exact and the store is already opened nearby.

## Problems

- Reading the article count requires an open database handle, but `resetDatabase`
  currently runs before `store.Open`. Resolution: when `-i` is set, open the
  store first (count articles), then close it before deleting, then let the
  normal `store.Open` path reopen. A missing or corrupt database is not an
  error for the prompt — treat it as 0 articles / 0 bytes and still prompt (or
  skip the prompt since there is nothing to lose; see below).

## Details

### Flag wiring

In `cmd/yerss/main.go`, alongside the existing `pflag.BoolVarP` calls:

```go
var interactive bool
pflag.BoolVarP(&interactive, "interactive", "i", false, "prompt before --init-db")
```

### Confirm flow

Replace the current unconditional `if initDB { resetDatabase(...) }` block with:

```
if initDB && interactive {
    if code, ok := confirmInitDB(cfg.DBPath()); !ok {
        os.Exit(code)      // refused, or error; database untouched
    }
}
if initDB {
    resetDatabase(cfg.DBPath())
    // print notice (existing)
}
```

`confirmInitDB(path)`:

1. Stat the database file. `size := 0` when it is missing (`os.IsNotExist`).
2. If the file exists, open the store read-only-ish to count articles
   (`SELECT COUNT(*) FROM articles`). If the open or count fails, treat the
   count as 0 but still prompt (the file exists and deleting it is still
   destructive). If the file is missing (size 0), skip the prompt entirely and
   return ok (nothing to lose — matches "a missing database is not an error").
3. Detect a TTY on stdin: `stat, _ := os.Stdin.Stat(); tty := stat.Mode()&os.ModeCharDevice != 0`.
   If not a TTY, print to stderr `"%s: -i requires a terminal; refusing to delete %s"` and return `ok=false` with exit code 1.
4. Print to stderr the confirmation line and read one line from `os.Stdin`:
   ```
   delete database (123.4 MB, 1203 articles) at <path>? [y/N]
   ```
   Trim the line; accept only `y`/`Y` (case-insensitive). Anything else — empty,
   `n`, `N`, EOF, garbage — is a refusal: print `"%s: not deleting %s"` and
   return `ok=false` with exit code 0.
5. Return `ok=true` on `y`.

Exit codes: refusal on a real answer exits 0 (the user chose not to delete),
consistent with `mv -i` declining. Non-tty exits 1 (the operation could not be
confirmed).

### Size formatting

Reuse the same human-readable size formatting the status bar uses. If it is
private to `internal/ui`, add a tiny local helper in `cmd/yerss` (bytes → B/KB/
MB/GB with one decimal) and unit-test it; do not export UI internals for this.

### Usage trim

In `pflag.Usage`, the `Files:` block currently reads:

```
Files:
  config  <path>   display, refresh, and keybinding settings
  feeds   <path>   one feed per line; the https:// scheme is assumed
  db      <path>   SQLite article store
```

Change it to list only the labels and resolved paths, keeping the `[data]`
relocation note:

```
Files:
  config  <path>
  feeds   <path>
  db      <path>

The config's [data] section can relocate the feeds file and database.
```

### Tests

- `main_test.go`: `-i` flag parses to true; `-i` with `-j`/`-a`/`-e`/`-c` is a
  no-op (no prompt, normal behavior).
- New `confirmInitDB` tests (table-driven, with a temp DB and a stubbed stdin):
  - missing DB file → returns ok without prompting;
  - TTY + `y` → ok, correct size/count text;
  - TTY + `n`/`<empty>`/`garbage` → not ok (exit 0), database untouched;
  - non-TTY → not ok (exit 1), database untouched.
- Verify `yerss -h` output no longer lists the per-file descriptions.

## Remaining Decisions

None.

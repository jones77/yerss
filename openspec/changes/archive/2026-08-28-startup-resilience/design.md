## Context

`yerss` currently launches unconditionally: `main` loads config, opens the store,
and immediately enters bubbletea's `Run()`. Two problems surface on a fresh
machine. First, `renderList` panics with `index out of range [-1]` because
`View()` is called before the first `tea.WindowSizeMsg` reaches `Update`, leaving
`m.height` at its zero value; the `lines[:m.height]` truncation then empties the
slice and `lines[len(lines)-1]` indexes `-1`. Second, `config.Load` silently
returns built-in defaults when the config file is missing and never creates the
config directory, `config.toml`, or `feeds.txt`, so a new user has nothing to
edit and no empty-feed state is surfaced until the TUI has already taken over the
terminal. See proposal.md for motivation; the specs define the required behavior.

Key existing facts that shape the design:
- `feed.UpsertFeed` is called **only** inside the `gofeed` success branch
  (`fetch.go`), so a row in the `feeds` table is proof that URL has parsed
  successfully at least once.
- `store.LastRefreshedAt()` derives from `MAX(last_fetched_at)` over the `feeds`
  table; there is no `meta` table.
- `Init()` already gates a startup refresh via `feed.Gate.NeedsStartupRefresh`.
- Tests build a `Model` headless and call renderers directly (`render_test.go`),
  but `newTestModel` hardcodes 80×24, which is why the zero-dimension panic was
  never caught.

## Goals / Non-Goals

**Goals:**
- Make every renderer crash-proof for any non-negative terminal dimensions.
- Materialize a real config directory and seed files on first run.
- Block TUI entry until at least one RSS feed has been verified live, without
  trapping offline users after the first successful verification.
- Avoid double-fetching feeds on the first run.
- Keep the config template and a registry of options in sync as the schema
  evolves, with a cheap drift check.

**Non-Goals:**
- In-app feed management (feeds remain edited in the text file).
- Re-validating already-trusted feeds at every startup.
- A CLI escape-hatch flag for the gate (the trust rule covers the offline case).
- Migrating or rewriting existing user `config.toml` values; self-update only
  appends comments and never alters live keys.

## Decisions

### D1: The gate runs in `main`, before `tea.NewProgram`

The verification gate executes in `cmd/yerss/main.go` after `store.Open` and
before `p.Run()`. Refusing to start must happen before the alt-screen TUI takes
over the terminal; bailing out from inside bubbletea requires panic recovery and
"Restoring terminal…" dance (visible in the original crash trace), which is a
poor user experience and harder to give a clean exit code.

**Alternatives considered:** Running the gate inside `ui.Init`. Rejected — by the
time `Init` runs the terminal is already in alt-screen mode, so a refusal would
tear down the TUI ungracefully and exit codes would be awkward to surface.

### D2: Offline trust signal reuses the existing `feeds` table (no new table)

A non-empty `feeds` table means at least one URL has parsed successfully before
(`UpsertFeed` is only called on parse success). The gate treats
`store.HasVerifiedFeed()` (a `SELECT COUNT(*) FROM feeds > 0` query) as the
"trust and skip the network check" signal. This makes every launch after the
first successful verification offline-safe, with no schema migration and no new
meta table.

**Alternatives considered:** A dedicated `meta(key, value)` table with a
`feeds_verified` boolean. Rejected — it duplicates information already
derivable from the `feeds` table and requires a migration. A marker file in the
data directory. Rejected — less robust than the database the app already owns.

### D3: The gate's fetch IS the first refresh (no double fetch)

On a fresh DB, the gate calls `FetchFeeds(store, urls)` directly (bounded by
timeouts — see D6). Because `UpsertFeed` stamps `last_fetched_at`, on success
`store.LastRefreshedAt()` returns "now", so `Init()`'s
`NeedsStartupRefresh(last, now)` returns false and the startup refresh is
skipped. The gate and the first refresh are therefore the same operation, with
zero redundant network work.

**Alternatives considered:** A lightweight "verify one feed, early-exit" gate
that leaves the full refresh to `Init`. Rejected — it re-fetches the verified
feed moments later and complicates the success criterion (when to stop). A full
`FetchFeeds` with `wg.Wait()` and "at least one succeeded" semantics is simpler
and gets all feeds into the DB before the TUI even starts.

### D4: Render safety is fixed at two layers

- **Layer B (defensive, localized):** `renderList` clamps the effective height to
  a minimum of 1 before the `lines[:height]` truncation, and skips overwriting
  the last line with the status bar when the resulting slice is empty. This
  makes the function crash-proof regardless of caller.
- **Layer A (structural):** `New()` seeds the model with a default width/height
  (e.g. 80×24) so the first frame renders something sane before
  `WindowSizeMsg` arrives; the real dimensions overwrite these on the first
  `WindowSizeMsg`.

Both layers are implemented because each guards a different failure mode: Layer A
prevents a degenerate first frame, Layer B prevents a crash from any 0-row
terminal (CI, piped output, unusual environments).

**Alternatives considered:** Only fixing Layer A. Rejected — a 0×0 terminal would
still crash. Only fixing Layer B. Rejected — the first frame would render into a
0×0 box, producing a blank/janky screen before `WindowSizeMsg`.

### D5: The config self-update is driven by an integer schema version, not the git commit

`releases.go` defines `ConfigSchemaVersion` (an `int` that increments whenever a
config option is added) and an ordered registry of additions per version. The
seeded `config.toml` header carries both the schema version and the build commit
(the commit is human-readable display only, defaulting to `"dev"` under plain
`go run` with no ldflags). The append-migration logic compares the file's
recorded schema version to the current one. Using a monotonic integer makes the
logic deterministic and independent of the build process; the commit string is
never parsed for control flow.

**Alternatives considered:** Driving the logic off the git commit hash directly.
Rejected — under `go run` without `-ldflags -X ...commit=...` the value is empty,
so the comparison would need a fallback anyway; commits are not monotonic and a
rebase could move the file's recorded commit "ahead" of the current one.

### D6: Self-update idempotency is keyed on option presence, not the version bump

Before appending an option introduced in a later schema version, the system
checks whether that option's key already appears anywhere in the file (commented
or live). Only absent keys are appended. The header schema-version bump is
secondary bookkeeping. This makes the operation safe to re-run and robust to
aborted writes: even if a previous append succeeded but the header bump failed,
the next run will not duplicate the option because the key is now present.

**Alternatives considered:** Trusting the header version bump as the sole
"already done" signal. Rejected — a crashed write between append and bump would
strand the file in a state where the option exists but the version says it
doesn't, causing duplicates on the next run.

### D7: Template/registry drift is caught by a test, not a generator

The full commented template is a maintained constant, and `releases.go` holds the
structured additions registry. A unit test asserts that every option key listed
in the registry appears in the template constant. This catches drift at CI time
without building a template renderer, keeping the template human-readable and
easy to format.

**Alternatives considered:** Generating the template from the registry. Rejected
— grouping, ordering, and prose comments around options are fiddly to generate
and the output is rarely as readable as a hand-maintained file; a presence test
gets the safety for free.

### D8: Distinct exit codes; no bypass flag

Empty/missing feeds file → exit 2; no working feed on a fresh DB → exit 3.
Existing load/DB errors remain exit 1. Distinct codes let scripts and tests
distinguish refusal reasons. No `--skip-check`/`--force` flag is added: after the
first successful verification the `feeds`-table trust (D2) already makes offline
launches work, and on a truly fresh DB with no network there is nothing cached to
read, so refusing is correct.

**Alternatives considered:** A `--skip-check` flag for CI/scripted launches.
Deferred — not needed given D2; can be added later if a real use case appears.

### D9: Fixed 10s per-feed / 30s overall deadlines, not configurable

The first-run verification uses hardcoded 10-second per-feed and 30-second
overall deadlines. These bound worst-case startup latency for the one path that
actually blocks (fresh DB). Making them configurable would add a `[refresh]`
knob whose value is only relevant on first run; the fixed values keep the config
surface unchanged.

**Alternatives considered:** A configurable `startup_timeout` in `[refresh]`.
Deferred — YAGNI until experience shows the fixed values are wrong for someone.

## Risks / Trade-offs

- **[First launch blocks up to 30s on bad network]** → A true first run with no
  reachable feed waits out the overall deadline before refusing. Mitigation: the
  `verifying feeds...` stderr notice (printed before the fetch) tells the user
  something is happening; per-feed early completion means a single good feed
  resolves the gate fast.
- **[Trust persists even if all feeds later go stale]** → Once the `feeds` table
  is non-empty, the gate never re-verifies liveness at startup, so a user whose
  feeds all 404 will still boot into the TUI and discover failures only via the
  async refresh status. Mitigation: acceptable — the TUI still shows cached
  articles and surfaces refresh errors in the status bar; re-gating would break
  offline use, which is the whole point of D2.
- **[Self-update appends to a manually-curated file]** → Appending commented
  blocks to a user's hand-formatted `config.toml` may land in a surprising
  spot. Mitigation: appends go at end of file under a clear "new in vN" banner;
  comments are inert so imperfect placement cannot break parsing; the
  key-presence guard prevents repetition.
- **[Template/registry drift]** → A hand-maintained template can fall behind the
  registry. Mitigation: the D7 presence test fails CI if any registry key is
  missing from the template.
- **[Zero-height render guards may hide the status bar]** → When height is too
  small, prioritizing "no crash" over "show everything" means the status bar may
  be omitted. Mitigation: acceptable; the real `WindowSizeMsg` restores full
  rendering within one frame, and the degenerate case is transient.

## Migration Plan

No data migration is required — the `feeds` table already exists and is read
as-is. For existing users who already have a `config.toml` but an older schema
version (or no header), the self-update runs best-effort on the next launch and
appends any new options as comments; a file with no parseable schema-version
header is left untouched (the append is skipped rather than guessed). New users
get the full bootstrap. Rollback is simply reverting the code; no schema changes
need undoing.

## Open Questions

None — all decisions are locked from the exploration phase. The fixed timeout
values (D9) and the absence of a bypass flag (D8) are the most likely candidates
for revisit once real users hit first-run on flaky networks.

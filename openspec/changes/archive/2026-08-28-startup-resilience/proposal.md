## Why

`yerss` crashes on a fresh start: `renderList` panics with `index out of range
[-1]` because `View()` is invoked before the first `WindowSizeMsg` arrives, so
`m.height` is still 0 and the line-truncation step empties its slice. Beyond the
crash, a first-time user gets no config directory, no `config.toml`, and no
`feeds.txt` to edit — nothing is ever materialized — and the app will happily
launch into a full-screen TUI with zero feeds, discovering the empty state only
after taking over the terminal. Startup should be crash-proof, self-bootstrapping,
and refuse to enter the TUI until at least one RSS feed has been verified live.

## What Changes

- **Render safety:** `renderList` (and the startup path generally) SHALL NOT
  panic when terminal dimensions are zero or smaller than the content; the model
  SHALL seed sane default dimensions so the first frame renders before
  `WindowSizeMsg` is processed.
- **First-run bootstrap:** On first run, the system SHALL create the config
  directory and seed a full commented `config.toml` (with a schema-version and
  build-commit header) plus an empty `feeds.txt` with a comment header, so the
  "refuse to start" message has a concrete file to point at.
- **Config self-update:** A `releases.go` registry SHALL track the config schema
  version and a running tab of config options added per version. On later runs,
  the system SHALL append newly-introduced options to an existing `config.toml`
  as comments (best-effort, idempotent, keyed on option presence) and bump the
  file's schema-version header.
- **Startup feed gate (pre-TUI):** Before launching the TUI, the system SHALL
  refuse to start when `feeds.txt` is empty or missing (exit code 2). When the
  feeds table already contains a verified feed, startup SHALL proceed without a
  network check (offline-safe). Otherwise (first run or wiped DB) the system
  SHALL fetch the configured feeds with a 10s per-feed / 30s overall deadline
  and refuse to start (exit code 3) if no feed parses successfully.
- **Tests:** Add render edge-case tests (zero/tiny dimensions, empty list,
  `listWindow` cases, truncation), config bootstrap/self-update tests, and
  startup-gate tests using the existing `withXDG` temp-dir helper.

## Capabilities

### New Capabilities

<!-- None — all three affected capabilities already exist. -->

### Modified Capabilities

- `configuration`: Adds first-run bootstrap (create config dir, seed `config.toml`
  + `feeds.txt`) and config-file self-update via a schema-version registry in
  `releases.go`.
- `feed-pipeline`: Adds a pre-TUI startup feed-verification gate with distinct
  refusal exit codes and an offline-safe trust rule based on the existing
  `feeds` table.
- `reader-ui`: Adds a render-safety invariant so list/article rendering never
  panics under zero or degenerate terminal dimensions, and seeds default model
  dimensions before the first `View()`.

## Impact

- **Code:** `internal/ui` (renderList guard, default dimensions in `New`),
  `internal/config` (bootstrap, `releases.go` registry, self-update on load),
  `internal/feed` (verify/gate helper), `cmd/yerss/main.go` (pre-TUI gate
  ordering, new exit codes 2 and 3, "verifying feeds..." stderr notice).
- **APIs:** No public API changes. New internal helpers (`config.Bootstrap`,
  `config.MigrateFile`, `feed.VerifyFeeds` or equivalent).
- **Dependencies:** None added. Reuses `gofeed`, `xdg`, and the existing store.
- **Filesystem:** Now creates `$XDG_CONFIG_HOME/yerss/` (previously never
  created) and seeds `config.toml` + `feeds.txt` there; may append comments to
  an existing `config.toml` on upgrade. Existing files are never overwritten or
  destructively modified.
- **Behavior:** First launch now requires network once (to verify ≥1 live feed);
  subsequent launches work offline. Previously the app started unconditionally
  and surfaced empty-feed state only inside the TUI.

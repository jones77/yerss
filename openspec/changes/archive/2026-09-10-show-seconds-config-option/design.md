## Context

See `proposal.md` — Why. Current state:

- Article-view top-border date uses `timeutil.LayoutDateTimeMinute`
  (`2006-01-02 15:04`) after the quick fix; article rows and the list status
  bar use `timeutil.LayoutTime` (`15:04`). `timeutil.LayoutDateTime`
  (`2006-01-02 15:04:05`) already exists and is currently unused.
- Display options (`padding_x`, `padding_y`, `scrollbar`, `theme`, `images`,
  `ascii`) live in `config.Display` (`internal/config/config.go`), read from
  the `[display]` TOML section, seeded into the first-run template
  (`config/bootstrap.go`), and tracked by the self-update registry
  (`config/releases.go`).

## Goals / Non-Goals

**Goals:**
- A `show_seconds` display option that toggles `HH:MM` ↔ `HH:MM:SS`
  consistently across every displayed time in the list and article views.
- Default off, matching the new minute-precision rendering.

**Non-Goals:**
- Not changing undated fallbacks (`--:--`), the day-group date labels, or the
  JSON export formats.

## Decisions

**1. `ShowSeconds bool` on `config.Display`, default `false`, TOML key
`show_seconds`.**
Mirrors the existing `scrollbar`/`padding` display options exactly: a field on
`Display`, a `*bool` in the raw TOML parsing with a false default, a
commented-out line in `seededDisplay`/`bootstrap.go`, and a config-release
entry in `releases.go` so existing configs self-update. Invalid values fall
back to `false` (same pattern as other display options).

**2. Layout selection helpers in `timeutil`.**
Add `LayoutTimeSeconds = "15:04:05"` and two tiny helpers so the three call
sites cannot drift:
`TimeLayout(showSeconds bool) string` → `15:04:05` or `15:04`, and
`DateTimeLayout(showSeconds bool) string` → `2006-01-02 15:04:05` or
`2006-01-02 15:04`. Call sites pass `m.sess.Config().Display.ShowSeconds`.

**3. Include the list status-bar "last refresh" time.**
Seconds should be consistent across the whole list view, not just article rows;
the status-bar refresh time is the third display site and uses the same helper.

## Risks / Trade-offs

- Widening the article top-border date prefix by three characters could push
  the title truncation sooner when seconds are enabled — a visible, expected
  trade-off of the option, not a defect. The truncation requirement already
  keeps the date always fully visible.
- Adding a config option carries bootstrap/self-update surface → mitigated by
  following the existing `scrollbar` pattern exactly and by the config
  self-update test that regenerates the seeded template from the release
  registry.

## Migration Plan

Config only; the default (`false`) preserves current behavior. Existing configs
gain a commented `# show_seconds = false` line on next run via the release
registry.

## Open Questions

None.
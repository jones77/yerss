## Why

Timestamps everywhere show seconds, which are noise for a reader. The article
view already dropped them in a quick fix; this change makes seconds a
configurable display option so both the list view and the article view can
show `HH:MM` by default and `HH:MM:SS` on request.

## What Changes

- New `show_seconds` boolean under the `[display]` section of the TOML config,
  defaulting to `false`.
- When `false` (default): the article row time renders as `HH:MM` and the
  article top-border date prefix renders as `YYYY-MM-DD HH:MM`.
- When `true`: the article row time renders as `HH:MM:SS` and the article
  top-border date prefix renders as `YYYY-MM-DD HH:MM:SS`.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `configuration`: add the `show_seconds` display option requirement alongside
  the existing `scrollbar` and padding options.
- `reader-ui`: the article row time (`HH:MM` today) and the article top-border
  date prefix (`YYYY-MM-DD HH:MM:SS` today) requirements change to follow the
  `show_seconds` option.

## Impact

- `internal/config/config.go`: add `ShowSeconds bool` to `Display`, default
  `false`, with a `show_seconds` TOML key and a config-release entry so the
  seeded config template and self-update registry stay in sync.
- `internal/timeutil/timeutil.go`: add `LayoutTimeSeconds = "15:04:05"`; the
  existing `LayoutTime` (`15:04`), `LayoutDateTime` (`… 15:04:05`), and
  `LayoutDateTimeMinute` (`… 15:04`) layouts are the `false`/`true` sides.
- `internal/ui/list.go`: article-row and "last refresh" times pick the layout
  from the option.
- `internal/ui/article.go`: the top-border date format picks `LayoutDateTime`
  or `LayoutDateTimeMinute` from the option.
- `internal/config/releases.go` and `internal/config/bootstrap.go`: the seeded
  template and self-update registry gain the `show_seconds` key.
- Tests for the config defaults, list-row time, and article top-border date.
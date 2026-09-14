## Why

Dates and times are already converted to the terminal's local timezone throughout
the app — the article list day headers, the article rows, and the article's top
border all use `.Local()` — but the usage documentation never says so, leaving
readers unsure which timezone the dates use.

## What Changes

- Document in `README.md` that publication dates and times are displayed in the
  user's local timezone, and that there is no in-app timezone setting: the app
  follows the system/terminal local time (the `TZ` environment).

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

<!-- None — documentation only; no yerss behavior changes. -->

This change opts out of specs (`skip_specs: true` in `.openspec.yaml`): it adds
user-facing documentation, not application behavior.

## Impact

- **Code:** None.
- **Files:** `README.md` gains a timezone note.
- **APIs:** None.
- **Dependencies:** None.
- **Behavior:** None.

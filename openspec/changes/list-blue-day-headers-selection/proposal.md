## Why

The list view's day-group headers render in the grey/dim role and use a bare
rail corner followed by a single space (`├ <date>`), which reads as ambiguous
with the article rows below. The headers should carry a visual dash connector
(`├─── <date>`) and render in the same blue role as the status bar so they read
as chrome. Separately, the selection highlight is a fixed `#707070` grey that
sits oddly against the dim-grey text in the same view; it should be a
theme-aware highlight: light `#d3d3d3` on light backgrounds and dark `#242424`
on dark backgrounds, keeping selected text readable in both modes.

## What Changes

- Day-group headers in the list view render as `{corner}──── {date}` (four
  horizontal dashes between the rail corner and the date, ASCII `+----`),
  reusing the existing `H` dash glyph so the Unicode/ASCII pair stays in sync.
- The entire day-header line (corner, dashes, and date) renders in the blue
  role (ANSI 12 bright blue in dark mode, ANSI 4 blue in light mode) — the same
  role as the status bar / help hint — instead of the grey/dim role.
- A selected day header renders blue on the selection background via a new
  `selStatus` precomputed style.
- The selection highlight becomes theme-aware: `#242424` in the dark palette
  and `#d3d3d3` in the light palette, replacing the fixed `#707070`. Selected
  text foregrounds are unchanged (white on `#242424`, black on `#d3d3d3` both
  remain readable).
- The `Dim` role, article-row layout, fold markers, and selection semantics are
  unchanged.

## Capabilities

### New Capabilities

### Modified Capabilities

- `reader-ui`: The "Article list rendering" requirement changes so day headers
  render a three-dash connector after the rail corner and render in the blue
  role; the "Standard terminal color roles" requirement changes so the
  selection highlight is theme-specific (`#242424` dark / `#d3d3d3` light)
  instead of a fixed `#707070`, and day headers move from the grey role to the
  blue role.

## Impact

- `internal/ui/render/theme.go` — palette `Selection` values and doc comment.
- `internal/ui/model.go` — new `selStatus` style in `modelStyles`/
  `buildModelStyles`.
- `internal/ui/list.go` — `renderDayHeader` format and role.
- `internal/ui/daygroups_test.go` — rail prefix assertions (`┌─── `, `+--- `).
- `internal/ui/polish_test.go`, `internal/ui/tags_test.go` — hardcoded
  `#707070` selection expectations become `#242424`.
- Spec: `openspec/specs/reader-ui/spec.md` (delta).
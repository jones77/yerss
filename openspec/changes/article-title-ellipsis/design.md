# Design: article-title-ellipsis

## Current top border

```go
func topBorder(w, padX int, g borderGlyphs, p Palette, date, title string) string {
    text := " " + date + " · " + title + " "
    text = truncate(text, w-2*padX-3)          // hard cut, no ellipsis
    fill := (w-2*padX-2) - 1 - runewidth.StringWidth(text)
    return padX-spaces + g.tl + g.h + text + h*fill + g.tr
}
```

When the title fills the budget, `fill == 0` and the title abuts `g.tr` with no
dash — the asymmetry.

## Glyph

Add an `ellipsis` field to `borderGlyphs`, set in `glyphsFor`:

```go
unicode: { ..., ellipsis: "…" }
ascii:   { ..., ellipsis: "..." }
```

This pair is also the example the ascii-flag change's AGENTS.md note will
reference for the "keep both modes in sync" rule.

## New layout

Frame the top border as two rails around a core:

```
leftRail  = g.tl + g.h          // "┌─"   corner + one dash
rightRail = g.h + g.tr          // "─╖"   one dash + corner
core      = " " + date + " · " + title + " "   // spaces around the content
```

The full line is `leftRail + core + fill + rightRail` and must total `w`.

```
innerW   = w - 2                       // columns between the two corners
coreW    = innerW - 2                  // minus the left dash and right dash
prefix   = " " + date + " · "
suffix   = " "
titleW   = coreW - runewidth.StringWidth(prefix) - runewidth.StringWidth(suffix)

if runewidth.StringWidth(title) <= titleW:
    // fits: fill the leftover with dashes, no ellipsis
    core = prefix + title + suffix
    fill = coreW - runewidth.StringWidth(core)        // >= 0
    line = leftRail + core + h*fill + rightRail       // rightRail dash counts as the trailing dash
else:
    // truncated: ellipsis + exactly the rightRail dash, no fill
    el = g.ellipsis
    cut = truncate(title, titleW - runewidth.StringWidth(el))
    core = prefix + cut + el + suffix
    line = leftRail + core + rightRail                // no fill dashes
```

`truncate` already cuts by display width, so CJK titles are handled. The
`runewidth.StringWidth(el)` is 1 for `…` and 3 for `...`, so the ASCII fallback
reserves the right amount of room.

## Edge cases

- **No date** (empty): `prefix = "  · "` is ugly; when `date == ""`, build
  `core = " " + title + " "` (drop the `· ` separator) so an untitled/undated
  article still renders cleanly.
- **Title exactly fills `titleW`:** fits branch, `fill == 0`, line is
  `leftRail + core + rightRail` — still symmetric (one dash each side). Better
  than today.
- **Very long title, tiny terminal:** `titleW` can go non-positive after
  subtracting prefix + ellipsis; clamp `titleW` to a minimum so at least the
  ellipsis shows, and let the date overflow protection come from the outer
  `renderArticleBorder` clamp (already clamps lines to `h`).

## Tests

- Title fits → no ellipsis, fill dashes present, width == w.
- Title too long → ends with `…`, exactly one dash before right corner, date
  intact, width == w.
- ASCII mode → ends with `...`.
- CJK title truncation respects display width.
- Empty date → no dangling `· `.

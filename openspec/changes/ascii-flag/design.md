# Design: ascii-flag

## Flag wiring

`cmd/yerss/main.go` already uses pflag for `-e`/`-c`. Add the bool alongside:

```go
var ascii bool
pflag.BoolVarP(&ascii, "ascii", "a", false, "force ASCII fallback glyphs")
```

The flag is a normal runtime flag (unlike `-e`/`-c`, it does not early-exit), so
it is parsed and then applied after the model is built.

## Applying the override

`ui.New` derives `m.ascii` from `cfg.Display.Ascii || detectAsciiNeeded()`.
Render uses `glyphsFor(m.ascii)` each frame, so flipping `m.ascii` after
construction takes effect immediately. Add a small setter:

```go
// internal/ui/model.go
func (m *Model) SetAscii(v bool) { m.ascii = v }
```

```go
// cmd/yerss/main.go
m := ui.New(cfg, st)
if ascii {
    m.SetAscii(true)   // override config + detection
}
```

No change to `glyphsFor`, `detectAsciiNeeded`, or the config loader.

## AGENTS.md note

Add a short subsection under the existing `## Project` section so contributors
(and agents) see it next to the project description:

> ### Unicode/ASCII glyph pairs
>
> When you add a Unicode glyph (box-drawing, `…`, fold markers, etc.), add its
> ASCII fallback in `glyphsFor` (`internal/ui/border.go`) and keep the two in
> sync. The `-a`/`--ascii` flag forces the fallback so both paths can be
> exercised in testing.

## Help text

pflag generates `--help` automatically; the new flag appears with its
description. No manual help formatting needed.

## Tests

- `cmd/yerss`: `-a` sets the model's ascii to true even when config ascii is
  false and detection returns false (drive via a small test harness or by
  calling the same wiring function).
- `Model.SetAscii(true)` makes `renderArticle` use ASCII glyphs.
- Existing ASCII-auto-detect and config-ascii tests remain green.

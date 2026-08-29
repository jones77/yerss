# Add -a/--ascii flag to force ASCII fallback

## Why

ASCII fallback is currently only reachable via the `ascii` config option or
terminal auto-detection. For testing and for terminals that misreport, there is
no quick command-line way to force the fallback to exercise the ASCII glyph path
without editing `config.toml`.

## What

- Add a `-a` / `--ascii` command-line flag (pflag) that forces ASCII fallback
  mode, overriding both the config setting and terminal detection.
- Add an `AGENTS.md` note: when introducing a Unicode glyph (box-drawing,
  ellipsis, fold markers, etc.), add its ASCII fallback in `glyphsFor` and keep
  the two in sync, since `-a` exercises the fallback path.

## Non-goals

- Adding new glyphs (each feature adds its own pair).
- Changing the config option or the terminal auto-detection logic.

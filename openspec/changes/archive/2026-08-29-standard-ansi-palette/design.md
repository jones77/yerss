## Context

See `proposal.md` — Why. The TUI currently ships bespoke hex palettes
(`internal/ui/theme.go`) under arbitrary role names (`Accent`, `Dim`, `Bold`,
`Border`, `StatusBar`), plus a hardcoded `#333333` selection background in the
list view and glamour's own base color for article body text. All four standard
colors chosen by the user map onto these existing roles with some merging.

## Goals / Non-Goals

**Goals:**
- All UI foreground colors come from the standard ANSI palette (0-15).
- The palette role names and code comments match the standard terminology.
- Light mode stays selectable (`auto`/`dark`/`light`) with readable ANSI values.

**Non-Goals:**
- No new theme configuration, no user-facing color settings.
- The `#333333` selection background is intentionally unchanged (a background,
  outside ANSI foreground numbering).
- Glyph/ASCII fallback behavior is untouched.

## Decisions

### 1. Collapse five palette roles into four standard roles

Current struct has five fields that map onto four distinct colors:

```
Accent (pink) ─┐
StatusBar (blue)┼─→ blue role   (12 dark / 4 light)
Border (grey) ─┐
Dim (grey-blue)┼─→ grey role    (8 / 8)
Bold (light)   ──→ bright role  (15 / 0 bold)
(none)         ──→ text role    (7 / 0)
```

The `Palette` struct SHALL be reduced to four fields named for the roles they
play, with `Accent` and `Border` removed:

```go
type Palette struct {
    Text      lipgloss.Color // 7 white / 0 black       — article body text
    Bright    lipgloss.Color // 15 bright white / 0 bold — unread titles, bold counts
    Dim       lipgloss.Color // 8 dark grey / 8         — border, rails, muted text
    StatusBar lipgloss.Color // 12 bright blue / 4 blue — status bar, inline text, popups
}
```

- `Accent` call sites (popup titles/borders) move to `StatusBar`.
- `Border` call sites (article frame, rails, bullets, scrollbar track) move to
  `Dim`.
- `Bold` renames to `Bright`.
- Alternative considered: keep the five field names and simply re-point their
  values. Rejected: it leaves misleading names (e.g. two names for the same
  blue) that contradict the "standard names" goal, and the merge is cheap and
  mechanical.

### 2. List-view tree rail adopts the grey role

The article-row tree rail (`corner + "-"` in `renderArticleRow`) is currently
rendered with no foreground color (default terminal fg), while day-header
corners are already dim. For consistency with the article border ("rails are
grey"), the article-row rail SHALL render in the grey role (`Dim`, ANSI 8).
Alternative considered: leave the rail default-colored. Rejected: it is a rail
and would stand out against the now-grey chrome.

### 3. Body text via glamour base-color override

Article body text is rendered by glamour, not the palette. `glamourStyleConfig`
(markdown.go:35) SHALL override the base document foreground to the text role
(`"7"` in dark mode, `"0"` in light mode) after selecting the dark/light style
config. Styled elements (headings, bold, italics, links, blockquotes) keep
glamour's own theme colors. `"7"`/`"0"` are valid lipgloss ANSI color strings,
which glamour resolves through lipgloss. This must be verified against the
actual `ansi.StyleConfig.Document.Color` field shape during implementation.

### 4. Selection background stays `#333333`

`list.go:279,286` keeps its hardcoded background. It is the one non-ANSI color
remaining; a mid-dark grey with no standard ANSI equivalent (ANSI 8 is too
light as a background bar in practice). Documented as an intentional exception.

### 5. Light palette maps to dark-variant ANSI values

```
        dark   light
Text    7      0
Bright  15     0 + bold
Dim     8      8
Status  12     4
```

Bright variants (12, 15) wash out on light backgrounds, so light mode uses the
standard dark variants (4, 0), with bold carrying the unread emphasis. `Dim`
(8) reads on both backgrounds, so it is shared. `resolvePalette` and
`detectLightBackground` are unchanged.

## Risks / Trade-offs

- [ANSI 8 can render as near-black on terminals that do not remap bright
  variants, making muted text hard to read] → Standard ANSI is the point of the
  change; the user's terminal theme controls the rendering. `#333333` selection
  still provides contrast.
- [glamour base-color override may not take on the `Document.Color` field, or
  may affect more than base text] → Verify with a focused test in the first
  task; fall back to setting the color that glamour's base body style already
  uses in each theme, pinned to the standard value.
- [Removing `Accent`/`Border` fields touches ~19 call sites] → Mechanical
  rename; `go build` and the existing polish tests catch stragglers.
- [Bright white (15) + bold is visually redundant on some terminals] → Harmless;
  keeps the light-mode "bold black" behavior uniform.

## Migration Plan

No data or config migration: `display.theme` values (`auto`/`dark`/`light`)
are unchanged and existing user configs keep working. The change is code-only;
rollback is a revert of `internal/ui/*` plus the `reader-ui` spec.

## Open Questions

None — the light-mode values, the tree-rail color, and the popup chrome color
were all resolved with the user during exploration.
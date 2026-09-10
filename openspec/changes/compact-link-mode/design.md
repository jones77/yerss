## Context

See `proposal.md` — Why. The relevant current state:

- Body links render through glamour v2, which always emits `text url` for
  regular links: `LinkElement.Render` writes the name span (OSC 8 + name, styled
  with `Styles.LinkText`) then the href span (` ` + OSC 8 + URL, styled with
  `Styles.Link`). `SkipHref` is hardcoded to `isFooterLinks` (table-only), and
  `renderText` ignores the `Conceal` style field, so there is no config knob.
- The app already hit-tests OSC 8 spans for left-click: `parseLinkSpans`
  (`article_selection.go:127`) extracts per-line `(start, end, url)` spans and
  `linkAtContentCell` maps a click cell to a URL; `updateArticleMouse` opens it.
  Clickability derives from the OSC 8 sequences, not the visible URL text.
- The article body is derived once in `computeDerivation` and cached in
  `derivCache` keyed by `derivationKey` (id, contentW, vpH, ascii, palette,
  style, content). The header URL is hand-rendered (`article.go:524`) and
  currently bypasses glamour, so it does not wrap.
- Keybindings come from the catalog in `config/keys.go`; `x` is bound to
  `ExpandToggle` in `ViewList` only. The help popup and seeded config derive
  from the catalog automatically.

## Goals / Non-Goals

**Goals:**
- Compact body links (name only, blue) that remain clickable via both
  Cmd/Ctrl-click and left-click, without waiting on an upstream glamour release.
- An `x` toggle to expanded mode (`text url`), session-persistent, default
  compact.
- Header URL wraps at the content width and renders in the blue role.

**Non-Goals:**
- Not changing glamour's upstream behavior in this change (noted as a future
  simplification, not a dependency).
- Not altering the links popup, text selection, or URL copy behavior.

## Decisions

**1. Suppress the href span by post-processing the rendered ANSI, not forking glamour.**
In compact mode, walk each rendered body segment with the existing
`compose.ScanANSI` machinery and drop the href span of each link. The emitted
structure is always `name-span` then ` ` then `href-span` with the same OSC 8
target, so the rule is: keep the first span for a target, drop any immediately
following span carrying the same target (and its leading space). The name span
keeps its OSC 8, so clickability survives unchanged.
- Alternative (rejected for now): an upstream glamour patch exposing `SkipHref`
  as `HideLinkURL`. Cleaner long-term, but blocks this change on a third-party
  release and adds a `replace`/vendor dependency. Post-processing is self-
  contained and unit-testable against `parseLinkSpans`' invariants.

**2. Body link color via the glamour style config.**
Set `cfg.LinkText.Color` and `cfg.Link.Color` in `GlamourStyleConfig` to the
blue role (ANSI 12 in dark, 4 in light), mirroring the palette's `StatusBar`.
`GlamourStyleConfig` already branches on the resolved dark/light style, so it
has what it needs. Both spans render blue in both modes; the href color matches
the name so expanded mode reads as one link unit.

**3. Toggle state is a model field that keys the derivation cache.**
Add `linkExpand bool` on the `Model`, defaulting false (compact). It joins
`derivationKey`, so toggling invalidates the cached derivation and recomposes
the open article at the same scroll offset. Post-processing runs inside
`computeDerivation` on the body segments (not inside `renderMarkdown`, which
also renders titles and popup text). The field persists across article switches
and is deliberately not persisted to disk.

**4. Header URL wraps by hand, one OSC 8 span per wrapped line.**
Replace the single raw URL string with width-wrapped lines of the stripped URL,
each line emitted as `ansi.SetHyperlink(fullURL) + line + ansi.ResetHyperlink()`
rendered through `m.styles.status`. One span per line avoids terminals that
drop hyperlinks across a line break. Reuse `image.LayoutWrapLines` for the wrap
(it hard-splits over-long words, correct for URLs).

**5. New `LinkExpandToggle` action bound to `x` in the article view.**
Adding an action to the catalog (`config/keys.go`) makes the help popup and the
seeded config template pick it up with no further wiring, and lets users rebind
it. A dedicated action gives clearer help than reusing `ExpandToggle`; reusing
that action (adding `ViewArticle` to its views) was considered and rejected
because its help label is list-specific ("toggle all day groups").
The action is a **default-on catalog binding, not a config opt-in**: like the
list view's expand toggle, `x` in the article view works out of the box with no
explicit config file line. Users who dislike it can rebind or unbind it through
the existing keybindings mechanism; there is no separate feature flag.

## Risks / Trade-offs

- Post-processing depends on glamour's exact span emission (`name` then `href`,
  same target). If glamour changes it, compact mode could mis-strip → mitigated
  by a focused test suite on the stripper asserting the `parseLinkSpans` output
  of compact-rendered lines (name span only, same URL), plus the upstream patch
  as an escape hatch.
- The derivation cache silently returning the wrong mode if the toggle is
  omitted from `derivationKey` → the key is updated in the same commit as the
  field; a test toggles and asserts a recompose happened.
- A link whose name text equals its URL (`[https://x](https://x)`) has two
  same-target spans both resembling URLs; the keep-first-drop-next rule still
  yields the name and drops the duplicate, so compact renders it once.
- Two distinct links to the same URL on one line: each pair is name-then-href,
  so per-target keep-first/drop-next handles both correctly.

## Migration Plan

In-repo change; no data migration. The list view's `x` binding is untouched.
Rollback is reverting the commit — no persistent state changes.

## Open Questions

- **Hover tooltip for link URLs** (deferred; see `proposal.md` — What Changes).
  If pursued: the app already runs `tea.WithMouseCellMotion()` and the article
  view handles `MouseActionMotion` for drag selection, so a transient overlay
  in `View` keyed on the hovered cell (hit-testing `linkAtContentCell`) can show
  the link's URL without touching the compact/expanded rendering or the
  derivation cache. Would need throttling (motion events fire per cell), frame
  clamping near the borders, and a decision on tooltip text (full vs stripped
  URL). Safe to add after this change — it does not alter the specs.
- Whether expanded mode should remember per-feed or per-article could be added
  later without touching the specs.
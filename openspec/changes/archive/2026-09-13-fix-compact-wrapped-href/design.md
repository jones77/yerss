## Context

`compose.StripLinkHrefs` post-processes rendered ANSI to remove the visible href
glamour appends to each body link (see proposal.md - Why). Glamour renders a
link as a name span followed by an href span carrying the same OSC 8 target; the
current logic counts per-target OSC 8 spans and drops each even occurrence,
assuming exactly one name span and one href span per link.

That assumption breaks under wrapping. Glamour's paragraph renderer calls
`lipgloss.Wrap`, whose `WrapWriter` re-emits `SetHyperlink` after every wrapped
newline. A link part (name or href) that wraps across N lines therefore becomes
N OSC 8 spans with the same target. The odd/even count then drops the wrong
spans: for a two-line href the name, first href line, and second href line are
the 1st, 2nd, and 3rd occurrences, so the second href line is treated as a
"name" and kept; for a wrapped name the second name line is treated as an "href"
and dropped.

## Goals / Non-Goals

**Goals:**

- In compact mode, drop the full visible href of each body link — the leading
  space plus every wrapped continuation line — regardless of how the URL wraps.
- Preserve the link name in full, including any wrapped name continuation lines.
- Keep the OSC 8 name span (full URL target) so the link remains clickable.

**Non-Goals:**

- Changing expanded mode, which shows the full URL and is already correct.
- Altering the `x` toggle, recompose, or viewport machinery.
- Patching glamour to expose `SkipHref` for body links.

## Decisions

- **Group same-target spans into logical units, then alternate keep/drop per
  unit.** A wrapped link part is a maximal run of same-target OSC 8 spans where
  each subsequent span is separated from the previous span's close only by a
  newline and style escapes (a mid-word wrap). A space (glamour's href prefix, or
  the inter-word text between links) ends a unit. The units for one target then
  alternate: the 1st is the name (keep), the 2nd the href (drop), the 3rd the
  next link's name (keep), and so on.
  - Rationale: glamour emits `name href name href …` in strict order, one name and
    one href per link; wrapping only multiplies the spans inside a unit, never
    changes the unit order. This correctly handles a wrapped href, a wrapped name,
    and adjacent links to the same URL.
- **Drop a href unit as a whole.** The drop region starts at the href unit's first
  span, extended back over its leading styled space (glamour's separator,
  unchanged from today), and ends at the last span of the unit. For a href pushed
  to the next line without an internal wrap (leading space on the prior line), the
  drop starts at the href's OSC 8 open, matching existing behavior.

### Alternatives considered

- **Upstream `SkipHref`:** glamour's `LinkElement.SkipHref` is hardcoded to table
  footer links; no public option forces it for body links. Rejected (fork/patch
  not warranted).
- **Strip at the markdown level:** rewriting `[text](url)` before glamour cannot
  keep the OSC 8 target while hiding the URL (an empty destination produces no
  hyperlink). Rejected — it would break clickability.

## Risks / Trade-offs

- [Over/under-drop across adjacent same-URL links] → The unit alternation plus a
  test with two links to the same URL guards against dropping a name or keeping a
  href.
- [Wrap-boundary edge cases] → The grouping key is "no space between spans",
  which holds for any mid-URL break because URLs contain no spaces. Tests use URLs
  long enough to wrap at least once.

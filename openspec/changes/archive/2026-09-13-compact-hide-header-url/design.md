## Context

`renderHeader` builds the reader header from the URL line, the bold title, and
the author line, and `renderHeaderLink` emits the URL line wrapped at the content
width with an OSC 8 hyperlink. `computeDerivation` calls `renderHeader`
unconditionally, so the URL line renders in both link modes. `linkExpand` is
already part of `derivationKey`, so toggling it recomputes the derivation — no
cache change is needed for this behavior.

## Goals / Non-Goals

**Goals:**

- Hide the header URL line in compact mode so the title becomes the first content
  line.
- Keep expanded mode rendering the header URL exactly as today.

**Non-Goals:**

- Changing title/author rendering.
- Changing body-link handling (that stays in `StripLinkHrefs`).
- Adding a new keybinding or mode.

## Decisions

- **Gate the header URL on `linkExpand` at render time.** Pass the link mode into
  `renderHeader` so it omits `renderHeaderLink` when compact; when `a.Link == ""`
  the existing early-return still applies.
  - Rationale: the header URL is not a body link — it has no "name" — so the
    compact rule is "omit the URL line" rather than "strip the href".
- **Recompute happens automatically.** Because `linkExpand` is in `derivationKey`,
  toggling `x` already re-derives the header; `headerLines` is recomputed from the
  shortened header so viewport composition and reading-position preservation keep
  working unchanged.

### Alternatives considered

- **Strip the URL text instead of omitting the line:** the header has only a URL,
  no link name, so there is nothing meaningful to leave; omitting the whole line
  is the correct compact behavior.
- **A separate header-only toggle:** unnecessary — the header should follow the
  same compact/expanded mode as body links.

## Risks / Trade-offs

- [Reading-position shift on toggle] → Toggling changes the header height; the
  existing `recomposeArticle` offset+inserted+snap already handles reflows, so no
  new machinery is required.
- [Open URL / Copy URL still work in compact mode] → These use
  `m.article.article.Link` directly, not the rendered header, so hiding the line
  does not affect them.

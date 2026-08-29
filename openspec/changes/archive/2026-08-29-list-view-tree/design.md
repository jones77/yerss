## Context

`renderArticleRow` (internal/ui/list.go) currently renders a two-space cursor
gutter, the title, then `· source · time` right-aligned, dropping the gap when
the title is truncated. `renderDayHeader` renders a fold marker (`▾`/`▸`) plus
the long date. `bucketDayGroups` builds labels from `longDate` with no notion
of "now". `glyphsFor` (internal/ui/border.go) has no tee glyph. The status
bar's `selectedArticlePosition` and all cursor/wrap/page logic operate on the
flattened visible-row list and are unaffected by styling changes.

## Goals / Non-Goals

**Goals:**

- Tree-rail rendering with bookend glyphs computed against the whole
  visible-row list.
- Row layout: `HH:MM title` left, org right, guaranteed one-space gap on
  truncation.
- `today, `/`yesterday, ` label prefixes while keeping `longDate` untouched.
- Full-row background selection for all row kinds.

**Non-Goals:**

- No change to fold behavior, keybindings, cursor/page movement, or the status
  bar.
- No change to day grouping/bucketing logic — only labels.
- No redesign of the empty-state message.

## Decisions

**Rail glyphs computed against the whole visible-row list, not the scroll
window.** `┌` is the prefix of visible row 0 and `└` of the last row
regardless of what the window shows. Alternative: re-bookending every window
so `┌`/`└` are always on screen — rejected because the glyphs would jump as
the cursor moves, and mid-list `├`-only screens are tree-correct. Collapsed
groups simply remove rows; the rail recomputes from the flattened list.

**Glyph mapping reuses existing fields; one new field.** `┌`/`└` map to
`borderGlyphs.tl`/`bl` (ASCII `+`), `─` to `h` (ASCII `-`); a new `tee` field
carries `├` (ASCII `+`). ASCII and Unicode stay in sync in `glyphsFor` per
AGENTS.md, exercised by the `-a` flag.

**Labels via an injected reference time.** `bucketDayGroups(arts []store.Article, now time.Time)`
compares each group's local midnight against `dayStart(now)` and
`dayStart(now.AddDate(0,0,-1))`, prefixing `today, `/`yesterday, ` to the
unchanged `longDate` output. Callers pass `time.Now()`; tests pass a fixed
reference so `TestBucketDayGroupsOrdering` stays deterministic. `Undated`
keeps its label. Prefix comparison uses local midnights, so a day whose
articles straddle midnight in another zone still buckets by local day (existing
behavior).

**Row layout: time joins the left block, org becomes the sole right field.**
Prefix is the rail (`├─ ` = three columns on article rows, `├ ` = two on
headers) plus `HH:MM ` and the title. The right block is the dim source
identifier only. Available title width = `m.width - prefix - orgW`; when the
title exceeds it, truncate with `ansi.Truncate(title, availW-1, g.ellipsis)`
and always render one space before the org (today's gap-dropping behavior is
removed). Unread/read styling and the dim right block carry over unchanged.

**Selection via the existing background style.** Day headers already render a
`#333333` background when selected; that style extends to article rows,
covering the rail columns, and the `> `/two-space gutter is deleted. No new
palette entry.

## Risks / Trade-offs

- [Older days lose month/year context in the sketch's short form] → Not taken:
  labels keep the full long date, only prefixed — no information is lost.
- [Truncation gap interacts with the org width at narrow terminals] → The
  layout clamps: title width floors at 1 column and the whole row is truncated
  to terminal width as today, so degenerate widths render (possibly ugly)
  without panicking.
- [Prefix shifts row-width math used by click hit-testing] → Hit-testing maps
  rows by Y only (`handleListClick`), so column changes are safe.

## Migration Plan

No data migration; pure rendering change. Rollback is a plain revert.

## Open Questions

(none — settled during exploration)

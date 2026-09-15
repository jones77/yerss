## Context

The article reader renders markdown through glamour into styled lines, then
lays out image blocks and composes the viewport. Two features (expanded link
mode, and per-article cached recomposition) already exist; see
proposal.md - Why. The constraints that shape this design:

- Rendering happens entirely on the UI goroutine; the cached derivation is
  reused so image-load messages never re-run markdown conversion.
- OSC 8 hyperlinks and SGR styling are carried in the rendered text; any
  re-flow must preserve both per output line.
- The restore path composes the article once at a hardcoded size before the
  first real `WindowSizeMsg` arrives, and image blocks render as placeholders
  until their stored blocks match the content width.
- Anchor helpers (`AnchorableLine`, `FindAnchorLine`) already exist in
  `internal/ui/compose/anchor.go` and are used by the link-mode toggle.

## Goals / Non-Goals

**Goals:**
- Re-flow a pushed-out expanded-mode URL so it starts on the link name's line
  and wraps within the content width, keeping every wrapped line clickable.
- Hide a wrapped URL in compact mode without leaving a whitespace-only
  residual line that splits the paragraph.
- Persist the reading position as a content anchor that survives restart even
  when block geometry differs, and re-apply it once the real terminal size is
  known.
- Preserve the reading position across image-load recomposition, counting only
  lines inserted above the viewport top.

**Non-Goals:**
- No new user-facing settings or keybindings.
- No changes to how links are harvested or how stored image blocks render.
- No changes to the list-view restore path beyond what already exists.

## Decisions

### D1: Re-flow expanded URLs with a dedicated post-render pass

`InlineLinkURLs(rendered, contentW)` operates on the glamour-rendered text,
not the markdown source. It flattens each affected paragraph into a flow of
styled characters (each carrying its OSC 8 URL, href, and SGR), then
greedily re-wraps at `contentW`, breaking after spaces and URL breakpoints
(`/-._`), re-emitting OSC 8 + SGR per output line. Only paragraphs that
actually have a href pushed onto its own line are re-flowed; blockquotes,
indented and list-item lines are skipped.

**Alternatives considered:** re-wrapping at the markdown level before
glamour. Rejected because glamour's push-out decision depends on its own
layout, so pre-render wrapping cannot reliably detect or fix the failure;
post-render re-flow is deterministic and testable at the same width the app
renders at.

### D2: Fix the wrapped-href residual by extending the strip's drop range

`StripLinkHrefs` was split into `collectLinkSpans` + `groupLinkUnits` so the
span-to-unit mapping is explicit. When a wrapped href's closing bracket lands
on a continuation line and only styled whitespace (fill padding + wrap
newline) follows it to the line end, the unit's drop now extends through that
trailing whitespace so the paragraph re-wraps as one flow instead of splitting
at a whitespace-only line.

**Alternatives considered:** cleaning up whitespace-only lines later in the
pipeline. Rejected as fragile — the residual line is indistinguishable from a
real paragraph break by then, and reflow logic is simpler when the strip
output already contains no dangling whitespace.

### D3: Content-anchored selection restore with re-apply on real size

`persistSelection` records, alongside the raw offset, an anchor: the image
block URL when the first visible line is inside a composed block, else the
first text line's leading content (`AnchorableLine`) plus the gap to the
viewport top. `restoreSelection` calls `RestoreAt`, which re-locates the
anchored line (block → placeholder row via `URLIndex`; text → `FindAnchorLine`
minus the saved gap, clamped) and falls back to the raw offset when no usable
anchor exists. Because the restore composition happens at 80x24, the first
real `WindowSizeMsg` re-applies the anchor via a `pendingRestore` field on the
model.

**Alternatives considered:** restoring only once and tolerating the size
mismatch. Rejected — the re-apply on real size is what makes the position
correct in the common resized/placeholder case. Re-computing the anchor at the
real size instead of re-applying the persisted one was rejected because the
persisted anchor is content-based and therefore size-independent by design.

### D4: Bump recompose offset only by lines above the fold

`RecomposeInPlace` computes, for the first image whose old anchor row is at or
below the old offset, the number of lines inserted above the viewport top and
bumps the offset by exactly that, then runs `SnapAfterRecompose`. A block
materializing entirely below the viewport top therefore does not shift the
reader.

**Alternatives considered:** bumping by the whole line-count delta. Rejected —
that shifts the reader whenever any block below the fold materializes, which
is the bug this change fixes.

## Risks / Trade-offs

- **Re-flow cost on every expanded-mode render** → The pass is linear in the
  rendered paragraph and only runs for paragraphs that actually have a
  pushed-out href; the cached derivation keeps re-renders bounded.
- **Anchor ambiguity after content drift** → A text anchor that no longer
  matches falls back to the raw offset; a block anchor that no longer exists
  likewise. The fallback is lossless (offset retained), so a missed anchor
  degrades to today's behavior.
- **OSC 8 across a re-wrap** → Emitting OSC 8 per wrapped line is the only
  portable way to keep each segment clickable; the cost is repeated
  sequences in the rendered text, which the re-flow already controls.
- **Recompose accounting interacts with snap rules** → `SnapAfterRecompose`
  still runs after the bump, so an image poking past the fold re-snaps to its
  caption as before; the new rule only prevents wholesale offset shifts.

## Migration Plan

No schema migration beyond the already-committed `LastSelection` fields
(`ArticleBlock`, `ArticleAnchor`, `ArticleGap`), which are additive JSON
fields with safe defaults (empty → raw-offset fallback). Rollback is a revert
of the implementation; persisted anchors from the new version degrade to the
raw offset if read by an older binary.

## Open Questions

None.
## Context

The `x` toggle (article_update.go:57-59) flips `m.linkExpand` and calls
`recomposeArticle()`. `recomposeArticle` (image.go:562-586) preserves position
by bumping the old offset by the total line delta
(`inserted := len(m.article.lines) - oldLen`), which is only correct for lines
inserted at the top of the article. Link expansion adds lines at *every* link
in the body, so the offset bump overshoots and the first visible line drifts
(confirmed with a scratch test: offset 20 → 50 when it should have become 26).

`articleState.lines` (article.go:425) is the full rendered body *including the
header* (`strings.Split(rendered, "\n")`), so the first visible line is always
`lines[viewport.YOffset]`.

## Goals / Non-Goals

**Goals:**
- After a compact↔expanded toggle, the content line that led the viewport
  before the toggle leads it again.
- Bottom-pinned behavior is preserved: a toggle while pinned to the end of the
  article stays pinned to the end (image.go:573-579 `atBottom` → `GotoBottom`).

**Non-Goals:**
- Changing scroll-keep behavior for image recomposes — the total-delta bump
  plus `SnapAfterRecompose` there is correct (images insert lines at a single
  point) and already tested.
- Preserving scroll position across article *switches* (out of scope).

## Decisions

### 1. Anchor on the first visible line's text, not the total line delta

Before the recompose, capture the first visible line's rendered text:
`anchor = ansi.Strip(lines[oldOffset])`, trimmed. After the recompose, line
`x` re-finds the same content by scanning the new `lines` for the first index
whose stripped text equals the anchor (falling back to the index whose stripped
text contains the anchor as a line-prefix, for the case where a wrapped link
name is split differently by the new mode). Then `SetYOffset(x)`.

Rationale: line-text is the only stable cross-mode identity available — there is
no line→source-position mapping out of the HTML→markdown→render pipeline. Link
*names* are identical across modes; only the URL text (and wrapping) differs,
so matching on visible text is reliable exactly where the toggle changes things.

**Alternative rejected:** preserve position by mapping the old first line to its
source offset. Not possible — content is fully rendered before the viewport
exists, and no reverse mapping is kept.

**Alternative rejected:** keep the total-delta bump but subtract the links'
share. The bump is baked into `recomposeArticle` and correct for image loads —
mutating it would risk the well-tested image path.

### 2. Do the anchor pass only on the toggle path

Add a dedicated recompose path used by `LinkExpandToggle` — e.g. a boolean
`preserveFirstLine` parameter threaded into the recompose plumbing, defaulting
to false for the image-load and test callers, true only from the toggle
handler (article_update.go:59). Alternatively a small standalone helper that
captures the anchor, calls `recomposeArticle()`, then re-positions. The helper
keeps `recomposeArticle` unchanged and therefore keeps its regression tests
green; preferred.

### 3. Keep bottom-pinning invariants

The anchor path only re-positions when the pre-toggle offset was not at the
bottom (`oldOffset >= oldLen - vpH` → keep `GotoBottom()`), preserving the
existing `atBottom` gate. If the re-found line index now exceeds the new total,
clamp to `GotoBottom()`.

## Risks / Trade-offs

- **Anchor line not found** (e.g. the first visible line was an image-load
  marker line or a header line that changed shape) → fall back to the current
  total-delta behavior; position drift is no worse than today.
- **Duplicate text lines**: if the same stripped text appears earlier in the
  re-rendered article, the scan could land on the wrong instance. First-match
  scan is deterministic; realistic articles rarely repeat an exact full visible
  line. → Scan from the old offset's neighborhood (first index with matching
  text) keeps this bounded.
- **Wrapped link names split differently across modes** → the prefix fallback
  (`stripped line starts with anchor`) anchors on the line carrying the same
  leading text.

## Open Questions

None — the spec behavior (anchor the first visible line, keep bottom pinning)
fully determines the approach and the task breakdown.
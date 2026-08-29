## Context

See `proposal.md` - Why. The relevant call sites and their current disagreement:

| Location | Computes | Uses |
|---|---|---|
| `article.go:79` `newArticleState` | `contentW = width-2*padX-2`, `vpH = height-2*padY-2` | raw `padY` |
| `article.go:248` `contentRect` | same, for mouse mapping | `clampPadY` |
| `border.go:61-72` `renderArticleBorder` | `textW`, `viewportH`, `effPadY` | `clampPadY` |

The repo targets Go 1.27, so `min`/`max` builtins are available (`scroll.go`
already uses them). `internal/ui` has extensive golden/behavior tests that pin
current rendering; they are the regression gate.

## Goals / Non-Goals

**Goals:**

- Make the article content geometry derivable from exactly one function.
- Eliminate the copy-pasted helper logic without changing rendered output.
- Keep the diff mechanical and reviewable, one helper at a time.

**Non-Goals:**

- No change to any glyph, color, padding default, or scroll behavior.
- No width-measurement unification (`runewidth` → `ansi.StringWidth`); deferred.
- No caching of `glyphsFor(m.ascii)` or other micro-optimizations.
- No restructuring of the `Model` fields or the `articleState` value/pointer
  indirection (`m.article.article`).

## Decisions

**D1 — One `contentGeom(w, h, padX, padY)` returning `(textW, viewportH, effPadY int)`.**

It must be the *clamped* variant so all three call sites agree. Rationale:
`renderArticleBorder` and `contentRect` already agree on `clampPadY`; only
`newArticleState` diverges. Unifying onto the clamped path is therefore the
correct conformance fix, not a new behavior.

- *Alternative considered*: keep raw `padY` and add clamping later. Rejected —
  it would preserve the exact bug we are removing.

The function lives in a new `geometry.go` next to the other small pure helpers,
so it has no implicit coupling to `Model`.

**D2 — `newArticleState` and `contentRect` both call `contentGeom`.**

`renderArticleBorder` receives `textW`/`viewportH`/`effPadY` via `contentGeom`
too, rather than recomputing its interior sizes. The border renderer keeps its
own glyph/scrollbar layout; only the shared dimensions move into the helper.

**D3 — `articleAtCursor() (*articleItem, bool)`.**

Returns the `*articleItem` under `m.list.cursor` and `false` when the cursor is
out of range or on a day header. Centralizes the
`visibleRows()` → bounds-check → `row.kind` → index-into-groups dance. Callers
that need the group (e.g. `persistSelection`, `restoreSelection`) keep their own
`visibleRows()` traversal since they operate on header rows too.

- *Alternative considered*: a richer `cursorTarget()` returning a discriminated
  union of header/article. Rejected — overkill; the header path is small and
  distinct.

**D4 — `wrapIndex(i, delta, n int) int`.**

Pure function returning `(i+delta) % n` with the negative-modulo fix-up. Used by
`moveListCursor` and `movePopupCursor`. Lives in `geometry.go`.

**D5 — `dayKey(t time.Time) string`.**

Returns `"undated"` for the zero time, else `t.Local()` truncated to midnight
and formatted `"20060102"`. Replaces the copies in `bucketDayGroups`,
`persistSelection`, and `restoreSelection`. The `json.go` `bucketArticles`
variant is intentionally left alone — it falls back to `FetchedAt` rather than
an undated bucket, so it is not the same rule.

**D6 — Merge `padToWidth` into `padRight`.**

`padToWidth` (`model.go`) and `padRight` (`width.go`) are identical
(`ansi.StringWidth` + trailing spaces). Keep `padRight` in `width.go`, delete
`padToWidth`, and update its single caller in `spliceStyled` (`model.go`).

## Risks / Trade-offs

- **Subtle geometry off-by-one** on resize/degenerate terminals → Mitigation:
  the existing `render_safety_test.go`, `border_top_test.go`, and mouse
  selection tests cover degenerate sizes; run the full `internal/ui` suite after
  each helper extraction, not just at the end.
- **Behavior change from the `clampPadY` unification** → Mitigation: this is the
  intended conformance fix and is already required by the "Render safety under
  degenerate dimensions" spec; a focused test asserts `newArticleState` and
  `contentRect` now agree for a small height + large `padY`.
- **Over-broad edit touching rendering** → Mitigation: commit each helper
  extraction as a separate change within the task list so `git diff` stays
  reviewable and a bisect can pinpoint any regression.

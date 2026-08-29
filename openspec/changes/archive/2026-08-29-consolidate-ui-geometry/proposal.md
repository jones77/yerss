## Why

The article content rectangle — the space inside the reader border where text
renders — is computed independently in three places (`newArticleState` in
`article.go`, `contentRect` in `article.go`, and `renderArticleBorder` in
`border.go`) and kept in sync **by hand**. Two of them already disagree subtly:
`newArticleState` uses raw `padY` while the other two use `clampPadY`, so mouse
hit-testing can diverge from what the border renderer actually draws on
degenerate terminals. The same package also carries several near-identical
helpers (wrap-around cursor, pad-to-width, day-group key, cursor→article lookup)
that have been copy-pasted as features landed. This is the mechanical cleanup
that removes the drift risk without changing any observable behavior.

## What Changes

- Introduce one shared function that computes the article content geometry
  (text width, viewport height, effective padding) and route all three call
  sites through it.
- Reuse `pageSize()` in `renderList` and `handleListClick` instead of
  re-deriving `m.height - 1` clamped to a minimum of 1 inline.
- Add a `wrapIndex(i, delta, n)` helper and use it for both the list cursor and
  the popup cursor, removing two identical modulo loops.
- Add a `dayKey(time.Time) string` helper for the `"undated"` /
  `"20060102"` calendar-day key, replacing three hand-written copies.
- Add an `articleAtCursor() (*articleItem, bool)` helper and use it in
  `openArticle`, `toggleReadAtCursor`, `persistSelection`, `restoreSelection`,
  and `handleListClick`.
- Merge `padToWidth` (in `model.go`) into `padRight` (in `width.go`) — they are
  functionally identical — leaving one width-padding helper.
- Sweep the `if x < 1 { x = 1 }` clamps to Go's built-in `max(1, x)` where the
  surrounding code is already idiomatic.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

_None._ This is a pure refactor: no spec-level behavior changes, so this change
declares `skip_specs: true`.

## Impact

- **UI layer**: `internal/ui` — `article.go`, `border.go`, `list.go`, `model.go`,
  `popup.go`, `selection.go`, `width.go`, plus a new small `geometry.go` (or
  equivalent) for the shared geometry helpers. No public API changes.
- **Tests**: existing `internal/ui/*_test.go` suite (render, border, mouse
  selection, day-groups, scroll, selection, polish) is the safety net; no new
  behavior is introduced, so these must continue to pass unchanged. A focused
  test for the extracted geometry function and the merged helpers is added.
- **Out of scope**: width-measurement library unification (`runewidth` vs
  `ansi.StringWidth`) is deliberately deferred to a separate change — it is
  rendering-behavior-sensitive and deserves its own careful treatment.

## Context

The failing inline image is `Screenshot-2026-10-01-at-2.18.31-PM.png?fit=2542%2C840` (article id 10431, position 1), a ~3:1 image that at ~100-column content width occupies roughly 33 rows — taller than the ~29-row viewport of the reported 31-line screen. AGENTS.md documents two interacting rules: a partially visible native placement is deleted (it would draw over the border), leaving those rows blank, while the halfblock path's `previewClippedImage` fills a poking-in image's visible rows so it never shows a blank strip. The native path currently lacks the equivalent fill. `snapYOffset` in `internal/ui/image.go` keeps single-line scrolls from leaving an inline image partially visible, with a down bottom-snap that fires on partial visibility.

## Goals / Non-Goals

**Goals:**
- A partially visible native photo shows its visible rows, never blank space.

**Non-Goals:**
- No change to halfblock-only rendering (it already fills).
- No change to native placement deletion rules themselves — only the fill of the visible rows where a placement is not shown.

## Decisions

**Reproduce from the local DB before changing code.** Per AGENTS.md, reproduce the article's real block geometry from the local SQLite DB and simulate the scrolls to determine whether the blank region comes from `snapYOffset` (a scroll lands in a partial-visibility state) or from the native placement cleanup (a placement is deleted while its rows remain in the window). The fix location follows from the reproduction.

**Fill partial-visibility rows with a preview rather than suppressing the snap.** Reusing the halfblock `previewClippedImage` approach (render the visible slice as halfblock art for the visible rows) is the lowest-risk path and mirrors the behavior already specified for the halfblock path. Suppressing or reworking the snap for tall-than-viewport images is a larger change and risks the bounce/loop guards documented in AGENTS.md.

## Risks / Trade-offs

- [Fix in the wrong layer (snap vs placement cleanup)] → reproduce with the real geometry and a focused test on the invoice photo before changing rules.
- [Preview flashes between native and halfblock on boundary rows] → keep the preview scoped to rows where the native placement is not shown; fully visible images stay native.

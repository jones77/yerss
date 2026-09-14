## 1. Tests

- [x] 1.1 Model test: toggle with a multi-paragraph article where the URL wraps (expands lines at every link) preserves the first visible line — assert `lines[newOffset]` (stripped) equals the pre-toggle first visible line, and offset is not the total-delta bump.
- [x] 1.2 Model test: toggle at the bottom (offset pinned past `oldLen-vpH`) keeps `GotoBottom` behavior — offset stays at the new bottom, no blank space.
- [x] 1.3 Model test: toggle back (expanded → compact) returns to the same anchor line.
- [x] 1.4 Unit test for the anchor-repositioning helper (line re-find with exact text and prefix fallback).

## 2. Implementation

- [x] 2.1 Add an anchor-preserving toggle path in `article_update.go` (`LinkExpandToggle`): capture the first visible stripped line, call `recomposeArticle()`, re-find that line in the recomposed `m.article.lines`, and `SetYOffset` to it; preserve `atBottom` / `GotoBottom` behavior and clamp overflow to the new bottom.

## 3. Verification

- [x] 3.1 `go test ./internal/ui/...`
- [x] 3.2 `go vet ./...`
- [x] 3.3 `openspec validate --all --strict`
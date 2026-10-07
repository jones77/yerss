## Why

In compact link mode (the default), a link name can render fused to the following word. In the Intercept article "Oregon Police Are Charging a Local News Outlet $2.5M for Records About One Cop" the phrase renders as `claimedsearching` instead of `claimed searching` — the space after the link name is lost. The surrounding HTML has the space (`claimed</a> searching`), so the loss happens in the markdown/render/strip pipeline.

## What Changes

- Fix the compact-mode link re-wrap so a stripped link name is separated from the following text by exactly one space (never zero, never two).
- Add a regression scenario for the missing-space case alongside the existing double-space coverage.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `compact-link-mode`: extend the "stripped link re-wrap preserves single word spacing" requirement so the re-wrap guarantees exactly one space between a stripped link name and following text, covering both the existing double-space case and the new missing-space case.

## Impact

- `internal/compose/links.go` (`StripLinkHrefs`) and/or `internal/compose/reflow.go` (`ReflowShortLines`) — wherever the space between a link name and following text is dropped.
- Possibly `internal/compose/inline.go` / `internal/ui/articleview/article.go` if the loss is traced to the markdown→render step rather than the strip step (to be confirmed during reproduction).
- No new dependencies.

## Why

`internal/ui` is a large package mixing list, article, popup, rendering primitives, and article composition. Splitting it by concern makes each piece easier to navigate and test. This is the last, optional refactor in the sequence.

## What Changes

- Reorganize `internal/ui` into subpackages: `components` (list, article, popup, help), `render` (border, glyphs, palette, width, scroll), and `compose` (markdown pipeline, image-block composition, snap), with the model as the thin top of the package.
- No behavior change.

## Capabilities

### New Capabilities

### Modified Capabilities

## Impact

- `internal/ui/**`: package split. If import cycles arise, fall back to a file-level reorganization within one package.

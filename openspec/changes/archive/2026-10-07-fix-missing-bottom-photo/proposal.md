## Why

In the Intercept article "Oregon Police Are Charging a Local News Outlet $2.5M for Records About One Cop", the inline invoice photo (captioned "The open invoice by the City of Eugene") renders as empty space instead of the photo. The image is a wide screenshot (`2542x840`, roughly 3:1) that, scaled to content width, is taller than the viewport; when the reader scrolls to it, the region where the photo should appear is blank.

## What Changes

- Ensure a native inline photo that is partially visible in the viewport shows its visible rows (a clipped preview) rather than blank space, matching the halfblock path's existing behavior of never showing a blank strip for a poking-in image.
- Reproduce the exact geometry from the local database before changing code, to pin whether the blank region comes from the scroll-snap rules (`snapYOffset`) or from the native-placement cleanup deleting a partially-visible placement.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `native-image-rendering`: add a requirement that a partially visible native photo renders its visible rows as a preview instead of blank space, so a photo that is taller than the viewport is never shown as empty rows.

## Impact

- `internal/ui/image.go` (`snapYOffset` and related scroll-snap logic) and/or the native image composition/cleanup path in `internal/image/`.
- The halfblock `previewClippedImage` logic may be generalized so the native path reuses it for the visible rows of a partially visible photo.
- No new dependencies.

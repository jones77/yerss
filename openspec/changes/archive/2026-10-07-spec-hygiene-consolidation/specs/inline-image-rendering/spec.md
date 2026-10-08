## REMOVED Requirements

### Requirement: Inline image discovery
**Reason**: Merged into the consolidated `image-rendering` capability.
**Migration**: See `openspec/specs/image-rendering/spec.md` — "Inline image discovery".

### Requirement: Inline image rendering with attribution
**Reason**: Merged into the consolidated `image-rendering` capability, which unifies the lead and inline attribution chain (figcaption → alt → link text → `photo: <source>`) and removes the document-global credit fallback.
**Migration**: See `openspec/specs/image-rendering/spec.md` — "Image attribution" and "Halfblock image rendering".

### Requirement: Linked inline images render without link fragments
**Reason**: Merged into the consolidated `image-rendering` capability's attribution requirement (the linked-image fragment consumption is covered by the link-text caption fallback).
**Migration**: See `openspec/specs/image-rendering/spec.md` — "Image attribution".

### Requirement: Inline images that duplicate the lead render in place
**Reason**: Merged into the consolidated `image-rendering` capability's dedup requirement.
**Migration**: See `openspec/specs/image-rendering/spec.md` — "Duplicate lead image suppressed".

### Requirement: Inline image persistence
**Reason**: Merged into the consolidated `image-loading` capability's storage requirement.
**Migration**: See `openspec/specs/image-loading/spec.md` — "Unified image storage".

### Requirement: Inline image loading is viewport-driven
**Reason**: Merged into the consolidated `image-loading` capability's load-frontier requirement.
**Migration**: See `openspec/specs/image-loading/spec.md` — "Viewport-driven load frontier".

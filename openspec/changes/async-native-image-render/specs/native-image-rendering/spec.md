## MODIFIED Requirements

### Requirement: Native photo render with block placeholder

On a full-image-capable terminal, opening an article with a lead image SHALL
display a halfblock block immediately — the stored block when present, otherwise
one rendered from a fetch — and SHALL fetch the real photo asynchronously. The
native render itself — decoding the photo, scaling it to the display box, and
encoding it for the terminal — SHALL also be performed asynchronously and never
on the UI thread, so that opening or resizing an article with a native photo
never freezes the UI on a CPU-bound operation. When the photo arrives it SHALL
be rendered natively (inline image protocol) in place of the block, sized to fit
the content width with aspect ratio preserved and capped to the viewport,
centered within the content width exactly like the halfblock block (its reserved
rows span only the photo's own cell box), and SHALL be part of the scrolling
content with the photo attribution centered beneath. The rendered native output
SHALL be cached in memory keyed by render size, so re-opening an article at the
same content width SHALL display the photo immediately from cache without
re-downloading or re-rendering. When the photo fetch fails, the block SHALL
remain as the final render. Resizing the terminal SHALL re-scale the native
render from the cached photo without re-fetching, asynchronously and without
blocking the UI. The fetched photo and its rendered native output SHALL be held
only in memory for the session and SHALL NOT be persisted.

#### Scenario: Block shown before photo fetch

- **WHEN** a full-image-capable terminal opens an article with a lead image URL and a stored block
- **THEN** the stored block renders immediately and the photo fetch runs in the background

#### Scenario: Native render replaces the block

- **WHEN** the background photo fetch succeeds on a full-image-capable terminal
- **THEN** the halfblock block is replaced by a native inline render of the photo sized to the content width, with the attribution still beneath

#### Scenario: Height-capped photo is centered

- **WHEN** a native render is narrowed below the content width by the viewport height cap or the attribution wrap
- **THEN** the photo and its reserved rows span only the photo's own cell box and the block is centered within the content width with equal left and right margins, with the attribution centered beneath, matching the halfblock block's placement

#### Scenario: Native fetch failure keeps the block

- **WHEN** the background photo fetch fails on a full-image-capable terminal
- **THEN** the halfblock block remains as the final render and no error is surfaced

#### Scenario: Resize re-scales native render without re-fetch

- **WHEN** the terminal is resized while a native photo render is displayed
- **THEN** the native render is re-scaled to the new content width from the cached photo with no network request

#### Scenario: Photo not persisted

- **WHEN** a full-image-capable terminal fetches a lead photo
- **THEN** the photo bytes are cached only in memory for the session and never written to the database

#### Scenario: Opening a high-resolution photo does not block the UI

- **WHEN** a full-image-capable terminal opens an article whose lead photo is high-resolution (a large source image)
- **THEN** the article text renders immediately and the photo's decode, scale, and encode happen in the background rather than on the UI thread

#### Scenario: Re-opened article shows the cached native render instantly

- **WHEN** the user re-opens an article whose photo was already rendered natively at the current content width
- **THEN** the photo appears immediately from the in-memory render cache, with no network request and no re-decode or re-encode

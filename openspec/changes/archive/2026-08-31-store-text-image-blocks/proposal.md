## Why

The database currently stores the raw compressed photo bytes of every fetched
lead image, so the article database bloats to megabytes while the UI can only
ever display a low-fidelity halfblock representation. The photo bytes are
fetched once, persisted forever, and re-decoded on every reopen. Instead the
system should persist the small text-generated halfblock art that is actually
displayed, and — on terminals that support real inline images (iTerm2, kitty,
WezTerm) — show that stored block immediately as a placeholder before fetching
the real photo and rendering it natively, without ever storing the photo.

## What Changes

- **BREAKING (storage)**: Stop persisting raw lead-image bytes in the
  `articles.image_data` column. The database SHALL persist the rendered
  halfblock text block instead (stored in a new `image_block` column), and
  legacy stored photo bytes SHALL be converted to blocks and then purged.
- New in-memory session cache holds decoded images (for re-rendering on resize)
  and, on native terminals, the fetched raw photo bytes (for re-scaling native
  renders); neither is persisted.
- On non-native terminals the stored block is the final render: the photo is
  fetched only when no block is stored yet, so reopen and restart render from
  the block with no network request.
- On terminals with full image capabilities (iTerm2/kitty/WezTerm detected via
  environment), opening an article SHALL show the stored halfblock block
  immediately, then fetch the real photo asynchronously and render it natively
  (OSC 1337 inline image / kitty graphics protocol), replacing the block. A
  failed native photo fetch falls back to the block.
- Terminal image capability detection is added (environment-based, mirroring
  the existing ASCII/light-background detection).

## Capabilities

### New Capabilities
- `native-image-rendering`: Rendering the real fetched photo inline via the
  terminal's native image protocol (OSC 1337 / kitty graphics) on
  full-image-capable terminals, with the stored halfblock block shown as an
  immediate placeholder, and environment-based capability detection.

### Modified Capabilities
- `reader-ui`: The article lead image rendering and async load lifecycle change:
  the database persists rendered text blocks instead of raw photo bytes, blocks
  serve as the offline/first render, and native-capable terminals swap in a
  real photo render after it loads.

## Impact

- `internal/store`: new `image_block` column + migration, block get/set methods,
  legacy `image_data` purge; tests updated.
- `internal/image`: cache hierarchy reworked (decoded cache, block cache, raw
  bytes cache), native protocol detection + encoding, message types extended.
- `internal/ui`: article image block composition chooses between stored block,
  halfblock re-render from decoded cache, and native render; load cmd batching;
  resize re-render; tests updated.
- No new external dependencies: image scaling uses `golang.org/x/image` (already
  indirect) and the image codecs already imported; native protocols are emitted
  as escape sequences.
# Design: store-text-image-blocks

## Context

Today `internal/image` fetches a lead photo, decodes it, renders halfblock art
on demand, and persists the raw compressed bytes into `articles.image_data`
(`store.SetImageData`/`GetImageData`). `internal/ui` renders the block from the
session's in-memory decoded-image cache (`m.imgCache`), which is the only
source that supports re-rendering at a new terminal width. See proposal.md for
the motivation.

## Goals / Non-Goals

Goals:
- Stop persisting raw photo bytes; persist rendered halfblock text blocks.
- Blocks serve as the instant/offline render (no network on reopen).
- On full-image-capable terminals, show the block first, then a native inline
  photo render from memory-cached bytes.
- Keep the in-memory decoded-image cache so viewport resizes re-render without
  re-fetching.

Non-Goals:
- No animated/GIF support; a lead photo is a single still frame.
- No user-facing configuration for image protocols (detection is
  environment-based, like ASCII/background detection).
- No offline storage of the native photo render.

## Decisions

### 1. Store blocks in a new `image_block` TEXT column; legacy `image_data` is purged on conversion

Add `image_block TEXT` via `migrateImageBlockColumns`. `SetImageBlock(id,
block)` writes the block and clears `image_data` (`UPDATE articles SET
image_block = ?, image_data = NULL`). `GetImageBlock` returns the stored block
or `""`. `GetImageData` stays (read-only) so the load path can detect legacy
bytes; `SetImageData` is removed — nothing writes raw bytes anymore.

Alternatives considered: reusing `image_data` for block text (ambiguous
content, messier legacy detection), or keeping both columns populated
(defeats "don't store the photo"). A dedicated column makes the content type
explicit and the legacy purge unambiguous: the first block write for an
article drops its photo.

### 2. Block width is self-describing; a width mismatch re-fetches

Mosaic renders exactly `w` cells per line (the width is passed in), so a stored
block's rendering width is `max(ansi.StringWidth(line))` over its lines. The
load path compares that to the requested `contentW`; on mismatch it treats the
stored block as a miss and re-fetches/re-renders/re-persists. This keeps the
"block fits the content width" contract true without storing a width field.

### 3. Load commands resolve through a cache hierarchy and emit block/photo messages

Replace `image.LoadCmd` with two commands, both of which mutate shared session
caches as a side effect before returning a message (the existing LoadCmd
already does this with `cache.Set`):

- `image.BlockCmd(cache, blocks, st, id, url, width, maxHeight, fetch)` (all
  terminals): resolves the halfblock at `width`.
  1. decoded image in `Cache` → render block at `width`.
  2. stored block in DB with matching width → serve its lines.
  3. legacy raw bytes in `image_data` → decode, render, persist block (purges
     the photo), serve.
  4. else network fetch → decode, render block at `width`, persist block,
     cache decoded image, serve.
  Returns `BlockMsg{Key, Lines}` or `FailedMsg{Key}`. With `fetch=false` (the
  native-terminal placeholder path) step 4 never runs — `PhotoCmd` owns the
  network so a fresh article is not downloaded twice.

- `image.PhotoCmd(cache, blocks, photos, st, id, url, width, maxHeight)`
  (full-image-capable terminals only): fetches the photo into the `Photos`
  cache and, when the photo was not already decoded, also renders and persists
  the halfblock block from it so future opens have a placeholder. Returns
  `PhotoMsg{Key}` or `FailedMsg{Key}`.

On a native terminal the UI fires both commands batched: `BlockCmd(fetch=false)`
serves a stored placeholder immediately ("block before photo"), and `PhotoCmd`
fetches the real photo. A fresh article with no stored block has no placeholder
to show, so only `PhotoCmd` runs to completion — a single download.

New session caches in the image package: `Blocks` (URL → rendered lines) and
`Photos` (URL → raw fetched bytes). The existing `Cache` (URL → decoded image)
is unchanged. Messages: `BlockMsg{Key, Lines}`, `PhotoMsg{Key}`,
`FailedMsg{Key}`; `LoadedMsg` is removed (the decoded image never crosses a
message anymore).

New `NativeRenderer` in the image package turns cached photo bytes into the
lines shown inline. The UI (non-native) keeps using the existing `Renderer`
(halfblocks) for decoded images.

### 4. Native rendering: OSC 1337 (iTerm2/WezTerm) and kitty graphics

`image.DetectProtocol()` reads the environment, mirroring `detectAsciiNeeded`:
`TERM_PROGRAM=iTerm.app`/`WezTerm` → OSC 1337; kitty-family terminals are
detected from `TERM_PROGRAM=kitty`/`ghostty`, a `TERM` matching `kitty`/`ghostty`,
or `KITTY_WINDOW_ID`/`KITTY_PID`/`GHOSTTY_RESOURCES_DIR` being set → kitty
graphics; otherwise none. A `detectProtocol` package var keeps it testable, and
the Model carries the resolved protocol in its `imgNative image.NativeRenderer`
field (set directly by in-package tests).

Both paths decode the photo to RGBA and re-encode PNG, then emit the escape.
iTerm2's OSC 1337 accepts `width`/`height` in cells; the kitty/ghostty escape
scales the image to the cell box via the `c`/`r` keys. The image is encoded at
the terminal's cell-pixel resolution — the window's physical size divided by
its cell count, queried via `TIOCGWINSZ` (falling back to 8×16) — so the
terminal displays it sharply rather than upscaling a low-res source. The kitty
payload is chunked into ≤4096-base64-char APC continuations (`m=1`/`m=0`) as
the protocol requires for remote clients — Ghostty rejects oversized single
chunks — with `C=1` (no cursor movement) so the reserved rows align, `q=1`
(quiet, no response into the app's stdin), and stable `i`/`p` ids so
re-transmissions replace the terminal's image and placement instead of
accumulating (Ghostty's known anonymous-placement leak). The longest encode
edge is capped so an oversized cell box cannot balloon the payload; the image
is identified by a stable URL-derived name/id.

The block lines span the fitted image width `w` (the escape line's padding and
each reserved row), and the reserved height is the rows `fitDims` computes, so
the escape occupies exactly the same cell box the halfblock block would: the
block reports its own width, `articleImageBlock`'s centering, attribution-wrap,
and height-cap budget logic apply to it unchanged, and a height-capped photo is
centered within the content width instead of left-aligned.

### 5. UI composition

`articleImageBlock` gains source precedence:
- native terminal: `Photos` hit → native lines; else block from `Blocks`
  (populated from DB or render); else nil (load in flight).
- non-native: `Cache` hit → re-render at `contentW`; else `Blocks` hit → stored
  lines; else nil.

It also reports whether the composed lines are a native render, recorded on
`articleState.nativeImg`.

`fireImageLoad` fires `NativeCmd` on native terminals and `BlockCmd` otherwise.
`onBlockLoaded`/`onPhotoLoaded` recompose the open article preserving the
scroll offset (same logic as today's `onImageLoaded`). `onImageFailed`
recomposes only when a block became available (the native DB-block path), so
the placeholder appears even when the photo fails. `WindowSizeMsg` recomposes as
today; on a width change where the article has an image URL but no session
cache matches (stored block at a different width), it re-fires the load command
once per mismatch.

### 6. Deleting kitty placements on frames that no longer show the photo

In the kitty graphics protocol, placements float above text and text-erase
commands have no effect on them, so a placement made by an article frame
survives the repaint into the list view (and stacks over the next article's
photo) unless explicitly deleted. `NativeRenderer.Clear()` emits
`<ESC>_Ga=d,d=a,q=1<ESC>\` — the delete action for all visible placements,
quiet like the transmit; it returns "" for OSC 1337, whose cell-bound images
are erased by the repaint itself. `Model.View()` prepends the clear sequence
(zero-width, position-independent) to any frame that does not display the open
article's native image: non-article views, articles without a native render
(the placeholder phase), and articles whose photo rows are scrolled outside
the viewport window (the caption may still be visible). A frame that does
display the photo omits the delete — its re-transmission with the stable
`i`/`p` ids replaces the placement — and scrolling the photo back into view
re-renders its line, re-transmitting the photo after any earlier delete. The
decision is stateless per frame, derived from `articleState.nativeImg` plus
the viewport window intersecting the photo's row range, so no placement
bookkeeping can drift.

Alternatives considered: deleting only the current article's id
(`d=i,i=<id>`), which needs tracking of every placed URL and still leaks on
transient states; and freeing image data (`d=A`), which buys nothing since the
transmit escape re-carries the full payload.

### 7. Down-snap lands on the caption

The photo and its caption are no longer one atomic scroll unit. A downward move
that would clip the photo snaps to `capStart` — the first attribution line — so
the caption stays readable, and the caption itself scrolls line by line (the
down-snap trigger range is the photo's rows [imgStart, capStart-1] only, so
there is no snap loop). Without an attribution `capStart` is the line past the
block, preserving the old skip-past behavior. Upward moves keep the full-block
reveal (`imgStart`). `articleState.capStart` (imgStart plus the image-row count
reported through `articleImageBlock`/`fitBlock`) also refines the kitty
placement-clear check: a frame showing only the caption deletes the photo's
placement, since its rows are out of the window and the transmit line is not
rendered.

## Risks / Trade-offs

- A stored block is fixed at its render width; a different terminal width after
  restart re-fetches the photo (by design, per spec). Mitigation: in-session
  the decoded image is cached, so resizes within a session never re-fetch.
- Re-emitting a native image escape every frame can be heavy. Mitigation: the
  photo is downscaled to display size before encoding so the payload is small,
  and iTerm2/kitty dedupe by name+size/id so the payload is not re-decoded.
- kitty/ghostty re-transmission can leak placements when the image id is reused
  without a placement id. Mitigation: a stable `p=` placement id makes
  re-transmissions replace the placement (the spec's documented replacement
  semantics), and chunking keeps each APC within the protocol's size limit so
  the image transmits at all.
- kitty's placement-in-stream can flicker under viewport re-rendering, a known
  kitty quirk. Mitigation: kitty support is secondary; iTerm2/WezTerm via OSC
  1337 is the reliable primary path, and a failed or absent native render
  always falls back to the halfblock block.
- Legacy DBs hold raw bytes; `image_data` is only purged once each article is
  reopened post-upgrade. Mitigation: the purge happens on the first load, and
  the block is small so the DB shrinks naturally.

## Migration Plan

- Schema: idempotent `ALTER TABLE articles ADD COLUMN image_block TEXT` (guarded
  by `PRAGMA table_info`, like the existing column migrations).
- Data: legacy `image_data` bytes are converted to blocks and purged lazily on
  article open; no one-time script. Pre-upgrade DBs keep photos until reopened,
  then drop them.
- Rollback: a DB with `image_block` populated can be read by the old build via
  `image_data` only if blocks were never written; since `SetImageBlock` clears
  `image_data`, rollback is not fully lossless. Acceptable for this app; the
  new build is the target.

## Open Questions

- Real-world sizing/placement tuning in iTerm2, WezTerm, kitty, and Ghostty
  (the `c`/`r` cell-box scaling and reserved-row math assume the terminal honors
  the protocol's cell grid); these do not affect the spec.
- Resolved: Ghostty's stable-`p=` re-transmission prevents same-image
  accumulation, but placements still survived view changes because nothing
  deleted them — decision 6 deletes all visible placements on frames that no
  longer show the photo.
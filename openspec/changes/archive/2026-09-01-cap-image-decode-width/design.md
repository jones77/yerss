## Context

Images enter through `BlockCmd` (halfblock path) and `PhotoCmd` (native photo
bytes), both of which `image.Decode` the full compressed bytes and retain the
decoded bitmap in `ImgCache` for the session; the native render
(`renderNativeFitted`) decodes the same bytes again to scale to the cell box.
The stdlib decoders always decode to full resolution, so the only way to bound
the *retained* bitmap is to downscale immediately after decode. `scaleTo`
(golang.org/x/image/draw CatmullRom) already exists for this. See proposal.md
for motivation.

## Goals / Non-Goals

**Goals:**
- Retain at most a 2048-pixel-wide decoded bitmap per image, so session memory
  is bounded regardless of source resolution.
- Keep the database and the raw-photo cache holding the original compressed
  bytes, unchanged.

**Non-Goals:**
- No re-encode, no downscale-then-persist, no database migration.
- No cache-count eviction (LRU) — out of scope here.
- No CDN URL rewriting to request pre-resized variants.

## Decisions

### D1 — `maxDecodeWidth = 2048`, applied by one shared `decodeCapped` helper

Add a package constant and a helper in `internal/image`:

```
decodeCapped(data []byte) (image.Image, error)
    // image.Decode; if bounds.Dx() > maxDecodeWidth, scaleTo to
    // (maxDecodeWidth, h) preserving aspect; else return the decoded image.
```

- **Why a helper, not a per-site change:** three decode sites feed rendering
  and caching (`BlockCmd` stored-photo branch, `BlockCmd` fetch branch,
  `PhotoCmd` placeholder, `renderNativeFitted`); one helper keeps the cap
  consistent and testable.
- **Why width, not longest edge:** the user specified "2048 pixels wide." A
  width cap preserves the dominant landscape-photo case; unusually tall
  portraits are accepted (see Risks).
- **Small images are never upscaled:** images at or below 2048 wide decode
  unchanged, preserving the existing small-logo handling.

### D2 — Use `decodeCapped` at every decode site, keep persistence untouched

`BlockCmd`, `PhotoCmd`, and `renderNativeFitted` replace their raw
`image.Decode(bytes.NewReader(data))` with `decodeCapped`. The decoded bitmap
cached in `ImgCache` is therefore capped. `Photos` and the database continue to
hold the original compressed bytes: `persistImage` and `SetArticleImage` are
unchanged, and no migration runs.

### D3 — Native render scales the capped image to the cell box as before

The native path already scales to the display box (`scaleTo` → `pixelDims`).
It now starts from the capped decode rather than the full-res decode, so the
terminal-side payload is unchanged and the transient source decode is bounded.

## Risks / Trade-offs

- [Transient full-res decode still allocates] → The stdlib decodes the whole
  source before we can downscale, so a single transient full-res allocation
  still occurs per decode. Mitigation: it happens off the UI goroutine inside a
  `tea.Cmd`, and only the *retained* cache is capped. A streaming/partial decode
  is out of stdlib scope.
- [Width cap does not bound very tall images] → A 2048-wide portrait could still
  be thousands of pixels tall. Mitigation: article photos are overwhelmingly
  landscape; accepted for now, and `maxDecodeWidth` is a single tunable if a
  height or longest-edge bound is wanted later.
- [Ultra-wide native sharpness] → 2048 is below the existing `pixelDims` 4096
  ceiling, so an extremely wide terminal (e.g. 300 cells) would upscale the
  source slightly. Mitigation: 2048 exceeds typical content widths; accepted.
- [23 MB originals remain in DB and on the wire] → Deliberately out of scope:
  the user wants compressed originals stored, and the CDN-resize optimization
  is deferred.

## Migration Plan

None. The change is behavior-only: no schema change, no stored-byte rewrite.
Rollback is simply reverting the code; stored bytes were never altered.
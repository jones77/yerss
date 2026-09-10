## 1. Bounded native render payload

- [x] 1.1 Change `pixelDims` in `internal/image/native.go` to clamp the longer
      encoded edge to a fixed `maxNativeEdge` (e.g. 1280px), scaling the shorter
      edge to preserve aspect ratio, keeping the existing behavior below the cap
- [x] 1.2 Verify the displayed cell box and aspect ratio are unchanged at
      typical widths (existing native render tests)

## 2. Kitty placement-reference re-show

- [x] 2.1 Add a placement-reference escape builder (`a=p,i=<id>,p=<id>,c=W,r=H,C=1`,
      no payload) to `internal/image/native.go`, carrying the same image id as
      the full transmit so `NativeRenderID` and delete-by-id keep working
- [x] 2.2 Drive the choice in the block's first line from `nativeSent`: full
      transmit when the id was not yet transmitted (or its data was freed),
      placement reference on later frames while the image stays visible
- [x] 2.3 Clear `nativeSent[url]` in `computeNativeClear` when a block is
      scrolled out or clipped (its data is freed by the `d=I` delete), so the
      scroll-back frame transmits once before referencing again

## 3. Tests and verification

- [x] 3.1 Add tests: a visible native image's later frames carry a placement
      reference (`a=p`) with no payload; scroll-out then scroll-back transmits
      once and then references; a re-render at a new size transmits once and
      then references; encoded pixel dims never exceed `maxNativeEdge`
- [x] 3.2 Update any existing kitty render/clear tests whose expected escape
      strings change (full transmit on first frame remains)
- [x] 3.3 Run the full `internal/ui` and `internal/image` suites plus `go vet`;
      manually verify on kitty/ghostty that a photo in view accepts input
      promptly (no multi-megabyte re-transmit per keystroke)
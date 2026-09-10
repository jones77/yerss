package image

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/sys/unix"
)

// NativeRenderer renders fetched photo bytes as a terminal-native inline
// image: the iTerm2 OSC 1337 inline-image protocol or the kitty graphics
// protocol. The photo is scaled to the on-screen cell box and re-encoded as
// PNG, then embedded in an escape sequence at the start of the block's first
// line; the remaining reserved rows are blank so the image occupies the same
// cell box the halfblock block would.
type NativeRenderer struct {
	Protocol Protocol
	// Ghostty marks a Ghostty terminal, whose kitty-graphics delete handling
	// is partial; the exit clear appends a delete-all fallback for it.
	Ghostty bool
}

// cellPixelSizeFn is overridable in tests; by default it detects the terminal
// cell size via TIOCGWINSZ.
var cellPixelSizeFn = detectCellPixelSize

// detectCellPixelSize returns the terminal cell width/height in pixels,
// derived from the physical window size divided by its cell count. It falls
// back to a typical 8×16 cell when the terminal does not report a pixel size.
func detectCellPixelSize() (cw, ch int) {
	if ws, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ); err == nil {
		if ws.Xpixel > 0 && ws.Ypixel > 0 && ws.Col > 0 && ws.Row > 0 {
			return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
		}
	}
	return 8, 16
}

// Render turns raw photo bytes into block lines that display the photo
// natively, fitted to width cells and capped to maxHeight rows (a height cap
// narrows the box to w cells). It decodes data once and renders it; callers
// that fit the same photo across several candidate heights decode it once and
// call RenderImage directly.
func (r NativeRenderer) Render(data []byte, url string, width, maxHeight int) ([]string, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return r.RenderImage(src, url, width, maxHeight)
}

// RenderImage turns a decoded image into block lines that display the photo
// natively, fitted to width cells and capped to maxHeight rows (a height cap
// narrows the box to w cells). The first line carries the escape sequence
// (zero visible width) followed by padding to the image's cell width w; each
// subsequent reserved row is a blank line of the same width, so the block
// spans exactly the image's cell box and the caller centers it within the
// content width like a halfblock block. url names the image and the render's
// pixel dimensions key the kitty image id, so each distinct render size is a
// distinct terminal image rather than a same-id re-transmit.
func (r NativeRenderer) RenderImage(src image.Image, url string, width, maxHeight int) ([]string, error) {
	if r.Protocol == ProtocolNone {
		return nil, fmt.Errorf("no native image protocol")
	}
	b := src.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		return nil, fmt.Errorf("image has no pixels")
	}
	w, h := fitDims(b.Dx(), b.Dy(), renderWidth(b.Dx(), b.Dy(), width), maxHeight)
	pxW, pxH := r.pixelDims(w, h)
	scaled := scaleTo(src, pxW, pxH)
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}
	payload := base64.StdEncoding.EncodeToString(buf.Bytes())
	esc := r.escape(url, renderID(url, pxW, pxH), w, h, buf.Len(), payload)
	lines := make([]string, max(1, h))
	lines[0] = esc + strings.Repeat(" ", w)
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", w)
	}
	return lines, nil
}

// maxNativeEdge caps the longer edge of a native render's encoded pixel
// dimensions. It is a fixed constant independent of the terminal's reported
// cell pixel size and well below the historical 4096px absolute cap, so a
// full-width photo's inline-image payload stays bounded on any terminal while
// remaining sharp at typical cell-pixel widths.
const maxNativeEdge = 1280

// pixelDims computes the pixel size the image should be encoded at: the
// on-screen cell box (w cells × h rows at the terminal's cell resolution), so
// the terminal displays it sharply rather than upscaling a low-res source. The
// longest edge is capped at maxNativeEdge, scaling the shorter edge to
// preserve aspect ratio, so an oversized cell box cannot balloon the payload;
// the terminal scales the encoded render back up to the full cell box, so the
// displayed photo is unchanged.
func (r NativeRenderer) pixelDims(w, h int) (pxW, pxH int) {
	cw, ch := cellPixelSizeFn()
	pxW, pxH = w*cw, h*ch
	if pxW > maxNativeEdge || pxH > maxNativeEdge {
		if pxW >= pxH {
			pxH = pxH * maxNativeEdge / pxW
			pxW = maxNativeEdge
		} else {
			pxW = pxW * maxNativeEdge / pxH
			pxH = maxNativeEdge
		}
	}
	return pxW, pxH
}

// escape builds the protocol-specific sequence embedding the PNG payload. The
// render id keys the kitty image so each distinct render size is independently
// addressable and deletable by the terminal.
func (r NativeRenderer) escape(url string, id uint32, w, h, size int, payload string) string {
	switch r.Protocol {
	case ProtocolITerm:
		name := base64.StdEncoding.EncodeToString([]byte("yerss-" + stableName(url)))
		return fmt.Sprintf("\x1b]1337;File=name=%s;size=%d;inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\x07",
			name, size, w, h, payload)
	case ProtocolKitty:
		return kittyTransmit(id, w, h, payload)
	}
	return ""
}

// Clear returns the protocol sequence that removes the terminal's native
// image placements from the screen, or "" when the protocol needs no
// explicit clear. In the kitty graphics protocol, placements float above
// text and the erase commands a frame repaint issues have no effect on
// them, so a placement made by one frame survives into later frames unless
// explicitly deleted; the delete action with d=a removes every visible
// placement. OSC 1337 inline images are cell-bound, so overwriting their
// cells erases them with no help. q=1 suppresses the OK response, matching
// the transmit.
func (r NativeRenderer) Clear() string {
	if r.Protocol != ProtocolKitty {
		return ""
	}
	return "\x1b_Ga=d,d=a,q=1\x1b\\"
}

// DeleteByID returns the kitty sequence that deletes the image (and its
// placements) transmitted under id, freeing the terminal's cached image data
// via the data-freeing delete action (d=I). Unlike Clear's d=a, which only
// removes placements visible on screen, deleting by id also releases the
// terminal's cached image data, so an image scrolled out of view or superseded
// by a re-render cannot linger in the terminal's image cache. It returns "" for
// id 0 (nothing was transmitted). q=1 suppresses the response.
func DeleteByID(id uint32) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("\x1b_Ga=d,d=I,i=%d,q=1\x1b\\", id)
}

// kittyChunkSize bounds a single kitty graphics chunk in base64 characters.
// The protocol requires remote clients to chunk the base64 payload into pieces
// no larger than 4096 bytes; terminals such as Ghostty reject oversized single
// APC transmissions.
const kittyChunkSize = 4096

// kittyTransmit builds the kitty graphics transmit sequence for payload (the
// base64-encoded PNG), displaying it scaled to w×h cells via the c and r keys
// and leaving the cursor in place via C=1 so the caller's reserved padding
// lines align with the image box. The payload is split into kittyChunkSize
// chunks; each chunk is its own APC sequence and the m key marks continuation
// (1 for all but the last chunk, 0 for the last). Both the image id (i) and
// the placement id (p) are stable so repeated transmissions replace the
// terminal's cached image and placement rather than accumulating new ones
// (Ghostty accumulates anonymous placements when p is omitted). q=1 suppresses
// OK responses so the terminal does not echo an acknowledgement into the app's
// stdin (which bubbletea would read as an ESC and bounce the view).
func kittyTransmit(id uint32, w, h int, payload string) string {
	head := fmt.Sprintf("\x1b_Ga=T,f=100,q=1,i=%d,p=%d,c=%d,r=%d,C=1", id, id, w, h)
	if len(payload) <= kittyChunkSize {
		return fmt.Sprintf("%s;%s\x1b\\", head, payload)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s,m=1;%s\x1b\\", head, payload[:kittyChunkSize])
	payload = payload[kittyChunkSize:]
	for len(payload) > kittyChunkSize {
		fmt.Fprintf(&b, "\x1b_Gm=1;%s\x1b\\", payload[:kittyChunkSize])
		payload = payload[kittyChunkSize:]
	}
	fmt.Fprintf(&b, "\x1b_Gm=0;%s\x1b\\", payload)
	return b.String()
}

// kittyPlacementRef builds the kitty graphics placement-reference sequence that
// re-shows the image transmitted under id, scaled to the w×h cell box, without
// re-transmitting its payload: the terminal re-places the cached image data.
// It carries the same stable image id (i) and placement id (p) as the full
// transmit, so identity tracking and delete-by-id cleanup are unchanged, and
// q=1 suppresses the OK response like the transmit.
func kittyPlacementRef(id uint32, w, h int) string {
	return fmt.Sprintf("\x1b_Ga=p,i=%d,p=%d,c=%d,r=%d,C=1,q=1\x1b\\", id, id, w, h)
}

// kittyTransmitBox reads the display box (c,r) out of a line's kitty transmit
// or placement-reference escape, returning 0,0 when the line carries no kitty
// graphics escape.
func kittyTransmitBox(lines []string) (w, h int) {
	if len(lines) == 0 {
		return 0, 0
	}
	idx := strings.Index(lines[0], "\x1b_G")
	if idx < 0 {
		return 0, 0
	}
	ctl := lines[0][idx:]
	if end := strings.IndexByte(ctl, ';'); end >= 0 {
		ctl = ctl[:end]
	} else if end := strings.Index(ctl, "\x1b\\"); end >= 0 {
		ctl = ctl[:end]
	}
	for _, kv := range strings.Split(ctl, ",") {
		switch {
		case strings.HasPrefix(kv, "c="):
			n, err := strconv.Atoi(kv[2:])
			if err != nil {
				return 0, 0
			}
			w = n
		case strings.HasPrefix(kv, "r="):
			n, err := strconv.Atoi(kv[2:])
			if err != nil {
				return 0, 0
			}
			h = n
		}
	}
	return w, h
}

// KittyPlacementReference returns line with the kitty transmit escape that
// opened it — including every continuation chunk of a chunked payload —
// replaced by a placement reference naming the same image id (the terminal
// re-shows the already-cached image without a payload re-transmit). The display
// box is carried across from the transmit, so the re-shown photo occupies the
// same cells. It returns line unchanged when it carries no kitty transmit
// escape.
func KittyPlacementReference(line string) string {
	idx := strings.Index(line, "\x1b_G")
	if idx < 0 {
		return line
	}
	rest := line[idx:]
	end := strings.LastIndex(rest, "\x1b\\")
	if end < 0 {
		return line
	}
	w, h := kittyTransmitBox([]string{line})
	id := NativeRenderID([]string{line})
	if id == 0 || w == 0 || h == 0 {
		return line
	}
	return line[:idx] + kittyPlacementRef(id, w, h) + rest[end+len("\x1b\\"):]
}

// stableName returns a stable non-empty identifier for a URL.
func stableName(url string) string {
	if url == "" {
		return "image"
	}
	return url
}

// renderID hashes the URL and the render's pixel dimensions into a kitty
// image/placement id, never zero (the protocol reserves image id 0). The
// payload is re-encoded at the terminal's current cell geometry, so a resize
// transmits different pixel data; keying the id by those dimensions makes each
// distinct render a distinct image the terminal can replace and delete
// independently, rather than re-transmitting different data under a same id
// that a terminal may accumulate without bound.
func renderID(url string, pxW, pxH int) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	var dims [8]byte
	binary.LittleEndian.PutUint32(dims[0:4], uint32(pxW))
	binary.LittleEndian.PutUint32(dims[4:8], uint32(pxH))
	_, _ = h.Write(dims[:])
	id := h.Sum32()
	if id == 0 {
		return 1
	}
	return id
}

// NativeRenderID returns the kitty image id a native render was transmitted
// under, read back out of its own escape sequence — the single source of truth
// for what the terminal received — so cleanup can delete the exact image. It
// returns 0 when the lines carry no kitty transmit (an iTerm OSC 1337 render or
// a halfblock block).
func NativeRenderID(lines []string) uint32 {
	if len(lines) == 0 {
		return 0
	}
	idx := strings.Index(lines[0], "\x1b_G")
	if idx < 0 {
		return 0
	}
	ctl := lines[0][idx:]
	if end := strings.IndexByte(ctl, ';'); end >= 0 {
		ctl = ctl[:end]
	} else if end := strings.Index(ctl, "\x1b\\"); end >= 0 {
		ctl = ctl[:end]
	}
	for _, kv := range strings.Split(ctl, ",") {
		if strings.HasPrefix(kv, "i=") {
			n, err := strconv.ParseUint(kv[2:], 10, 32)
			if err != nil {
				return 0
			}
			return uint32(n)
		}
	}
	return 0
}

// scaleTo resizes src to w×h pixels, returning src unchanged when it already
// matches.
func scaleTo(src image.Image, w, h int) image.Image {
	b := src.Bounds()
	if b.Dx() == w && b.Dy() == h {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

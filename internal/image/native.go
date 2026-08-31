package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"image"
	"image/png"
	"os"
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
// narrows the box to w cells). The first line carries the escape sequence
// (zero visible width) followed by padding to the image's cell width w; each
// subsequent reserved row is a blank line of the same width, so the block
// spans exactly the image's cell box and the caller centers it within the
// content width like a halfblock block. url gives the image a stable
// terminal-side name/id so repeated renders are reused by the terminal rather
// than re-decoded.
func (r NativeRenderer) Render(data []byte, url string, width, maxHeight int) ([]string, error) {
	if r.Protocol == ProtocolNone {
		return nil, fmt.Errorf("no native image protocol")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
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
	esc := r.escape(url, w, h, buf.Len(), payload)
	lines := make([]string, max(1, h))
	lines[0] = esc + strings.Repeat(" ", w)
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", w)
	}
	return lines, nil
}

// pixelDims computes the pixel size the image should be encoded at: the
// on-screen cell box (w cells × h rows at the terminal's cell resolution), so
// the terminal displays it sharply rather than upscaling a low-res source. The
// longest edge is capped so an oversized image cannot balloon the payload.
func (r NativeRenderer) pixelDims(w, h int) (pxW, pxH int) {
	cw, ch := cellPixelSizeFn()
	pxW, pxH = w*cw, h*ch
	const maxEdge = 4096
	if pxW > maxEdge || pxH > maxEdge {
		if pxW >= pxH {
			pxH = pxH * maxEdge / pxW
			pxW = maxEdge
		} else {
			pxW = pxW * maxEdge / pxH
			pxH = maxEdge
		}
	}
	return pxW, pxH
}

// escape builds the protocol-specific sequence embedding the PNG payload. The
// url-derived name/id keeps the image stable across re-renders.
func (r NativeRenderer) escape(url string, w, h, size int, payload string) string {
	switch r.Protocol {
	case ProtocolITerm:
		name := base64.StdEncoding.EncodeToString([]byte("yerss-" + stableName(url)))
		return fmt.Sprintf("\x1b]1337;File=name=%s;size=%d;inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\x07",
			name, size, w, h, payload)
	case ProtocolKitty:
		return kittyTransmit(stableID(url), w, h, payload)
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

// DeletePlacement returns the kitty sequence that deletes the placement (and
// its image) for url by its stable id. Unlike Clear's d=a, which only removes
// placements visible on screen, deleting by id removes the image's placement
// wherever it is, so an image scrolled out of the article still has its stale
// placement removed. It returns "" for non-kitty protocols. q=1 suppresses the
// response.
func DeletePlacement(url string) string {
	if stableID(url) == 0 {
		return ""
	}
	return fmt.Sprintf("\x1b_Ga=d,d=i,i=%d,q=1\x1b\\", stableID(url))
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

// stableName returns a stable non-empty identifier for a URL.
func stableName(url string) string {
	if url == "" {
		return "image"
	}
	return url
}

// stableID hashes a URL into a kitty image/placement id, never zero (the
// protocol reserves image id 0).
func stableID(url string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	id := h.Sum32()
	if id == 0 {
		return 1
	}
	return id
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

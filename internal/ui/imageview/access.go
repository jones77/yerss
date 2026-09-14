package imageview

import tea "github.com/charmbracelet/bubbletea"

// Loading reports whether url has an image load or native render in flight.
func (v *ImageView) Loading(url string) bool { return v.loading[url] }

// NativePending reports whether url's photo render is deferred until its block
// scrolls into view.
func (v *ImageView) NativePending(url string) bool { return v.nativePending[url] }

// NativeSent returns the kitty image id last transmitted for url, and whether
// one is recorded.
func (v *ImageView) NativeSent(url string) (uint32, bool) {
	id, ok := v.nativeSent[url]
	return id, ok
}

// NativeSentSnapshot returns a copy of the transmitted-id tracking, for
// diagnostics.
func (v *ImageView) NativeSentSnapshot() map[string]uint32 {
	out := make(map[string]uint32, len(v.nativeSent))
	for url, id := range v.nativeSent {
		out[url] = id
	}
	return out
}

// SetNativeSent records url as transmitted under id (test hook).
func (v *ImageView) SetNativeSent(url string, id uint32) { v.nativeSent[url] = id }

// SemCapacity reports the concurrency bound of the image command semaphore.
func (v *ImageView) SemCapacity() int { return cap(v.sem) }

// GateCmd wraps an image command with the shared concurrency bound.
func (v *ImageView) GateCmd(cmd tea.Cmd) tea.Cmd { return v.gateCmd(cmd) }

// ComputeNativeClear computes the native-image clear sequence for the current
// frame and advances the nativeSent/nativeClearArticleID tracking. It is the
// mutation half of RecomputeNativeClear, exposed for tests.
func (v *ImageView) ComputeNativeClear(active bool, view ArticleView) string {
	return v.computeNativeClear(active, view)
}

// InlineAttrFor returns the composed caption for an inline image by URL.
func (v *ImageView) InlineAttrFor(url string, view ArticleView) string {
	return v.inlineAttrFor(url, view)
}
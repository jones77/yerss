package image

import (
	"os"
	"strings"
)

// Protocol identifies a terminal native-image protocol.
type Protocol int

const (
	// ProtocolNone means the terminal has no native-image support and lead
	// images render as halfblock blocks only.
	ProtocolNone Protocol = iota
	// ProtocolITerm is the iTerm2 OSC 1337 inline-image protocol, also
	// supported by WezTerm and Konsole.
	ProtocolITerm
	// ProtocolKitty is the kitty graphics protocol, also supported by
	// Ghostty and WezTerm.
	ProtocolKitty
)

// String returns a short protocol label.
func (p Protocol) String() string {
	switch p {
	case ProtocolITerm:
		return "iterm"
	case ProtocolKitty:
		return "kitty"
	}
	return "none"
}

// detectProtocol is a variable so tests can override the environment probe.
var detectProtocol = detectEnvProtocol

// DetectProtocol reports the terminal's native-image protocol, inferred from
// the environment the same way ASCII fallback and light-background detection
// work. Full-image-capable terminals are iTerm2/WezTerm (OSC 1337 inline
// images) and kitty-family terminals (kitty graphics protocol). Under tmux the
// protocol is disabled unless passthrough is explicitly allowed, since inline
// images would otherwise be captured by the multiplexer.
func DetectProtocol() Protocol { return detectProtocol() }

// detectGhostty is a variable so tests can override the environment probe.
var detectGhostty = detectEnvGhostty

// DetectGhostty reports whether the terminal is Ghostty, inferred from the
// environment the same way DetectProtocol works (TERM_PROGRAM is "ghostty" or
// GHOSTTY_RESOURCES_DIR is set). Ghostty's kitty-graphics delete handling is
// partial, so the cleanup path adds a delete-all fallback on it.
func DetectGhostty() bool { return detectGhostty() }

func detectEnvGhostty() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty" || os.Getenv("GHOSTTY_RESOURCES_DIR") != ""
}

func detectEnvProtocol() Protocol {
	if os.Getenv("TMUX") != "" && os.Getenv("TMUX_ALLOW_PASSTHROUGH") != "1" {
		return ProtocolNone
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm":
		return ProtocolITerm
	case "kitty", "ghostty":
		return ProtocolKitty
	}
	term := os.Getenv("TERM")
	if strings.Contains(term, "kitty") || strings.Contains(term, "ghostty") {
		return ProtocolKitty
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" || os.Getenv("KITTY_PID") != "" ||
		os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return ProtocolKitty
	}
	return ProtocolNone
}
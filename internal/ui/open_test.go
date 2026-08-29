package ui

import (
	"runtime"
	"strings"
	"testing"
)

func TestOpenURLCmdArgumentSafety(t *testing.T) {
	url := "https://example.com/a&calc" // metacharacters cmd.exe would interpret
	cmd := openCommand(url)

	// The URL is always the final argv element, passed as a single argument
	// rather than being shell-split.
	if got := cmd.Args[len(cmd.Args)-1]; got != url {
		t.Fatalf("URL not passed as a single argument: args=%q", cmd.Args)
	}
	// No shell form is ever used with the raw URL.
	if strings.Contains(strings.Join(cmd.Args, " "), "/c") {
		t.Fatalf("shell form used with raw URL: args=%q", cmd.Args)
	}

	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "rundll32" {
			t.Errorf("windows opener should be rundll32, got %q", cmd.Args[0])
		}
		if cmd.Args[1] != "url.dll,FileProtocolHandler" {
			t.Errorf("windows opener missing protocol handler: %q", cmd.Args)
		}
	case "darwin":
		if cmd.Args[0] != "open" {
			t.Errorf("darwin opener should be open, got %q", cmd.Args[0])
		}
	default:
		if cmd.Args[0] != "xdg-open" {
			t.Errorf("linux opener should be xdg-open, got %q", cmd.Args[0])
		}
	}
}
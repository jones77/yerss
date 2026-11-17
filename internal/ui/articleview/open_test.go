package articleview

import (
	"errors"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"yerss/internal"
)

func TestDebugInfoSnapshot(t *testing.T) {
	s, _ := newTestState(t)
	s.width = 106
	s.height = 31
	s.Viewport.YOffset = 7
	s.lines = make([]string, 200)
	s.ascii = true
	s.LinkExpand = true
	s.article = &internal.Article{ID: 42, Title: "Hello World", Link: "https://example.com/x"}

	got := debugInfoSnapshot(&s)
	for _, want := range []string{
		"size: 106x31",
		"scroll: 7/200",
		"id: 42",
		"title: Hello World",
		"url: https://example.com/x",
		"ascii: true",
		"link_expand: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("snapshot missing %q:\n%s", want, got)
		}
	}
}

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

func cmdPaths(cmds []*exec.Cmd) []string {
	names := make([]string, len(cmds))
	for i, c := range cmds {
		names[i] = strings.Join(c.Args, " ")
	}
	return names
}

func TestClipboardCmdsLinuxChain(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only helper chain")
	}
	orig := wslInterop
	defer func() { wslInterop = orig }()

	t.Setenv("WAYLAND_DISPLAY", "")
	wslInterop = func() bool { return false }
	if got := cmdPaths(clipboardCmds()); !reflect.DeepEqual(got, []string{"xclip -selection clipboard"}) {
		t.Errorf("x11 helpers = %v", got)
	}

	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	wslInterop = func() bool { return false }
	if got := cmdPaths(clipboardCmds()); !reflect.DeepEqual(got, []string{"wl-copy", "xclip -selection clipboard"}) {
		t.Errorf("wayland helpers = %v", got)
	}

	t.Setenv("WAYLAND_DISPLAY", "")
	wslInterop = func() bool { return true }
	if got := cmdPaths(clipboardCmds()); !reflect.DeepEqual(got, []string{"xclip -selection clipboard", "clip.exe"}) {
		t.Errorf("wsl helpers = %v", got)
	}
}

func TestClipboardHelperPicksFirstAvailable(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(name string) (string, error) {
		if name == "xclip" {
			return "/usr/bin/xclip", nil
		}
		return "", errors.New("not found")
	}
	cmd, _ := clipboardHelper([]*exec.Cmd{
		exec.Command("wl-copy"),
		exec.Command("xclip", "-selection", "clipboard"),
	})
	if cmd == nil {
		t.Fatal("expected the available helper to be picked")
	}
	if got := strings.Join(cmd.Args, " "); got != "xclip -selection clipboard" {
		t.Errorf("picked helper = %q", got)
	}
}

func TestClipboardHelperNoneAvailable(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	cmd, tried := clipboardHelper([]*exec.Cmd{exec.Command("wl-copy"), exec.Command("xclip")})
	if cmd != nil {
		t.Fatalf("want nil helper, got %q", cmd.Path)
	}
	if tried != "wl-copy, xclip" {
		t.Errorf("tried list = %q, want %q", tried, "wl-copy, xclip")
	}
}

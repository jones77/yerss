package articleview

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// URLActionMsg reports the outcome of an open/copy command.
type URLActionMsg struct {
	Action string
	Label  string
	Err    error
}

// openCommand builds the platform command that opens url in the system
// browser. The URL is always passed as a single exec argument and no shell is
// invoked, so feed-controlled metacharacters cannot inject commands.
func openCommand(url string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url)
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return exec.Command("xdg-open", url)
	}
}

// openURLCmd opens url in the system browser via a tea.Cmd.
func OpenURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if err := openCommand(url).Run(); err != nil {
			return URLActionMsg{Action: "open", Err: err}
		}
		return URLActionMsg{Action: "open", Label: "opened URL"}
	}
}

// copyTextCmd copies text to the system clipboard via a tea.Cmd. The text is
// piped to the platform clipboard helper's standard input without a shell, so
// feed-controlled content cannot inject commands. label is the success status
// line (e.g. "copied article text").
func CopyTextCmd(text, label string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "windows":
			cmd = exec.Command("clip")
		default:
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return URLActionMsg{Action: "copy", Err: err}
		}
		return URLActionMsg{Action: "copy", Label: label}
	}
}

// copyURLCmd copies url to the system clipboard via a tea.Cmd.
func CopyURLCmd(url string) tea.Cmd {
	return CopyTextCmd(url, "copied URL")
}

// debugInfoSnapshot assembles the diagnostic block copied by Copy Debug Info.
// The fixed key: value structure captures the terminal geometry, the article's
// scroll position and rendered length, the article identity, and the rendering
// mode flags needed to reproduce layout bugs.
func debugInfoSnapshot(s *State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "app: yerss\n")
	fmt.Fprintf(&b, "size: %dx%d\n", s.width, s.height)
	fmt.Fprintf(&b, "scroll: %d/%d\n", s.Viewport.YOffset, len(s.lines))
	if s.article != nil {
		fmt.Fprintf(&b, "id: %d\n", s.article.ID)
		fmt.Fprintf(&b, "title: %s\n", ansi.Strip(s.article.Title))
		fmt.Fprintf(&b, "url: %s\n", s.article.Link)
	}
	fmt.Fprintf(&b, "ascii: %t\n", s.ascii)
	fmt.Fprintf(&b, "link_expand: %t\n", s.LinkExpand)
	return b.String()
}

// CopyDebugInfoCmd copies a diagnostic snapshot of the article session to the
// system clipboard via a tea.Cmd.
func CopyDebugInfoCmd(s *State) tea.Cmd {
	return CopyTextCmd(debugInfoSnapshot(s), "copied debug info")
}
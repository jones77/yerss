package articleview

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
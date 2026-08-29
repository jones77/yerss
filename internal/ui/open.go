package ui

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if err := openCommand(url).Run(); err != nil {
			return urlActionMsg{action: "open", err: err}
		}
		return urlActionMsg{action: "open"}
	}
}

// copyURLCmd copies url to the system clipboard via a tea.Cmd.
func copyURLCmd(url string) tea.Cmd {
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
		cmd.Stdin = strings.NewReader(url)
		if err := cmd.Run(); err != nil {
			return urlActionMsg{action: "copy", err: err}
		}
		return urlActionMsg{action: "copy"}
	}
}
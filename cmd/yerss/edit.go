package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"yerss/internal/config"
)

// exitNoEditor is the exit code used when neither $EDITOR nor $VISUAL is set,
// distinct from the startup gate's 2/3 refusal codes.
const exitNoEditor = 4

// resolveEditor returns the editor command from $EDITOR, falling back to
// $VISUAL. It reports false when neither variable is set; there is no built-in
// fallback editor.
func resolveEditor() (string, bool) {
	if e := os.Getenv("EDITOR"); e != "" {
		return e, true
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v, true
	}
	return "", false
}

// shellQuote wraps s in single quotes, escaping embedded single quotes, so the
// string survives shell parsing as a single argument.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// runEditor invokes the resolved editor against path through the shell, so
// editor strings that carry arguments (e.g. "code --wait") work. It returns
// the editor's exit code, or a non-zero code when no editor is configured or
// the shell cannot be started.
func runEditor(path string) int {
	editor, ok := resolveEditor()
	if !ok {
		fmt.Fprintln(os.Stderr, "yerss: neither $EDITOR nor $VISUAL is set; set one to use -e/--edit-feeds or -c/--edit-config")
		return exitNoEditor
	}
	cmd := exec.Command("sh", "-c", editor+" "+shellQuote(path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "yerss: cannot start editor %q: %v\n", editor, err)
		return 1
	}
	return 0
}

// runEditConfig opens config.toml at the XDG default path in the editor and
// returns the exit code. The file is ensured to exist first.
func runEditConfig() int {
	return runEditFile(config.DefaultConfigPath())
}

// runEditFeeds resolves the feeds file path, honoring a [data] feeds_file
// override when the config parses and falling back to the XDG default
// otherwise, ensures it exists, and opens it in the editor. It returns the
// editor's exit code.
func runEditFeeds() int {
	path := config.DefaultFeedsPath()
	if cfg, err := config.Load(""); err == nil {
		path = cfg.FeedsFile()
	}
	return runEditFile(path)
}

// runEditFile ensures path exists (via the config bootstrap, which seeds any
// absent file) and opens it in the editor.
func runEditFile(path string) int {
	if err := config.Bootstrap(config.DefaultConfigPath(), path); err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot ensure %s exists: %v\n", path, err)
		return 1
	}
	return runEditor(path)
}
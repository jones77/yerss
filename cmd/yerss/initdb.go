package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"yerss/internal/store"
	"yerss/internal/textutil"
)

// stdinReader returns the stream confirmInitDB reads a confirmation from. It
// is a package variable so tests can substitute a stub.
var stdinReader = func() io.Reader { return os.Stdin }

// stdinIsTTY reports whether standard input is attached to a terminal. It is
// a package variable so tests can force either branch.
var stdinIsTTY = func() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

// confirmInitDB prompts before --init-db deletes the database: it reports the
// on-disk size and article count, then asks for a y/N confirmation on standard
// input. It returns an exit code and ok=false to refuse deletion, or ok=true to
// proceed. A missing database file is not an error — there is nothing to lose —
// so it returns ok=true without prompting. A non-terminal standard input cannot
// be confirmed, so it refuses with exit code 1; a declined answer refuses with
// exit code 0 (the user chose not to delete).
func confirmInitDB(path string) (code int, ok bool) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, true
	}
	size := int64(0)
	if err == nil {
		size = fi.Size()
	}

	count := 0
	if st, oerr := store.Open(path); oerr == nil {
		count, _ = st.ArticleCount()
		st.Close()
	}

	if !stdinIsTTY() {
		fmt.Fprintf(os.Stderr, "yerss: -i requires a terminal; refusing to delete %s\n", path)
		return 1, false
	}

	fmt.Fprintf(os.Stderr, "delete database (%s, %d articles) at %s? [y/N] ", textutil.FormatSize(size), count, path)
	line, _ := bufio.NewReader(stdinReader()).ReadString('\n')
	if strings.EqualFold(strings.TrimSpace(line), "y") {
		return 0, true
	}
	fmt.Fprintf(os.Stderr, "yerss: not deleting %s\n", path)
	return 0, false
}

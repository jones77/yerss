package feed

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadFeeds reads feed entries from a plain-text file, one per line. Blank
// lines and lines starting with '#' are skipped. Entries are scheme-less:
// a leading "https://" or "http://" is stripped from every entry (lookups
// always use HTTPS), and when any entry carried a scheme the file is
// rewritten in place with the canonical scheme-less form, preserving comments
// and blank lines. The rewrite is best-effort; a failure leaves the file
// untouched while the returned entries stay normalized.
func LoadFeeds(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	urls, rawLines, changed, err := readEntries(f)
	f.Close()
	if err != nil {
		return nil, err
	}
	if changed {
		rewriteFeedsFile(path, rawLines)
	}
	return urls, nil
}

// readEntries scans the open feeds file into its raw lines and the
// normalized entries, reporting whether any entry carried a scheme.
func readEntries(f *os.File) (urls []string, rawLines []string, changed bool, err error) {
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		raw := sc.Text()
		rawLines = append(rawLines, raw)
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		stripped := stripScheme(line)
		if stripped != line {
			changed = true
		}
		urls = append(urls, stripped)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("feeds file read: %w", err)
	}
	return urls, rawLines, changed, nil
}

// rewriteFeedsFile writes the file back with scheme-bearing entries replaced
// by their normalized form, preserving every other line verbatim. Errors are
// ignored: the fetch still works because lookups canonicalize entries.
func rewriteFeedsFile(path string, rawLines []string) {
	out := make([]string, len(rawLines))
	for i, raw := range rawLines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			out[i] = raw
			continue
		}
		out[i] = stripScheme(line)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// stripScheme removes a leading https:// or http:// (case-insensitive) so
// entries are stored scheme-less.
func stripScheme(entry string) string {
	lower := strings.ToLower(entry)
	for _, p := range []string{"https://", "http://"} {
		if strings.HasPrefix(lower, p) {
			return entry[len(p):]
		}
	}
	return entry
}

// CanonicalURL returns the URL a feeds.txt entry is fetched from and stored
// under. Entries are scheme-less and always looked up over HTTPS; an entry
// that already carries a scheme (which LoadFeeds strips, but callers may
// pass raw URLs) is used as-is.
func CanonicalURL(entry string) string {
	if strings.Contains(entry, "://") {
		return entry
	}
	return "https://" + entry
}

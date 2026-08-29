package feed

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadFeeds reads feed URLs from a plain-text file, one URL per line.
// Blank lines and lines starting with '#' are skipped.
func LoadFeeds(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("feeds file %s: %w", path, err)
	}
	return urls, nil
}
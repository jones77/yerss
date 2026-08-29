package ui

import "fmt"

// formatSize renders a byte count with a binary unit suffix: whole bytes for
// values under 1 KB, otherwise one decimal place (e.g. `450 B`, `1.4 MB`).
func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	size := float64(bytes)
	for _, u := range []string{"KB", "MB", "GB", "TB"} {
		size /= 1024
		if size < 1024 || u == "TB" {
			return fmt.Sprintf("%.1f %s", size, u)
		}
	}
	return fmt.Sprintf("%.1f TB", size)
}
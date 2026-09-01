package ui

import "fmt"

// formatMB renders a byte count as a floored whole number of megabytes with no
// space between the number and the MB suffix (e.g. "45MB"). Values under one
// megabyte render as "0MB". It is the status bar's formatter for the session's
// decoded-image footprint and the on-disk database size; it deliberately
// differs from textutil.FormatSize, which the --init-db prompt still uses.
func formatMB(bytes int64) string {
	return fmt.Sprintf("%dMB", bytes>>20)
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// feedsSeed is the placeholder content written to feeds.txt on first run.
// Lines starting with '#' are ignored by the feed loader, so the example feed
// stays inert until the user uncomments it.
const feedsSeed = `# yerss feeds — one RSS URL per line. Lines starting with # are ignored.
# Uncomment the line below to add Drop Site News as an example feed.
# https://www.dropsitenews.com/feed
`

// seededConfig renders the full commented config template written on first
// run. Every supported option appears commented out, grouped by section, with
// a header recording the current schema version and build commit. It is
// drift-checked against ConfigReleases by a test.
func seededConfig() string {
	return fmt.Sprintf(`# yerss config — schema v%d (commit: %s)
#
# All options below are commented out; built-in defaults are used until you
# uncomment and set a value. Feed URLs go in %s, one per line.

# Data directory holding the SQLite database.
# Leave empty to use $XDG_DATA_HOME/yerss.
# [data]
# dir = ""
# feeds_file = ""

# [refresh]
# min_interval = "15m"
# cooldown = "60s"

# [display]
# theme = "auto"
# ascii = false
# padding_x = 2
# padding_y = 1

%s
`, ConfigSchemaVersion, BuildCommit, defaultFeedsFile, seededKeybindings())
}

// Bootstrap creates the config directory and seeds config.toml and feeds.txt
// only when each file is absent. Existing files are never overwritten or
// truncated. An error is returned when the directory cannot be created or a
// seed file cannot be written; callers treat this as non-fatal because
// built-in defaults require no config file to function.
func Bootstrap(configPath, feedsPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(seededConfig()), 0o644); err != nil {
			return err
		}
	}
	if _, err := os.Stat(feedsPath); os.IsNotExist(err) {
		if err := os.WriteFile(feedsPath, []byte(feedsSeed), 0o644); err != nil {
			return err
		}
	}
	return nil
}
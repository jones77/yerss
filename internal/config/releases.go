package config

// BuildCommit is the VCS revision embedded at build time via
// -ldflags "-X yerss/internal/config.BuildCommit=<sha>". It is display-only
// in the seeded config header and never used for control flow; it falls back
// to "dev" under plain `go run`.
var BuildCommit = "dev"

// ConfigSchemaVersion is a monotonically increasing integer that increments
// whenever a configuration option is added. It drives the config self-update:
// files recording an older schema version in their header have the newer
// options appended as comments on the next load.
const ConfigSchemaVersion = 1

// ConfigOption records a single configuration option and the schema version
// that introduced it.
type ConfigOption struct {
	Version int    // schema version that introduced the option
	Section string // TOML section name
	Key     string // option key within the section
	Default string // rendered default value (TOML literal)
	Doc     string // one-line description
}

// ConfigReleases is the running tab of configuration options introduced per
// schema version. It is the source of truth for the config self-update and is
// drift-checked against the seeded template by a test. The non-keybinding
// options are listed explicitly; the keybinding options are derived from the
// action catalog (see keys.go) so the registry cannot drift from the runtime
// defaults.
var ConfigReleases = allConfigOptions()

func allConfigOptions() []ConfigOption {
	opts := []ConfigOption{
		// v1 baseline.
		{Version: 1, Section: "data", Key: "dir", Default: `""`, Doc: "data directory for the SQLite database"},
		{Version: 1, Section: "data", Key: "feeds_file", Default: `""`, Doc: "path to the feeds file"},
		{Version: 1, Section: "refresh", Key: "min_interval", Default: `"15m"`, Doc: "startup refresh gate interval"},
		{Version: 1, Section: "refresh", Key: "cooldown", Default: `"60s"`, Doc: "manual refresh cooldown"},
		{Version: 1, Section: "display", Key: "theme", Default: `"auto"`, Doc: "theme: auto, dark, or light"},
		{Version: 1, Section: "display", Key: "ascii", Default: `false`, Doc: "ASCII border fallback"},
		{Version: 1, Section: "display", Key: "padding_x", Default: `2`, Doc: "horizontal article padding"},
		{Version: 1, Section: "display", Key: "padding_y", Default: `1`, Doc: "vertical article padding"},
	}
	return append(opts, keybindingOptions()...)
}
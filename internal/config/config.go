package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

const (
	defaultConfigFile = "config.toml"
	defaultFeedsFile  = "feeds.txt"
	defaultDBFile     = "yerss.sqlite"
)

// Data holds file path overrides. Empty values fall back to XDG defaults.
type Data struct {
	Dir       string `toml:"dir"`
	FeedsFile string `toml:"feeds_file"`
}

// RefreshOptions holds the startup gate and manual refresh cooldown durations.
type RefreshOptions struct {
	MinInterval string `toml:"min_interval"`
	Cooldown    string `toml:"cooldown"`
}

// Display holds rendering preferences.
type Display struct {
	Theme     string
	Ascii     bool
	Images    string
	PaddingX  int
	PaddingY  int
	Scrollbar string
}

// Config is the fully-resolved application configuration.
type Config struct {
	Data        Data
	Refresh     RefreshOptions
	Display     Display
	Keybindings Keymap
}

// Default returns a Config populated with built-in defaults.
func Default() *Config {
	return &Config{
		Refresh:     RefreshOptions{MinInterval: "15m", Cooldown: "60s"},
		Display:     Display{Theme: "auto", Images: "auto", PaddingX: 2, PaddingY: 1, Scrollbar: "single"},
		Keybindings: DefaultKeybindings(),
	}
}

// DataDir returns the resolved data directory.
func (c *Config) DataDir() string {
	if c.Data.Dir != "" {
		return c.Data.Dir
	}
	return filepath.Join(xdg.DataHome, "yerss")
}

// DBPath returns the resolved SQLite database path.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir(), defaultDBFile)
}

// FeedsFile returns the resolved feeds file path.
func (c *Config) FeedsFile() string {
	if c.Data.FeedsFile != "" {
		return c.Data.FeedsFile
	}
	return DefaultFeedsPath()
}

// DefaultConfigPath returns the XDG default config file path, without parsing
// any config. Used when the config file itself may be the thing being fixed.
func DefaultConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "yerss", defaultConfigFile)
}

// DefaultFeedsPath returns the XDG default feeds file path, without parsing
// any config.
func DefaultFeedsPath() string {
	return filepath.Join(xdg.ConfigHome, "yerss", defaultFeedsFile)
}

// MinInterval returns the parsed startup refresh gate duration.
func (c *Config) MinInterval() time.Duration {
	d, err := time.ParseDuration(c.Refresh.MinInterval)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

// Cooldown returns the parsed manual refresh cooldown duration.
func (c *Config) Cooldown() time.Duration {
	d, err := time.ParseDuration(c.Refresh.Cooldown)
	if err != nil {
		return 60 * time.Second
	}
	return d
}

// normalizeImages maps an images config value to the supported set
// (auto/on/off), treating any unrecognized value as the auto default.
func normalizeImages(v string) string {
	switch v {
	case "auto", "on", "off":
		return v
	}
	return "auto"
}

// normalizeScrollbar maps a scrollbar config value to the supported set
// (single/double), treating any unrecognized value as the single default.
func normalizeScrollbar(v string) string {
	switch v {
	case "single", "double":
		return v
	}
	return "single"
}

type displayOpts struct {
	Theme     string `toml:"theme"`
	Images    string `toml:"images"`
	PaddingX  *int   `toml:"padding_x"`
	PaddingY  *int   `toml:"padding_y"`
	Ascii     *bool  `toml:"ascii"`
	Scrollbar *string `toml:"scrollbar"`
}

// Load reads configuration from path, falling back to the default XDG
// location when path is empty and to built-in defaults when the file is
// missing. It returns an error if the TOML is invalid or keybindings
// conflict.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}
	cfg := Default()

	// First-run bootstrap and schema self-update are best-effort: built-in
	// defaults require no config file, so failures here never abort startup.
	_ = Bootstrap(path, filepath.Join(filepath.Dir(path), defaultFeedsFile))
	_ = MigrateFile(path)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("config %s: %w", path, err)
	}

	var raw struct {
		Data        Data                    `toml:"data"`
		Refresh     RefreshOptions          `toml:"refresh"`
		Display     displayOpts             `toml:"display"`
		Keybindings map[string][]string     `toml:"keybindings"`
	}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("config %s: %w", path, err)
	}

	if raw.Data.Dir != "" {
		cfg.Data.Dir = raw.Data.Dir
	}
	if raw.Data.FeedsFile != "" {
		cfg.Data.FeedsFile = raw.Data.FeedsFile
	}
	if raw.Refresh.MinInterval != "" {
		cfg.Refresh.MinInterval = raw.Refresh.MinInterval
	}
	if raw.Refresh.Cooldown != "" {
		cfg.Refresh.Cooldown = raw.Refresh.Cooldown
	}
	if raw.Display.Theme != "" {
		cfg.Display.Theme = raw.Display.Theme
	}
	if raw.Display.Images != "" {
		cfg.Display.Images = normalizeImages(raw.Display.Images)
	}
	if raw.Display.PaddingX != nil {
		cfg.Display.PaddingX = *raw.Display.PaddingX
	}
	if raw.Display.PaddingY != nil {
		cfg.Display.PaddingY = *raw.Display.PaddingY
	}
	if raw.Display.Ascii != nil {
		cfg.Display.Ascii = *raw.Display.Ascii
	}
	if raw.Display.Scrollbar != nil {
		cfg.Display.Scrollbar = normalizeScrollbar(*raw.Display.Scrollbar)
	}

	if len(raw.Keybindings) > 0 {
		for action, keys := range raw.Keybindings {
			cfg.Keybindings[Action(action)] = keys
		}
	}

	parsed, err := ParseKeybindings(cfg.Keybindings)
	if err != nil {
		return nil, err
	}
	cfg.Keybindings = parsed
	return cfg, nil
}
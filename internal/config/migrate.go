package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// schemaHeaderRe matches the schema-version header written by the bootstrap
// template: "# yerss config — schema vN (commit: X)".
var schemaHeaderRe = regexp.MustCompile(`(?m)^# yerss config — schema v(\d+)`)

// schemaVersionRe matches the version digits in that header for replacement.
var schemaVersionRe = regexp.MustCompile(`(?m)^(# yerss config — schema v)\d+`)

// fileSchemaVersion extracts the schema version from a config file's header.
// It returns false when the file has no parseable schema-version header.
func fileSchemaVersion(content []byte) (int, bool) {
	m := schemaHeaderRe.FindSubmatch(content)
	if m == nil {
		return 0, false
	}
	v, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return 0, false
	}
	return v, true
}

// MigrateFile best-effort upgrades an existing config.toml in place: options
// introduced after the file's recorded schema version are appended as
// commented-out lines beneath a "new in vN" banner, and the header schema
// version is bumped to the current one. An option is appended only when its
// key does not already appear anywhere in the file (whether commented or
// live), which makes the operation idempotent and robust to aborted writes.
// Files without a parseable schema-version header are left untouched. Errors
// are returned to the caller, which treats them as non-fatal.
func MigrateFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ver, ok := fileSchemaVersion(content)
	if !ok || ver >= ConfigSchemaVersion {
		return nil
	}

	body := string(content)
	var toAppend []ConfigOption
	for _, opt := range ConfigReleases {
		if opt.Version > ver && !strings.Contains(body, opt.Key+" = ") {
			toAppend = append(toAppend, opt)
		}
	}

	if len(toAppend) > 0 {
		var b strings.Builder
		b.WriteString("\n")
		prevVer, prevSection := 0, ""
		for _, opt := range toAppend {
			if opt.Version != prevVer {
				if prevVer != 0 {
					b.WriteString("\n")
				}
				b.WriteString("# new in v" + strconv.Itoa(opt.Version) + "\n")
				prevVer, prevSection = opt.Version, ""
			}
			if opt.Section != prevSection {
				if prevSection != "" {
					b.WriteString("\n")
				}
				b.WriteString("# [" + opt.Section + "]\n")
				prevSection = opt.Section
			}
			b.WriteString("# " + opt.Key + " = " + opt.Default + "\n")
		}
		content = append(content, []byte(b.String())...)
	}

	content = schemaVersionRe.ReplaceAll(content, []byte("${1}"+strconv.Itoa(ConfigSchemaVersion)))
	return os.WriteFile(path, content, 0o644)
}
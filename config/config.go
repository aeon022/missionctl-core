// Package config gives every suite tool the same two directory conventions
// they'd already reinvented independently: a config dir under ~/.config/<tool>
// and a data dir under ~/.local/share/<tool> (mailctl/taskctl use
// ~/Library/Application Support/<tool> instead for their SQLite DB — that's
// a deliberate macOS-native choice, not something this package overrides).
package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir returns ~/.config/<tool>, creating it if it doesn't exist.
func Dir(tool string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", tool)
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// DataDir returns ~/.local/share/<tool>, creating it if it doesn't exist.
func DataDir(tool string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".local", "share", tool)
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// ResolveDir returns override — expanded ("~/..." to the home directory)
// and created if it doesn't exist — when it's non-empty, and DataDir(tool)
// otherwise. shared reports whether override was used, so a caller can
// feed it straight into syncdir.JournalMode: whether a tool's data lives
// somewhere safe to treat as possibly-synced is decided entirely by
// whether the user configured a directory of their own, with no separate
// flag for them to get right.
func ResolveDir(tool, override string) (dir string, shared bool) {
	if override == "" {
		return DataDir(tool), false
	}
	if strings.HasPrefix(override, "~") {
		home, _ := os.UserHomeDir()
		override = filepath.Join(home, strings.TrimPrefix(override, "~"))
	}
	_ = os.MkdirAll(override, 0o755)
	return override, true
}

// Package config gives every suite tool the same two directory conventions
// they'd already reinvented independently: a config dir under ~/.config/<tool>
// and a data dir under ~/.local/share/<tool> (mailctl/taskctl use
// ~/Library/Application Support/<tool> instead for their SQLite DB — that's
// a deliberate macOS-native choice, not something this package overrides).
package config

import (
	"os"
	"path/filepath"
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

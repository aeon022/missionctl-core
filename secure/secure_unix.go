//go:build unix

// Package secure makes a suite tool's local data private. Tools hold mail, finances, notes, health
// and credentials; until now their files were created world-readable (umask 022).
package secure

import (
	"os"
	"path/filepath"
	"syscall"
)

// Init sets umask 077 for this process (everything it creates is owner-only from now on) and tightens
// what earlier versions already created: the tool's config and data directories become 0700 and the
// files directly inside them 0600. User-chosen data_dir folders are left alone (they may be shared).
func Init(tool string) {
	syscall.Umask(0o077)
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	for _, d := range []string{
		filepath.Join(home, ".config", tool),
		filepath.Join(home, ".local", "share", tool),
		filepath.Join(home, "Library", "Application Support", tool),
		filepath.Join(home, ".local", "share", "missionctl"), // shared activity log
	} {
		Tighten(d)
	}
}

// Tighten makes dir 0700 and the regular files directly inside it 0600. Symlinks are skipped.
func Tighten(dir string) {
	fi, err := os.Lstat(dir)
	if err != nil || !fi.IsDir() {
		return
	}
	_ = os.Chmod(dir, 0o700)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Type().IsRegular() {
			_ = os.Chmod(filepath.Join(dir, e.Name()), 0o600)
		}
	}
}

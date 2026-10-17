//go:build unix

package secure

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestInitTightensExistingAndNewFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	data := filepath.Join(home, ".local", "share", "democtl")
	os.MkdirAll(data, 0o755)
	db := filepath.Join(data, "demo.db")
	os.WriteFile(db, []byte("x"), 0o644)

	Init("democtl")

	for p, want := range map[string]os.FileMode{data: 0o700, db: 0o600} {
		fi, err := os.Stat(p)
		if err != nil || fi.Mode().Perm() != want {
			t.Errorf("%s: %v %v, want %o", p, fi, err, want)
		}
	}
	fresh := filepath.Join(home, "new.txt")
	os.WriteFile(fresh, []byte("y"), 0o666)
	if fi, _ := os.Stat(fresh); fi.Mode().Perm() != 0o600 {
		t.Errorf("files created after Init must be owner-only, got %o", fi.Mode().Perm())
	}
}

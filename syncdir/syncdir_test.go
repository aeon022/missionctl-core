package syncdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJournalMode(t *testing.T) {
	if got := JournalMode(false); got != "WAL" {
		t.Errorf("JournalMode(false) = %q, want WAL", got)
	}
	if got := JournalMode(true); got != "DELETE" {
		t.Errorf("JournalMode(true) = %q, want DELETE", got)
	}
}

// A long-lived TUI must not lock a daemon, CLI call or MCP server out: locks are shared.
func TestAcquireIsSharedAndInUseSeesIt(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if InUse(dbPath) {
		t.Fatal("nothing holds the database yet")
	}
	lock1, err := Acquire(dbPath)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	lock2, err := Acquire(dbPath)
	if err != nil {
		t.Fatalf("second Acquire must succeed (shared lock): %v", err)
	}
	if !InUse(dbPath) {
		t.Fatal("InUse must report an open database")
	}
	lock1.Release()
	if !InUse(dbPath) {
		t.Fatal("still open while the second holder lives")
	}
	lock2.Release()
	if InUse(dbPath) {
		t.Fatal("released: no longer in use")
	}
}

func TestReleaseNil(t *testing.T) {
	var l *Lock
	if err := l.Release(); err != nil {
		t.Fatalf("Release on nil Lock: %v", err)
	}
}

func TestICloudPlaceholder(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	if isPlaceholder, _ := ICloudPlaceholder(dbPath); isPlaceholder {
		t.Fatal("expected no placeholder when neither file exists")
	}

	if err := os.WriteFile(dbPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if isPlaceholder, _ := ICloudPlaceholder(dbPath); isPlaceholder {
		t.Fatal("expected no placeholder when the real file exists")
	}
	if err := os.Remove(dbPath); err != nil {
		t.Fatal(err)
	}

	placeholder := filepath.Join(dir, ".test.db.icloud")
	if err := os.WriteFile(placeholder, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	isPlaceholder, path := ICloudPlaceholder(dbPath)
	if !isPlaceholder {
		t.Fatal("expected placeholder to be detected")
	}
	if path != placeholder {
		t.Errorf("placeholderPath = %q, want %q", path, placeholder)
	}
}

func TestAcquireCreatesMissingDataDir(t *testing.T) {
	// first run on a new machine: ~/.local/share/<tool> doesn't exist yet
	db := filepath.Join(t.TempDir(), "does", "not", "exist", "tool.db")
	l, err := Acquire(db)
	if err != nil {
		t.Fatalf("Acquire in a missing directory: %v", err)
	}
	defer l.Release()
	if _, err := os.Stat(db + ".lock"); err != nil {
		t.Errorf("lock file not created: %v", err)
	}
}

package syncdir

import (
	"errors"
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

func TestAcquireContention(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	lock1, err := Acquire(dbPath)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}

	if _, err := Acquire(dbPath); !errors.Is(err, ErrLocked) {
		t.Fatalf("second Acquire: got %v, want ErrLocked", err)
	}

	if err := lock1.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	lock2, err := Acquire(dbPath)
	if err != nil {
		t.Fatalf("Acquire after Release: %v", err)
	}
	if err := lock2.Release(); err != nil {
		t.Fatalf("Release: %v", err)
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

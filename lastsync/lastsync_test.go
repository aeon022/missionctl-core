package lastsync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "last_synced")

	if _, ok := Load(path); ok {
		t.Fatal("expected Load of a nonexistent file to report not-ok")
	}

	want := time.Date(2026, 8, 1, 12, 30, 0, 0, time.UTC)
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}

	got, ok := Load(path)
	if !ok {
		t.Fatal("expected Load to succeed after Save")
	}
	if !got.Equal(want) {
		t.Errorf("Load() = %v, want %v", got, want)
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "last_synced")
	if err := os.WriteFile(path, []byte("not-a-timestamp"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, ok := Load(path); ok {
		t.Fatal("expected Load of a corrupt file to report not-ok")
	}
}

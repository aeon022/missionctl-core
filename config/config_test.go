package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirCreatesAndReturnsPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := Dir("testtool")
	want := filepath.Join(home, ".config", "testtool")
	if dir != want {
		t.Errorf("expected %s, got %s", want, dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected Dir to create the directory, got err=%v", err)
	}
}

func TestDataDirCreatesAndReturnsPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := DataDir("testtool")
	want := filepath.Join(home, ".local", "share", "testtool")
	if dir != want {
		t.Errorf("expected %s, got %s", want, dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected DataDir to create the directory, got err=%v", err)
	}
}

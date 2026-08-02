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

func TestResolveDirFallsBackToDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir, shared := ResolveDir("testtool", "")
	want := filepath.Join(home, ".local", "share", "testtool")
	if dir != want || shared {
		t.Errorf("ResolveDir(_, \"\") = (%s, %v), want (%s, false)", dir, shared, want)
	}
}

func TestResolveDirUsesOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	override := filepath.Join(home, "CustomSyncFolder", "testtool")

	dir, shared := ResolveDir("testtool", override)
	if dir != override || !shared {
		t.Errorf("ResolveDir override = (%s, %v), want (%s, true)", dir, shared, override)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected ResolveDir to create the override directory, got err=%v", err)
	}
}

func TestResolveDirExpandsTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir, shared := ResolveDir("testtool", "~/TildeSync/testtool")
	want := filepath.Join(home, "TildeSync", "testtool")
	if dir != want || !shared {
		t.Errorf("ResolveDir tilde = (%s, %v), want (%s, true)", dir, shared, want)
	}
}

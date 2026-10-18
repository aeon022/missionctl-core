package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplySwitchesPaletteAndNotifies(t *testing.T) {
	orig := Current()
	t.Cleanup(func() { _ = Apply(orig) })

	calls := 0
	OnChange(func() { calls++ })

	if err := Apply("dracula"); err != nil {
		t.Fatal(err)
	}
	if Current() != "dracula" || Blue.Dark != presets["dracula"].Blue.Dark {
		t.Errorf("dracula not applied: current=%s blue=%+v", Current(), Blue)
	}
	blueV2 := BlueV2
	if err := Apply("nord"); err != nil {
		t.Fatal(err)
	}
	if BlueV2 == blueV2 {
		t.Error("the v2 colors must follow the switch")
	}
	if calls != 2 {
		t.Errorf("listeners called %d times, want 2", calls)
	}
	if err := Apply("no-such-theme"); err == nil || Current() != "nord" {
		t.Errorf("an unknown theme must be rejected and leave the palette alone (err=%v current=%s)", err, Current())
	}
	if !contains(Presets(), "gruvbox") || !contains(Presets(), "terminal") {
		t.Errorf("Presets() = %v", Presets())
	}
}

func TestApplyBackToClassicRestoresDefaultsExactly(t *testing.T) {
	orig := Current()
	t.Cleanup(func() { _ = Apply(orig) })
	_ = Apply("dracula")
	_ = Apply("classic")
	if Blue != defaults.Blue || Green != defaults.Green || Red != defaults.Red {
		t.Errorf("classic must reproduce the original palette: blue=%+v want %+v", Blue, defaults.Blue)
	}
}

func TestSavePresetKeepsOverridesAndReplacesTheLine(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".config", "missionctl", "theme.yaml")

	if err := Save("nord"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); !strings.HasPrefix(string(b), "preset: nord") {
		t.Errorf("new file: %q", b)
	}
	if err := os.WriteFile(path, []byte("preset: gruvbox\nblue:\n  dark: \"#123456\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save("dracula"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "preset: dracula") || strings.Contains(string(b), "gruvbox") || !strings.Contains(string(b), "#123456") {
		t.Errorf("preset replaced, overrides kept: %q", b)
	}
	if Save("nope") == nil {
		t.Error("unknown preset must be rejected")
	}
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func TestCycleAdvancesWrapsAndSaves(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	orig := Current()
	t.Cleanup(func() { _ = Apply(orig) })

	names := Presets()
	last := names[len(names)-1]
	if err := Apply(last); err != nil {
		t.Fatal(err)
	}
	next, err := Cycle()
	if err != nil || next != names[0] || Current() != names[0] {
		t.Fatalf("Cycle from the last preset must wrap to %q, got %q (%v)", names[0], next, err)
	}
	if b, _ := os.ReadFile(filepath.Join(home, ".config", "missionctl", "theme.yaml")); !strings.Contains(string(b), "preset: "+names[0]) {
		t.Errorf("Cycle must save the choice: %q", b)
	}
}

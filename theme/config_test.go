package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOverrides(t *testing.T) {
	origBlue, origAmber := Blue, Amber
	t.Cleanup(func() { Blue, Amber = origBlue, origAmber })

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".config", "missionctl"), 0755); err != nil {
		t.Fatal(err)
	}
	yamlContent := "blue:\n  dark: \"99\"\n"
	path := filepath.Join(dir, ".config", "missionctl", "theme.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	Blue, Amber = origBlue, origAmber // reset before exercising loadOverrides
	loadOverrides()

	if Blue.Dark != "99" {
		t.Errorf("Blue.Dark = %q, want %q (override should apply)", Blue.Dark, "99")
	}
	if Blue.Light != origBlue.Light {
		t.Errorf("Blue.Light = %q, want unchanged default %q (only dark was overridden)", Blue.Light, origBlue.Light)
	}
	if Amber != origAmber {
		t.Errorf("Amber changed to %+v, want unchanged default %+v (not mentioned in config)", Amber, origAmber)
	}
}

func TestLoadOverrides_NoConfigFile(t *testing.T) {
	origBlue := Blue
	t.Cleanup(func() { Blue = origBlue })

	t.Setenv("HOME", t.TempDir()) // no ~/.config/missionctl/theme.yaml here
	loadOverrides()

	if Blue != origBlue {
		t.Errorf("Blue = %+v, want unchanged default %+v when no config file exists", Blue, origBlue)
	}
}

func writeThemeConfig(t *testing.T, yamlContent string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".config", "missionctl"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".config", "missionctl", "theme.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadOverrides_Preset(t *testing.T) {
	origBlue := Blue
	t.Cleanup(func() { Blue = origBlue })

	writeThemeConfig(t, "preset: dracula\n")
	loadOverrides()

	want := presets["dracula"].Blue.Dark
	if Blue.Dark != want {
		t.Errorf("Blue.Dark = %q, want Dracula's %q", Blue.Dark, want)
	}
}

func TestLoadOverrides_PresetWithPerKeyOverride(t *testing.T) {
	origBlue, origGreen := Blue, Green
	t.Cleanup(func() { Blue, Green = origBlue, origGreen })

	writeThemeConfig(t, "preset: dracula\nblue:\n  dark: \"#123456\"\n")
	loadOverrides()

	if Blue.Dark != "#123456" {
		t.Errorf("Blue.Dark = %q, want per-key override %q to win over preset", Blue.Dark, "#123456")
	}
	wantGreen := presets["dracula"].Green.Dark
	if Green.Dark != wantGreen {
		t.Errorf("Green.Dark = %q, want Dracula's %q (untouched by per-key overrides)", Green.Dark, wantGreen)
	}
}

func TestLoadOverrides_UnknownPreset(t *testing.T) {
	origBlue := Blue
	t.Cleanup(func() { Blue = origBlue })

	writeThemeConfig(t, "preset: not-a-real-theme\n")
	loadOverrides()

	if Blue != origBlue {
		t.Errorf("Blue = %+v, want unchanged default %+v for an unknown preset name", Blue, origBlue)
	}
}

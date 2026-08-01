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

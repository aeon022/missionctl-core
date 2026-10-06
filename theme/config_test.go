package theme

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

func TestTerminalPresetUsesOnlyAnsiPaletteColors(t *testing.T) {
	p, ok := presets["terminal"]
	if !ok {
		t.Fatal("terminal preset missing")
	}
	for name, c := range map[string]*colorOverride{
		"blue": p.Blue, "green": p.Green, "red": p.Red, "amber": p.Amber, "muted": p.Muted, "subtle": p.Subtle,
		"selected_bg": p.SelectedBg, "selected_fg": p.SelectedFg, "hover_bg": p.HoverBg, "on_accent": p.OnAccent,
	} {
		if c == nil {
			t.Errorf("%s not set", name)
			continue
		}
		for _, v := range []string{c.Light, c.Dark} {
			if n, err := strconv.Atoi(v); err != nil || n < 0 || n > 15 {
				t.Errorf("%s uses %q — the terminal preset must stay inside ANSI 0-15 so it follows the terminal theme", name, v)
			}
		}
	}
}

func TestDefaultPresetIsTerminalClassicRestoresOldPalette(t *testing.T) {
	orig := Blue
	t.Cleanup(func() { Blue = orig })

	t.Setenv("HOME", t.TempDir()) // no theme.yaml at all
	Blue = lipglossAdaptive("1", "2")
	loadOverrides()
	if want := presets["terminal"].Blue; Blue.Dark != want.Dark || Blue.Light != want.Light {
		t.Errorf("no config: Blue = %+v, want the terminal preset %+v", Blue, want)
	}

	writeThemeConfig(t, "preset: classic\n")
	loadOverrides()
	if Blue.Light != "25" || Blue.Dark != "33" {
		t.Errorf("preset: classic must restore the original palette, got %+v", Blue)
	}

	writeThemeConfig(t, "preset: no-such-theme\n")
	Blue = lipglossAdaptive("1", "2")
	loadOverrides()
	if want := presets["terminal"].Blue; Blue.Dark != want.Dark {
		t.Errorf("unknown preset must fall back to the default, got %+v", Blue)
	}

	writeThemeConfig(t, "preset: classic\nblue:\n  dark: \"77\"\n")
	loadOverrides()
	if Blue.Dark != "77" || Blue.Light != "25" {
		t.Errorf("a per-key override wins over the preset: %+v", Blue)
	}
}

func lipglossAdaptive(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

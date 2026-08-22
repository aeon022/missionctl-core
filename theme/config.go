package theme

import (
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"go.yaml.in/yaml/v4"
)

// colorOverride mirrors lipgloss.AdaptiveColor for YAML — either field left
// blank keeps that mode's default.
type colorOverride struct {
	Light string `yaml:"light"`
	Dark  string `yaml:"dark"`
}

type themeConfig struct {
	Preset     string         `yaml:"preset"`
	Blue       *colorOverride `yaml:"blue"`
	Green      *colorOverride `yaml:"green"`
	Red        *colorOverride `yaml:"red"`
	Amber      *colorOverride `yaml:"amber"`
	Muted      *colorOverride `yaml:"muted"`
	Subtle     *colorOverride `yaml:"subtle"`
	SelectedBg *colorOverride `yaml:"selected_bg"`
	SelectedFg *colorOverride `yaml:"selected_fg"`
	HoverBg    *colorOverride `yaml:"hover_bg"`
	OnAccent   *colorOverride `yaml:"on_accent"`
}

//go:embed presets/*.yaml
var presetFS embed.FS

// presets holds the built-in named palettes, keyed by lowercase filename
// without extension (e.g. "dracula"). Parsed once at package init since the
// embedded files never change at runtime.
var presets = loadPresets()

func loadPresets() map[string]themeConfig {
	entries, err := presetFS.ReadDir("presets")
	if err != nil {
		return nil
	}
	out := make(map[string]themeConfig, len(entries))
	for _, e := range entries {
		data, err := presetFS.ReadFile("presets/" + e.Name())
		if err != nil {
			continue
		}
		var cfg themeConfig
		if yaml.Unmarshal(data, &cfg) != nil {
			continue
		}
		name := strings.ToLower(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
		out[name] = cfg
	}
	return out
}

// loadOverrides reads ~/.config/missionctl/theme.yaml, if present, and
// resolves the palette in two tiers on top of the package defaults:
//
//  1. `preset:` — one of the built-in named palettes (currently catppuccin,
//     dracula, gruvbox, nord, one-dark, solarized, tokyo-night), applied
//     wholesale first. An unknown/misspelled name is silently ignored, same
//     as a missing config file — not every user wants to customize.
//  2. Per-key overrides (`blue:`, `green:`, …) from the same file, applied
//     last so they win over anything the preset set — a user can pick a
//     preset and still tweak a single color.
//
// Silently does nothing when the file is missing or unreadable — same "not
// configured is not an error" convention diaryctl's notectl writeback uses.
func loadOverrides() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "missionctl", "theme.yaml"))
	if err != nil {
		return
	}
	var cfg themeConfig
	if yaml.Unmarshal(data, &cfg) != nil {
		return
	}
	if cfg.Preset != "" {
		if preset, ok := presets[strings.ToLower(cfg.Preset)]; ok {
			applyConfig(preset)
		}
	}
	applyConfig(cfg)
}

func applyConfig(cfg themeConfig) {
	apply(&Blue, cfg.Blue)
	apply(&Green, cfg.Green)
	apply(&Red, cfg.Red)
	apply(&Amber, cfg.Amber)
	apply(&Muted, cfg.Muted)
	apply(&Subtle, cfg.Subtle)
	apply(&SelectedBg, cfg.SelectedBg)
	apply(&SelectedFg, cfg.SelectedFg)
	apply(&HoverBg, cfg.HoverBg)
	apply(&OnAccent, cfg.OnAccent)
}

func apply(c *lipgloss.AdaptiveColor, o *colorOverride) {
	if o == nil {
		return
	}
	if o.Light != "" {
		c.Light = o.Light
	}
	if o.Dark != "" {
		c.Dark = o.Dark
	}
}

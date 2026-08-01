package theme

import (
	"os"
	"path/filepath"

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

// loadOverrides reads ~/.config/missionctl/theme.yaml, if present, and
// applies any colors it sets on top of the package defaults. Silently does
// nothing when the file is missing or unreadable — same "not configured is
// not an error" convention diaryctl's notectl writeback uses, since not
// every user wants to customize the palette.
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

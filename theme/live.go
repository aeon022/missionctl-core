package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Live theme switching: tools can offer a theme picker that takes effect immediately.
// Colors are plain package variables, so Apply rewrites them, rebuilds the derived styles and
// recomputes the v2 colors; packages (or tools) that cached styles built from these colors
// register a refresh function with OnChange.

type palette struct {
	Blue, Green, Red, Amber, Muted, Subtle, SelectedBg, SelectedFg, HoverBg, OnAccent lipgloss.AdaptiveColor
}

var (
	defaults  palette
	current   = DefaultPreset
	listeners []func()
)

func captureDefaults() {
	defaults = palette{Blue, Green, Red, Amber, Muted, Subtle, SelectedBg, SelectedFg, HoverBg, OnAccent}
}

func resetToDefaults() {
	Blue, Green, Red, Amber, Muted, Subtle = defaults.Blue, defaults.Green, defaults.Red, defaults.Amber, defaults.Muted, defaults.Subtle
	SelectedBg, SelectedFg, HoverBg, OnAccent = defaults.SelectedBg, defaults.SelectedFg, defaults.HoverBg, defaults.OnAccent
}

// Presets lists the built-in palette names, sorted.
func Presets() []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Current is the name of the active preset.
func Current() string { return current }

// OnChange registers fn to run after every Apply (use it to rebuild cached styles).
func OnChange(fn func()) { listeners = append(listeners, fn) }

// Apply switches to the named preset right now. Per-key overrides from theme.yaml are not
// layered on top (a switch means "show me this palette"); Save keeps the choice for next launch.
func Apply(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	preset, ok := presets[name]
	if !ok {
		return fmt.Errorf("unknown theme %q (known: %s)", name, strings.Join(Presets(), ", "))
	}
	resetToDefaults()
	applyConfig(preset)
	current = name
	buildStyles()
	refreshV2()
	for _, fn := range listeners {
		fn()
	}
	return nil
}

var presetLine = regexp.MustCompile(`(?m)^preset:.*$`)

// Save makes name the preset every tool of the suite starts with: it sets `preset:` in
// ~/.config/missionctl/theme.yaml, keeping any per-key overrides already in the file.
func Save(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if _, ok := presets[name]; !ok {
		return fmt.Errorf("unknown theme %q", name)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "missionctl", "theme.yaml")
	data, _ := os.ReadFile(path)
	line := "preset: " + name
	var out string
	if presetLine.Match(data) {
		out = presetLine.ReplaceAllString(string(data), line)
	} else {
		out = line + "\n" + string(data)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(out), 0o600)
}

// Package themepicker is the small "T" popup every suite TUI shares: a list of the built-in colour
// themes that applies each one live while you move (the screen behind the popup changes with it),
// saves the choice for the whole suite on enter and restores the previous theme on esc.
//
// A tool keeps one Picker in its model; while Open() is true it sends key presses to Key and draws
// View() over its screen with overlay.Center (not CenterDim — the live preview is the point).
package themepicker

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aeon022/missionctl-core/theme"
)

// Picker is the popup's state. The zero value is closed.
type Picker struct {
	open   bool
	names  []string
	cursor int
	orig   string
	err    error
}

// Open shows the picker with the active theme selected.
func (p *Picker) Open() {
	p.names = theme.Presets()
	p.orig = theme.Current()
	p.cursor, p.err = 0, nil
	for i, n := range p.names {
		if n == p.orig {
			p.cursor = i
		}
	}
	p.open = true
}

// IsOpen reports whether the popup is showing and should receive key presses.
func (p Picker) IsOpen() bool { return p.open }

// Key handles one key press (the tea.KeyPressMsg String()). It returns the status text to flash after the
// popup closed ("Theme: nord", or an error), "" while it stays open. Unknown keys are swallowed.
func (p *Picker) Key(key string) (done string) {
	if !p.open {
		return ""
	}
	move := func(d int) {
		p.cursor = (p.cursor + d + len(p.names)) % len(p.names)
		p.err = theme.Apply(p.names[p.cursor]) // live preview
	}
	switch key {
	case "j", "down", "tab", "T":
		move(1)
	case "k", "up", "shift+tab":
		move(-1)
	case "enter", "space", " ":
		p.open = false
		name := p.names[p.cursor]
		if err := theme.Apply(name); err != nil {
			return "Theme: " + err.Error()
		}
		_ = theme.Save(name)
		return "Theme: " + name
	case "esc", "q", "ctrl+c":
		p.open = false
		_ = theme.Apply(p.orig)
	}
	return ""
}

// View draws the popup (about 36 columns wide); place it with overlay.Center.
func (p Picker) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Foreground(theme.BlueV2).Bold(true).Render("Colour theme")
	b.WriteString(title + "\n\n")
	cur := lipgloss.NewStyle().Background(theme.SelectedBgV2).Foreground(theme.SelectedFgV2).Bold(true)
	for i, n := range p.names {
		row := "  " + n + strings.Repeat(" ", max(0, 16-len(n))) + " "
		sw := swatch()
		if i == p.cursor {
			b.WriteString(cur.Render("▌ "+n+strings.Repeat(" ", max(0, 16-len(n)))+" ") + sw + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(theme.MutedV2).Render(row) + sw + "\n")
		}
	}
	if p.err != nil {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.RedV2).Render(p.err.Error()) + "\n")
	}
	hint := lipgloss.NewStyle().Foreground(theme.SubtleV2).Render("↑↓ preview · enter keep · esc cancel")
	b.WriteString("\n" + hint)
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.BlueV2).Padding(1, 2).Render(b.String())
}

// swatch is a row of the active theme's accent colours. Only the selected row's colours are the
// live theme, so it is the same for every row: the point is the whole screen changing behind the popup.
func swatch() string {
	dot := func(c lipgloss.Style) string { return c.Render("●") }
	return dot(lipgloss.NewStyle().Foreground(theme.BlueV2)) + dot(lipgloss.NewStyle().Foreground(theme.GreenV2)) +
		dot(lipgloss.NewStyle().Foreground(theme.AmberV2)) + dot(lipgloss.NewStyle().Foreground(theme.RedV2))
}

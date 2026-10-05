// Package emptystate renders the "nothing here" and "loading…" screens the
// same way in every suite TUI: a small centered block, muted, with an
// optional hint on what to do next.
package emptystate

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aeon022/missionctl-core/theme"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(theme.MutedV2).Bold(true)
	hintStyle  = lipgloss.NewStyle().Foreground(theme.SubtleV2)
)

// Render centers icon, title and hint in a width×height area. Empty parts
// are skipped; with width/height 0 it returns just the block, uncentered.
func Render(width, height int, icon, title, hint string) string {
	var lines []string
	if icon != "" {
		lines = append(lines, icon)
	}
	if title != "" {
		lines = append(lines, titleStyle.Render(title))
	}
	if hint != "" {
		lines = append(lines, hintStyle.Render(hint))
	}
	block := strings.Join(lines, "\n")
	if width <= 0 || height <= 0 {
		return block
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Align(lipgloss.Center).Render(block))
}

// Loading is Render for the waiting state; spinner is the caller's current
// spinner frame (e.g. m.spinner.View()).
func Loading(width, height int, spinner, text string) string {
	if text == "" {
		text = "Loading…"
	}
	return Render(width, height, spinner, text, "")
}

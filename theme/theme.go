// Package theme is the shared Lipgloss color palette and base styles used
// across the missionctl suite's Bubble Tea TUIs (mailctl, calctl, taskctl,
// notectl, budgetctl, timectl, diaryctl). It exists because these seven
// tools had already converged on an identical palette by copy-paste; this
// package makes that convergence explicit instead of accidental.
//
// habctl intentionally keeps its own distinct palette (its own visual
// identity) and is not expected to adopt this package.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	Blue   = lipgloss.AdaptiveColor{Light: "25", Dark: "33"}
	Green  = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}
	Red    = lipgloss.AdaptiveColor{Light: "160", Dark: "203"}
	Amber  = lipgloss.AdaptiveColor{Light: "214", Dark: "220"}
	Muted  = lipgloss.AdaptiveColor{Light: "243", Dark: "246"}
	Subtle = lipgloss.AdaptiveColor{Light: "250", Dark: "239"}

	SelectedBg = lipgloss.AdaptiveColor{Light: "189", Dark: "17"}
	SelectedFg = lipgloss.AdaptiveColor{Light: "16", Dark: "255"}
)

var (
	Header = lipgloss.NewStyle().Bold(true).Foreground(Blue)
	Help   = lipgloss.NewStyle().Foreground(Muted)
	OK     = lipgloss.NewStyle().Foreground(Green)
	Err    = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Key    = lipgloss.NewStyle().Foreground(Amber)

	Selected = lipgloss.NewStyle().Background(SelectedBg).Foreground(SelectedFg).Padding(0, 1)
)

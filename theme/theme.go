// Package theme is the shared Lipgloss color palette and base styles used
// across the missionctl suite's Bubble Tea TUIs (mailctl, calctl, taskctl,
// notectl, budgetctl, timectl, diaryctl). It exists because these seven
// tools had already converged on an identical palette by copy-paste; this
// package makes that convergence explicit instead of accidental.
//
// The palette below are the defaults; ~/.config/missionctl/theme.yaml can
// override any of them (see loadOverrides) — consuming tools need no code
// changes to pick up a user's overrides, since it's applied here at import
// time before any tool's own styles are built from these colors.
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
	Subtle = lipgloss.AdaptiveColor{Light: "250", Dark: "244"}

	SelectedBg = lipgloss.AdaptiveColor{Light: "189", Dark: "17"}
	SelectedFg = lipgloss.AdaptiveColor{Light: "16", Dark: "255"}

	// HoverBg previews the row under the mouse cursor before a click
	// commits it as the selection — deliberately a plain neutral gray,
	// less saturated than SelectedBg, so hover reads as "not yet chosen."
	HoverBg = lipgloss.AdaptiveColor{Light: "252", Dark: "236"}

	// OnAccent is foreground text rendered on top of a saturated accent
	// background (e.g. an active tab) — plain white in both modes, since
	// it needs to read against Blue/Green/etc, not against the terminal's
	// own background. mailctl/notectl/budgetctl each had this hardcoded
	// as lipgloss.Color("15") independently before converging here.
	OnAccent = lipgloss.AdaptiveColor{Light: "15", Dark: "15"}
)

var (
	Header   lipgloss.Style
	Help     lipgloss.Style
	OK       lipgloss.Style
	Err      lipgloss.Style
	Key      lipgloss.Style
	Selected lipgloss.Style

	// Hover must be applied by wrapping an already-styled composite string
	// in one Render() call (like Selected/styleCursor do) — it needs a
	// Background set to do that safely. A Background-less style (e.g.
	// Underline-only) makes lipgloss fall back to rendering the string
	// rune-by-rune, which corrupts any ANSI escape codes already embedded
	// in the content (nested color styling gets printed as literal text).
	// Found via a forced-ANSI render while wiring up hover in taskctl.
	Hover lipgloss.Style
)

func init() {
	loadOverrides()
	buildStyles()
}

// buildStyles constructs the derived styles from the current color vars.
// Must run after loadOverrides — lipgloss.Style.Foreground/Background take
// the color by value, so a style built before an override is applied would
// keep the old color baked in.
func buildStyles() {
	Header = lipgloss.NewStyle().Bold(true).Foreground(Blue)
	Help = lipgloss.NewStyle().Foreground(Muted)
	OK = lipgloss.NewStyle().Foreground(Green)
	Err = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Key = lipgloss.NewStyle().Foreground(Amber)
	Selected = lipgloss.NewStyle().Background(SelectedBg).Foreground(SelectedFg).Padding(0, 1)
	Hover = lipgloss.NewStyle().Background(HoverBg)
}

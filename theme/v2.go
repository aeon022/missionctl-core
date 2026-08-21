package theme

import (
	"image/color"
	"os"

	spinnerv2 "charm.land/bubbles/v2/spinner"
	lipglossv2 "charm.land/lipgloss/v2"
	"github.com/charmbracelet/lipgloss"
)

// V2 mirrors of this package's colors/styles, for suite tools on Bubble
// Tea v2 (bubbletea/bubbles/lipgloss v2 — new charm.land import paths).
// v1's lipgloss.AdaptiveColor doesn't exist in v2 at all: it's replaced by
// a LightDarkFunc resolved against a live tea.BackgroundColorMsg, meant
// for styles rebuilt per-render. This package (like every consumer's own
// style vars) builds its styles once at init instead, so these resolve
// light-vs-dark once too — via the same one-shot terminal query v1's
// AdaptiveColor did internally — rather than threading a background-color
// message through every consumer's Update loop just to reach parity with
// what was implicit before. Additive only: nothing above this file
// changes, so mailctl/calctl/taskctl/budgetctl/timectl/diaryctl staying on
// v1 are unaffected.
var (
	BlueV2   color.Color
	GreenV2  color.Color
	RedV2    color.Color
	AmberV2  color.Color
	MutedV2  color.Color
	SubtleV2 color.Color

	SelectedBgV2 color.Color
	SelectedFgV2 color.Color
	HoverBgV2    color.Color
	OnAccentV2   color.Color
)

var (
	HeaderV2   lipglossv2.Style
	HelpV2     lipglossv2.Style
	OKV2       lipglossv2.Style
	ErrV2      lipglossv2.Style
	KeyV2      lipglossv2.Style
	SelectedV2 lipglossv2.Style
	HoverV2    lipglossv2.Style
)

// resolve picks c's light or dark value the same way lipgloss v1's
// AdaptiveColor did — once, at startup — rather than v2's default of
// resolving per-render against a live background-color message.
func resolve(c lipgloss.AdaptiveColor, dark bool) color.Color {
	if dark {
		return lipglossv2.Color(c.Dark)
	}
	return lipglossv2.Color(c.Light)
}

func init() {
	dark := lipglossv2.HasDarkBackground(os.Stdin, os.Stdout)

	BlueV2 = resolve(Blue, dark)
	GreenV2 = resolve(Green, dark)
	RedV2 = resolve(Red, dark)
	AmberV2 = resolve(Amber, dark)
	MutedV2 = resolve(Muted, dark)
	SubtleV2 = resolve(Subtle, dark)
	SelectedBgV2 = resolve(SelectedBg, dark)
	SelectedFgV2 = resolve(SelectedFg, dark)
	HoverBgV2 = resolve(HoverBg, dark)
	OnAccentV2 = resolve(OnAccent, dark)

	HeaderV2 = lipglossv2.NewStyle().Bold(true).Foreground(BlueV2)
	HelpV2 = lipglossv2.NewStyle().Foreground(MutedV2)
	OKV2 = lipglossv2.NewStyle().Foreground(GreenV2)
	ErrV2 = lipglossv2.NewStyle().Foreground(RedV2).Bold(true)
	KeyV2 = lipglossv2.NewStyle().Foreground(AmberV2)
	SelectedV2 = lipglossv2.NewStyle().Background(SelectedBgV2).Foreground(SelectedFgV2).Padding(0, 1)
	HoverV2 = lipglossv2.NewStyle().Background(HoverBgV2)
}

// NewSpinnerV2 is NewSpinner for a Bubble Tea v2 consumer.
func NewSpinnerV2(style lipglossv2.Style) spinnerv2.Model {
	sp := spinnerv2.New()
	sp.Spinner = spinnerv2.MiniDot
	sp.Style = style
	return sp
}

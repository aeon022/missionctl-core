package theme

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// NewSpinner builds the MiniDot spinner every suite TUI already uses for
// sync/AI-call feedback, styled with the given (usually muted) style.
func NewSpinner(style lipgloss.Style) spinner.Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = style
	return sp
}

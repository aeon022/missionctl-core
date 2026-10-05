// Package statusbar lays out the one-line footer every suite TUI has: key
// hints on the left, status (sync age, profile, errors) on the right, never
// wider than the terminal.
package statusbar

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aeon022/missionctl-core/theme"
	"github.com/charmbracelet/x/ansi"
)

var (
	keyStyle  = lipgloss.NewStyle().Foreground(theme.AmberV2).Bold(true)
	descStyle = lipgloss.NewStyle().Foreground(theme.SubtleV2)
)

// Hint renders one "key description" pair.
func Hint(key, desc string) string { return keyStyle.Render(key) + " " + descStyle.Render(desc) }

// Hints joins key/description pairs with two spaces. Pairs are in priority
// order: when width is too small the LAST ones are dropped, so the most
// important hints always survive a narrow terminal.
func Hints(width int, pairs ...[2]string) string {
	var parts []string
	used := 0
	for _, p := range pairs {
		h := Hint(p[0], p[1])
		w := lipgloss.Width(h)
		if len(parts) > 0 {
			w += 2
		}
		if width > 0 && used+w > width {
			break
		}
		parts = append(parts, h)
		used += w
	}
	return strings.Join(parts, "  ")
}

// Line puts left and right on one row of exactly width cells. If they don't
// both fit, right is kept whole and left is truncated with "…"; if right
// alone doesn't fit it is truncated too.
func Line(width int, left, right string) string {
	if width <= 0 {
		return left + "  " + right
	}
	rw := lipgloss.Width(right)
	if rw >= width {
		return ansi.Truncate(right, width, "…")
	}
	room := width - rw
	if right != "" {
		room-- // at least one space between
	}
	if lipgloss.Width(left) > room {
		left = ansi.Truncate(left, room, "…")
	}
	gap := width - lipgloss.Width(left) - rw
	return left + strings.Repeat(" ", max(gap, 0)) + right
}

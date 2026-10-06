package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/aeon022/missionctl-core/config"
	"github.com/aeon022/missionctl-core/theme"
	"github.com/charmbracelet/x/ansi"
)

// BorderStyle is the suite-wide frame style for panels.
type BorderStyle string

const (
	BorderRounded BorderStyle = "rounded"
	BorderSharp   BorderStyle = "sharp"
	BorderNone    BorderStyle = "none"
)

// Borders returns the configured panel border style: MISSIONCTL_BORDERS or
// `borders: rounded|sharp|none` in ~/.config/missionctl/ui.yaml (default
// rounded).
func Borders() BorderStyle {
	v := strings.ToLower(os.Getenv("MISSIONCTL_BORDERS"))
	if v == "" {
		if home, err := os.UserHomeDir(); err == nil {
			s := config.NewStore("ui")
			s.AddPath(home + "/.config/missionctl")
			if s.Read() == nil {
				v = strings.ToLower(s.GetString("borders"))
			}
		}
	}
	switch BorderStyle(v) {
	case BorderSharp, BorderNone:
		return BorderStyle(v)
	}
	return BorderRounded
}

var borderGlyphs = map[BorderStyle][6]string{ // TL TR BL BR horizontal vertical
	BorderRounded: {"╭", "╮", "╰", "╯", "─", "│"},
	BorderSharp:   {"┌", "┐", "└", "┘", "─", "│"},
}

// Panel draws content in a framed box exactly width×height cells, with the
// title set into the top border (╭─ Tasks ──────╮). A focused panel gets the
// accent color, others are dimmed. Content is truncated/padded to fit; with
// borders: none the title becomes a divider line and the content follows.
func Panel(width, height int, title, content string, focused bool) string {
	if width < 4 || height < 1 {
		return ""
	}
	bc := theme.SubtleV2
	if focused {
		bc = theme.BlueV2
	}
	edge := lipgloss.NewStyle().Foreground(bc)
	titleSt := lipgloss.NewStyle().Foreground(theme.MutedV2)
	if focused {
		titleSt = lipgloss.NewStyle().Foreground(theme.BlueV2).Bold(true)
	}

	bs := Borders()
	lines := strings.Split(content, "\n")
	if bs == BorderNone {
		out := []string{Divider(width, title)}
		for i := 0; i < height-1; i++ {
			out = append(out, fitLine("", lines, i, width))
		}
		return strings.Join(out[:min(len(out), height)], "\n")
	}

	g := borderGlyphs[bs]
	inner := width - 2
	top := edge.Render(g[0] + g[4])
	if title != "" {
		t := " " + ansi.Truncate(title, max(inner-4, 1), "…") + " "
		top += titleSt.Render(t)
		top += edge.Render(strings.Repeat(g[4], max(inner-1-lipgloss.Width(t), 0)))
	} else {
		top += edge.Render(strings.Repeat(g[4], max(inner-1, 0)))
	}
	top += edge.Render(g[1])

	out := []string{top}
	for i := 0; i < height-2; i++ {
		out = append(out, edge.Render(g[5])+fitLine(" ", lines, i, inner)+edge.Render(g[5]))
	}
	if height >= 2 {
		out = append(out, edge.Render(g[2]+strings.Repeat(g[4], inner)+g[3]))
	}
	return strings.Join(out, "\n")
}

// fitLine returns lines[i] (or "") with an optional left pad, truncated and
// padded to exactly w cells.
func fitLine(pad string, lines []string, i, w int) string {
	l := ""
	if i < len(lines) {
		l = lines[i]
	}
	l = pad + ansi.Truncate(l, max(w-lipgloss.Width(pad), 0), "…")
	return l + strings.Repeat(" ", max(w-lipgloss.Width(l), 0))
}

// Tabs renders a tab bar within width: the active tab as a filled pill, the
// others dimmed, each optionally followed by a count. When there are too many
// to fit, tabs far from the active one are replaced by "…".
func Tabs(width int, labels []string, active int, counts []int) string {
	render := func(i int) string {
		l := labels[i]
		if i < len(counts) && counts[i] > 0 {
			l += fmt.Sprintf(" %d", counts[i])
		}
		if i == active {
			return Pill(l, Info)
		}
		return lipgloss.NewStyle().Foreground(theme.MutedV2).Padding(0, 1).Render(l)
	}
	if len(labels) == 0 {
		return ""
	}
	lo, hi := 0, len(labels)-1
	build := func(lo, hi int) string {
		var parts []string
		if lo > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.SubtleV2).Render("…"))
		}
		for i := lo; i <= hi; i++ {
			parts = append(parts, render(i))
		}
		if hi < len(labels)-1 {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.SubtleV2).Render("…"))
		}
		return strings.Join(parts, " ")
	}
	out := build(lo, hi)
	for width > 0 && lipgloss.Width(out) > width && (lo < active || hi > active) {
		// drop the end farther from the active tab first
		if hi-active >= active-lo && hi > active {
			hi--
		} else if lo < active {
			lo++
		} else {
			hi--
		}
		out = build(lo, hi)
	}
	return out
}

// Duration formats a duration compactly: 45s, 12m, 2h, 1h 05m, 1d 3h.
func Duration(d time.Duration) string {
	d = d.Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		h, m := int(d.Hours()), int(d.Minutes())%60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	days, h := int(d.Hours())/24, int(d.Hours())%24
	if h == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, h)
}

// Frame stacks header, body and footer into exactly height lines: the body is
// padded with blank lines (or cut) to fill the space between them. Always
// returning a constant height avoids stale lines from a previous, taller
// frame in alt-screen mode. With height <= 0 the parts are just joined.
func Frame(height int, header, body, footer string) string {
	h := strings.Split(header, "\n")
	f := strings.Split(footer, "\n")
	if header == "" {
		h = nil
	}
	if footer == "" {
		f = nil
	}
	if height <= 0 {
		return strings.Join(append(append(h, body), f...), "\n")
	}
	room := max(height-len(h)-len(f), 0)
	b := strings.Split(body, "\n")
	if len(b) > room {
		b = b[:room]
	}
	for len(b) < room {
		b = append(b, "")
	}
	all := append(append(h, b...), f...)
	if len(all) > height {
		all = all[:height]
	}
	return strings.Join(all, "\n")
}

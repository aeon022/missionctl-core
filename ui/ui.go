// Package ui is the suite's visual vocabulary: the small, reusable pieces
// (pills, key caps, bars, sparklines, toasts, headers, selection rows …) every
// TUI draws with, so all tools look like one product. Everything returns
// plain styled strings built on missionctl-core/theme, so it follows the
// user's theme preset and light/dark mode.
package ui

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/aeon022/missionctl-core/config"
	"github.com/aeon022/missionctl-core/theme"
	"github.com/charmbracelet/x/ansi"
)

// Kind is the semantic color of a pill, toast, dot or bar.
type Kind int

const (
	Info Kind = iota
	OK
	Warn
	Err
	Muted
)

func kindColor(k Kind) color.Color {
	switch k {
	case OK:
		return theme.GreenV2
	case Warn:
		return theme.AmberV2
	case Err:
		return theme.RedV2
	case Muted:
		return theme.SubtleV2
	}
	return theme.BlueV2
}

// ── Pills, dots, key caps ─────────────────────────────────────────────────────

// Pill renders text as a filled badge: " overdue ".
func Pill(text string, k Kind) string {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.OnAccentV2).Background(kindColor(k)).Padding(0, 1).Render(text)
}

// Dot is a colored status dot (●).
func Dot(k Kind) string { return lipgloss.NewStyle().Foreground(kindColor(k)).Render("●") }

// KeyCap draws a key as a reversed cap: " enter ".
func KeyCap(key string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.OnAccentV2).Background(theme.SubtleV2).Padding(0, 1).Render(key)
}

// Hint is a key cap followed by a dimmed description.
func Hint(key, desc string) string {
	return KeyCap(key) + " " + lipgloss.NewStyle().Foreground(theme.MutedV2).Render(desc)
}

// ── Bars, sparklines, heat ────────────────────────────────────────────────────

var eighths = []rune("▏▎▍▌▋▊▉█")

// Bar draws a width-cell horizontal bar for ratio 0..1 with 1/8-cell
// resolution. When worseWhenFull is true the color escalates green → amber →
// red as ratio approaches and passes 1 (budgets); otherwise it stays blue
// (progress). Ratios above 1 draw a full bar.
func Bar(width int, ratio float64, worseWhenFull bool) string {
	if width < 1 {
		return ""
	}
	r := math.Max(0, math.Min(1, ratio))
	total := int(math.Round(r * float64(width) * 8))
	full, rem := total/8, total%8
	var fill strings.Builder
	fill.WriteString(strings.Repeat("█", full))
	used := full
	if rem > 0 && full < width {
		fill.WriteRune(eighths[rem-1])
		used++
	}
	k := Info
	if worseWhenFull {
		switch {
		case ratio >= 1:
			k = Err
		case ratio >= 0.8:
			k = Warn
		default:
			k = OK
		}
	}
	track := strings.Repeat("░", max(width-used, 0))
	return lipgloss.NewStyle().Foreground(kindColor(k)).Render(fill.String()) +
		lipgloss.NewStyle().Foreground(theme.SubtleV2).Render(track)
}

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

// Spark draws one block per value, scaled to the largest value. All-zero (or
// empty) input draws the lowest block so the width stays constant.
func Spark(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	hi := 0.0
	for _, v := range values {
		hi = math.Max(hi, v)
	}
	var b strings.Builder
	for _, v := range values {
		i := 0
		if hi > 0 && v > 0 {
			i = int(math.Round(v / hi * 7))
		}
		b.WriteRune(sparkRunes[min(max(i, 0), 7)])
	}
	return lipgloss.NewStyle().Foreground(theme.BlueV2).Render(b.String())
}

// Heat draws one heatmap cell for level 0..levels (0 = empty).
func Heat(level, levels int) string {
	if level <= 0 || levels <= 0 {
		return lipgloss.NewStyle().Foreground(theme.SubtleV2).Render("▪")
	}
	cells := []rune("▪◼■")
	idx := min(level*len(cells)/(levels+1), len(cells)-1)
	k := OK
	if frac := float64(level) / float64(levels); frac < 0.4 {
		k = Muted
	}
	return lipgloss.NewStyle().Foreground(kindColor(k)).Bold(level == levels).Render(string(cells[idx]))
}

// ── Toast, header, divider, selection row ────────────────────────────────────

var kindIcon = map[Kind]string{Info: "info", OK: "ok", Warn: "warn", Err: "err"}

// Toast renders a one-line message with an icon: "✓ Saved".
func Toast(k Kind, text string) string {
	icon := Icon(kindIcon[k])
	return lipgloss.NewStyle().Foreground(kindColor(k)).Bold(true).Render(icon) + " " + text
}

// Header lays out a full-width title bar: left, centered mid, right — the
// middle is dropped first, then the left is truncated; always exactly width
// cells.
func Header(width int, left, mid, right string) string {
	if width <= 0 {
		return left + "  " + mid + "  " + right
	}
	w := lipgloss.Width
	if w(left)+w(mid)+w(right)+4 > width {
		mid = ""
	}
	if w(left)+w(right)+2 > width {
		left = ansi.Truncate(left, max(width-w(right)-2, 0), "…")
	}
	gapTotal := max(width-w(left)-w(mid)-w(right), 0)
	l, r := gapTotal/2, gapTotal-gapTotal/2
	if mid == "" {
		l, r = gapTotal, 0
	}
	return left + strings.Repeat(" ", l) + mid + strings.Repeat(" ", r) + right
}

// Divider draws a rule of exactly width cells, with an optional label.
func Divider(width int, label string) string {
	rule := lipgloss.NewStyle().Foreground(theme.SubtleV2)
	if label == "" {
		return rule.Render(strings.Repeat("─", max(width, 0)))
	}
	l := " " + label + " "
	n := max(width-lipgloss.Width(l)-2, 0)
	return rule.Render("──") + lipgloss.NewStyle().Foreground(theme.MutedV2).Render(l) + rule.Render(strings.Repeat("─", n))
}

// Row pads text to width and, when selected, marks it with an accent bar and
// a full-width background so the selection reads at a glance.
func Row(width int, selected bool, text string) string {
	text = ansi.Truncate(text, max(width-2, 0), "…")
	pad := strings.Repeat(" ", max(width-2-lipgloss.Width(text), 0))
	if !selected {
		return "  " + text + pad
	}
	bar := lipgloss.NewStyle().Foreground(theme.BlueV2).Render("▌")
	// Dimmed (Subtle) text is often the same color as the selection background
	// (e.g. ANSI 8 in the terminal theme) and would vanish — lift it to Muted.
	text = swapForeground(text, theme.SubtleV2, theme.MutedV2)
	return bar + withBackground(" "+text+pad, theme.SelectedBgV2)
}

// fgParams returns the SGR parameters lipgloss emits for c as a foreground,
// e.g. ["90"] or ["38","5","244"].
func fgParams(c color.Color) []string {
	probe := lipgloss.NewStyle().Foreground(c).Render("\x00")
	i := strings.Index(probe, "\x00")
	if i < 0 || !strings.HasPrefix(probe, "\x1b[") {
		return nil
	}
	end := strings.Index(probe, "m")
	if end < 2 || end > i {
		return nil
	}
	return strings.Split(probe[2:end], ";")
}

var sgrRe = regexp.MustCompile("\x1b\\[([0-9;]*)m")

// swapForeground rewrites every SGR sequence in text that sets the foreground
// to from so that it sets to instead, wherever from appears among the
// sequence's parameters (it may be combined with bold etc.).
func swapForeground(text string, from, to color.Color) string {
	f, t := fgParams(from), fgParams(to)
	if len(f) == 0 || len(t) == 0 || strings.Join(f, ";") == strings.Join(t, ";") {
		return text
	}
	return sgrRe.ReplaceAllStringFunc(text, func(seq string) string {
		params := strings.Split(sgrRe.FindStringSubmatch(seq)[1], ";")
		for i := 0; i+len(f) <= len(params); i++ {
			if strings.Join(params[i:i+len(f)], ";") == strings.Join(f, ";") {
				np := append(append(append([]string{}, params[:i]...), t...), params[i+len(f):]...)
				return "\x1b[" + strings.Join(np, ";") + "m"
			}
		}
		return seq
	})
}

// HoverRow is Row for the row under the mouse: no accent bar, a neutral
// hover background, and — unlike painting a plain style over the text — the
// row's own colors (dates, pills, amounts) are kept.
func HoverRow(width int, text string) string {
	text = ansi.Truncate(text, max(width-2, 0), "…")
	pad := strings.Repeat(" ", max(width-2-lipgloss.Width(text), 0))
	text = swapForeground(text, theme.SubtleV2, theme.MutedV2)
	return "  " + withBackground(text+pad, theme.HoverBgV2)
}

// withBackground paints bg behind text that may already contain styled
// segments (pills, colored amounts): every SGR reset inside text is followed
// by the background again, so the selection bar stays unbroken while the
// segments keep their own colors.
func withBackground(text string, bg color.Color) string {
	probe := lipgloss.NewStyle().Background(bg).Render("\x00")
	i := strings.Index(probe, "\x00")
	if i < 0 {
		return text
	}
	pre, post := probe[:i], probe[i+1:]
	text = strings.ReplaceAll(text, "\x1b[0m", "\x1b[0m"+pre)
	text = strings.ReplaceAll(text, "\x1b[m", "\x1b[m"+pre)
	return pre + text + post
}

// ── Text helpers ─────────────────────────────────────────────────────────────

// RelTime says when t is relative to now in calendar days: today, tomorrow,
// yesterday, in 3d, 3d ago; further out a short date.
func RelTime(t, now time.Time) string {
	day := func(x time.Time) time.Time {
		return time.Date(x.Year(), x.Month(), x.Day(), 0, 0, 0, 0, now.Location())
	}
	d := int(math.Round(day(t.In(now.Location())).Sub(day(now)).Hours() / 24))
	switch {
	case d == 0:
		return "today"
	case d == 1:
		return "tomorrow"
	case d == -1:
		return "yesterday"
	case d > 1 && d <= 6:
		return fmt.Sprintf("in %dd", d)
	case d < -1 && d >= -6:
		return fmt.Sprintf("%dd ago", -d)
	case t.Year() == now.Year():
		return t.Format("Jan 2")
	}
	return t.Format("2006-01-02")
}

// MidEllipsis shortens s to width cells keeping the start and the end
// (“Rechnung…Q3.pdf”), which keeps file extensions and ids readable.
func MidEllipsis(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 1 {
		return strings.Repeat("…", max(width, 0))
	}
	keepTail := (width - 1) / 2
	keepHead := width - 1 - keepTail
	r := []rune(s)
	head := string(r[:min(keepHead, len(r))])
	for lipgloss.Width(head) > keepHead {
		head = string([]rune(head)[:len([]rune(head))-1])
	}
	tail := string(r[max(len(r)-keepTail, 0):])
	for lipgloss.Width(tail) > keepTail {
		tail = string([]rune(tail)[1:])
	}
	return head + "…" + tail
}

// Money formats amount right-aligned in width cells: green for income, red
// for spending, the cents dimmed, e.g. "+3,200.00".
func Money(amount float64, width int) string {
	sign := "+"
	k := OK
	switch {
	case amount < 0:
		sign, k = "-", Err
	case amount == 0:
		sign, k = " ", Muted
	}
	abs := math.Abs(amount)
	whole := int64(abs)
	cents := int64(math.Round((abs - float64(whole)) * 100))
	if cents == 100 {
		whole, cents = whole+1, 0
	}
	ws := groupThousands(whole)
	plain := fmt.Sprintf("%s%s.%02d", sign, ws, cents)
	pad := strings.Repeat(" ", max(width-len(plain), 0))
	return pad + lipgloss.NewStyle().Foreground(kindColor(k)).Render(sign+ws) +
		lipgloss.NewStyle().Foreground(theme.SubtleV2).Render(fmt.Sprintf(".%02d", cents))
}

func groupThousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	var out []string
	for len(s) > 3 {
		out = append([]string{s[len(s)-3:]}, out...)
		s = s[:len(s)-3]
	}
	return strings.Join(append([]string{s}, out...), ",")
}

// ── Icons ────────────────────────────────────────────────────────────────────

// iconSet maps a semantic name to its Unicode glyph and (optional) Nerd Font
// glyph. Unicode is the default and never needs a special font.
var iconSet = map[string][2]string{
	"ok":       {"✓", ""},
	"err":      {"✗", ""},
	"warn":     {"⚠", ""},
	"info":     {"ℹ", ""},
	"task":     {"✓", ""},
	"calendar": {"◷", ""},
	"timer":    {"⏱", ""},
	"diary":    {"✎", ""},
	"budget":   {"€", ""},
	"habit":    {"✦", ""},
	"note":     {"▤", ""},
	"mail":     {"✉", ""},
	"search":   {"⌕", ""},
	"bell":     {"◔", ""},
	"star":     {"★", ""},
	"attach":   {"⚲", ""},
}

// NerdIcons reports whether Nerd Font glyphs are enabled: MISSIONCTL_ICONS=nerd
// or `icons: nerd` in ~/.config/missionctl/ui.yaml. Default: off.
func NerdIcons() bool {
	if v := strings.ToLower(os.Getenv("MISSIONCTL_ICONS")); v != "" {
		return v == "nerd"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	s := config.NewStore("ui")
	s.AddPath(home + "/.config/missionctl")
	if s.Read() != nil {
		return false
	}
	return strings.EqualFold(s.GetString("icons"), "nerd")
}

// Icon returns the glyph for name ("" if unknown) in the configured set.
func Icon(name string) string {
	g, ok := iconSet[name]
	if !ok {
		return ""
	}
	if NerdIcons() {
		return g[1]
	}
	return g[0]
}

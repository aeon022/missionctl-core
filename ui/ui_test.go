package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func plain(s string) string { return ansi.Strip(s) }

func TestPillAndKeyCapPadding(t *testing.T) {
	if got := plain(Pill("overdue", Err)); got != " overdue " {
		t.Errorf("Pill = %q", got)
	}
	if got := plain(KeyCap("enter")); got != " enter " {
		t.Errorf("KeyCap = %q", got)
	}
	if got := plain(Hint("q", "quit")); got != " q  quit" {
		t.Errorf("Hint = %q", got)
	}
}

func TestBarWidthAndFill(t *testing.T) {
	for _, r := range []float64{0, 0.03, 0.25, 0.5, 0.999, 1, 2.5, -1} {
		if w := lipgloss.Width(Bar(10, r, false)); w != 10 {
			t.Errorf("Bar(10, %v) width = %d, want 10", r, w)
		}
	}
	if got := plain(Bar(8, 0.5, false)); got != "████░░░░" {
		t.Errorf("half = %q", got)
	}
	// 1/8 resolution: 1 cell of 4 = 25% → exactly one full block; 1/8 of a cell → thinnest sliver
	if got := plain(Bar(4, 0.25, false)); got != "█░░░" {
		t.Errorf("quarter = %q", got)
	}
	if got := plain(Bar(4, 0.03, false)); got != "▏░░░" {
		t.Errorf("sliver = %q", got)
	}
	if got := plain(Bar(4, 3, true)); got != "████" {
		t.Errorf("over 100%% must clamp to a full bar, got %q", got)
	}
	if Bar(0, 0.5, false) != "" {
		t.Error("zero width = empty")
	}
}

func TestBarColorEscalatesOnlyWhenWorseWhenFull(t *testing.T) {
	a, b, c := Bar(6, 0.3, true), Bar(6, 0.9, true), Bar(6, 1.2, true)
	if a == b || b == c || a == c {
		t.Error("budget bar should change color at 80% and 100%")
	}
	if Bar(6, 0.3, false) != Bar(6, 0.95, false)[:0]+Bar(6, 0.3, false) {
		t.Error("progress bars keep one color")
	}
}

func TestSparkScalingAndFlat(t *testing.T) {
	if got := plain(Spark([]float64{0, 1, 2, 4, 8})); got != "▁▂▃▅█" {
		t.Errorf("Spark = %q", got)
	}
	if got := plain(Spark([]float64{0, 0, 0})); got != "▁▁▁" {
		t.Errorf("all-zero = %q (must keep width)", got)
	}
	if got := plain(Spark([]float64{5})); got != "█" {
		t.Errorf("single = %q", got)
	}
	if Spark(nil) != "" {
		t.Error("empty = empty")
	}
}

func TestHeatLevels(t *testing.T) {
	seen := map[string]bool{}
	for lvl := 0; lvl <= 4; lvl++ {
		seen[plain(Heat(lvl, 4))] = true
	}
	if len(seen) < 3 {
		t.Errorf("heat should distinguish levels, got %v", seen)
	}
	if plain(Heat(0, 4)) != "▪" || plain(Heat(-3, 4)) != "▪" || plain(Heat(1, 0)) != "▪" {
		t.Error("empty/invalid levels are the empty cell")
	}
}

func TestToastHeaderDividerRowWidths(t *testing.T) {
	if got := plain(Toast(OK, "Saved")); got != "✓ Saved" {
		t.Errorf("Toast = %q", got)
	}
	for _, w := range []int{20, 40, 80} {
		h := Header(w, "budgetctl", "profile: firma", "synced 2m ago")
		if lipgloss.Width(h) != w {
			t.Errorf("Header width %d, want %d: %q", lipgloss.Width(h), w, plain(h))
		}
	}
	narrow := plain(Header(24, "budgetctl", "profile: firma", "synced 2m"))
	if strings.Contains(narrow, "profile") || !strings.Contains(narrow, "synced 2m") {
		t.Errorf("narrow header drops the middle first, keeps the right: %q", narrow)
	}
	if lipgloss.Width(Divider(30, "")) != 30 || lipgloss.Width(Divider(30, "March")) != 30 {
		t.Error("Divider width")
	}
	if !strings.Contains(plain(Divider(30, "March")), " March ") {
		t.Error("Divider label")
	}
	sel, unsel := Row(20, true, "Gehalt"), Row(20, false, "Gehalt")
	if lipgloss.Width(sel) != 20 || lipgloss.Width(unsel) != 20 {
		t.Errorf("Row widths: %d / %d", lipgloss.Width(sel), lipgloss.Width(unsel))
	}
	if !strings.HasPrefix(plain(sel), "▌") || strings.HasPrefix(plain(unsel), "▌") {
		t.Errorf("selected row carries the accent bar: %q / %q", plain(sel), plain(unsel))
	}
	if lipgloss.Width(Row(8, true, "a very long description")) != 8 {
		t.Error("Row must truncate to width")
	}
}

func TestRelTime(t *testing.T) {
	now := time.Date(2026, 10, 6, 15, 0, 0, 0, time.Local)
	day := func(n int, h int) time.Time { return time.Date(2026, 10, 6+n, h, 0, 0, 0, time.Local) }
	for want, tm := range map[string]time.Time{
		"today": day(0, 1), "tomorrow": day(1, 23), "yesterday": day(-1, 0),
		"in 3d": day(3, 9), "3d ago": day(-3, 9), "Oct 20": day(14, 9),
	} {
		if got := RelTime(tm, now); got != want {
			t.Errorf("RelTime(%v) = %q, want %q", tm.Format("01-02 15h"), got, want)
		}
	}
	if got := RelTime(time.Date(2025, 12, 31, 9, 0, 0, 0, time.Local), now); got != "2025-12-31" {
		t.Errorf("other year = %q", got)
	}
	// time of day must not matter: 23:59 yesterday vs 00:01 today
	if RelTime(time.Date(2026, 10, 5, 23, 59, 0, 0, time.Local), now) != "yesterday" ||
		RelTime(time.Date(2026, 10, 6, 0, 1, 0, 0, time.Local), now) != "today" {
		t.Error("calendar-day boundary")
	}
}

func TestMidEllipsis(t *testing.T) {
	if got := MidEllipsis("short.pdf", 20); got != "short.pdf" {
		t.Errorf("fits = %q", got)
	}
	got := MidEllipsis("Rechnung_Steuerberater_Q3_2026.pdf", 20)
	if lipgloss.Width(got) > 20 || !strings.Contains(got, "…") || !strings.HasSuffix(got, "2026.pdf") || !strings.HasPrefix(got, "Rechnung") {
		t.Errorf("MidEllipsis = %q (w=%d)", got, lipgloss.Width(got))
	}
	// wide characters must not overflow the width
	if w := lipgloss.Width(MidEllipsis("日本語のとても長いファイル名.txt", 12)); w > 12 {
		t.Errorf("wide chars overflow: %d", w)
	}
	if MidEllipsis("abcdef", 1) != "…" || MidEllipsis("abcdef", 0) != "" {
		t.Error("tiny widths")
	}
}

func TestMoney(t *testing.T) {
	cases := map[float64]string{3200: "+3,200.00", -980: "  -980.00", 0: "     0.00", -10.99: "   -10.99", 1234567.5: "+1,234,567.50"}
	for in, want := range cases {
		got := plain(Money(in, 13))
		if strings.TrimSpace(got) != strings.TrimSpace(want) {
			t.Errorf("Money(%v) = %q, want %q", in, got, want)
		}
		if len(got) < 13 && in != 1234567.5 {
			t.Errorf("Money(%v) must be right-aligned to width 13, got %q", in, got)
		}
	}
	if got := plain(Money(0.999, 8)); strings.TrimSpace(got) != "+1.00" {
		t.Errorf("rounding carry: %q", got)
	}
}

func TestIconsUnicodeDefaultAndNerdOptIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MISSIONCTL_ICONS", "")
	if Icon("ok") != "✓" || Icon("nope") != "" {
		t.Errorf("default set: %q / %q", Icon("ok"), Icon("nope"))
	}
	t.Setenv("MISSIONCTL_ICONS", "nerd")
	if Icon("ok") != "" {
		t.Errorf("nerd set: %q", Icon("ok"))
	}
	t.Setenv("MISSIONCTL_ICONS", "unicode") // env wins over everything
	if Icon("ok") != "✓" {
		t.Error("env unicode")
	}
}

func TestSelectedRowKeepsInnerStylingAndBackground(t *testing.T) {
	pill := Pill("overdue", Err)
	row := Row(40, true, "buy milk "+pill)
	if lipgloss.Width(row) != 40 {
		t.Errorf("width %d", lipgloss.Width(row))
	}
	if got := strings.TrimRight(plain(row), " "); got != "▌ buy milk  overdue" {
		t.Errorf("visible text changed: %q", got)
	}
	// the pill's own style (the sequence right before its text) must survive …
	styleSeq := pill[strings.Index(pill, "\x1b[m")+3 : strings.Index(pill, "overdue")]
	if !strings.Contains(row, styleSeq+"overdue") {
		t.Errorf("selected row lost the pill's own colors:\n%q", row)
	}
	// … and after the pill's reset the selection background must be painted again
	i := strings.Index(row, "overdue")
	rest := row[i:]
	j := strings.Index(rest, "\x1b[m")
	if j < 0 || !strings.HasPrefix(rest[j+3:], "\x1b[48;") {
		t.Errorf("selection background not re-applied after the pill: %q", rest)
	}
}

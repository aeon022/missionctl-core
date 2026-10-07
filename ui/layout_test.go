package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPanelExactSizeTitleInBorder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MISSIONCTL_BORDERS", "")
	for _, c := range []struct{ w, h int }{{30, 6}, {12, 3}, {80, 10}} {
		p := Panel(c.w, c.h, "Tasks", "one\ntwo", true)
		lines := strings.Split(p, "\n")
		if len(lines) != c.h {
			t.Errorf("%dx%d: %d lines", c.w, c.h, len(lines))
		}
		for i, l := range lines {
			if lipgloss.Width(l) != c.w {
				t.Errorf("%dx%d line %d width %d: %q", c.w, c.h, i, lipgloss.Width(l), ansi.Strip(l))
			}
		}
	}
	p := ansi.Strip(Panel(30, 6, "Tasks", "one\ntwo", false))
	lines := strings.Split(p, "\n")
	if !strings.HasPrefix(lines[0], "╭─ Tasks ─") || !strings.HasSuffix(lines[0], "╮") || !strings.HasPrefix(lines[5], "╰") {
		t.Errorf("frame:\n%s", p)
	}
	if !strings.HasPrefix(lines[1], "│ one") {
		t.Errorf("content row: %q", lines[1])
	}
}

func TestPanelTruncatesContentAndLongTitle(t *testing.T) {
	p := Panel(20, 4, "A very long panel title indeed", strings.Repeat("x", 100)+"\nl2\nl3\nl4\nl5", false)
	for _, l := range strings.Split(p, "\n") {
		if lipgloss.Width(l) != 20 {
			t.Errorf("overflow: %d %q", lipgloss.Width(l), ansi.Strip(l))
		}
	}
	if strings.Count(p, "\n")+1 != 4 || strings.Contains(ansi.Strip(p), "l4") {
		t.Errorf("extra content rows must be cut:\n%s", ansi.Strip(p))
	}
	if Panel(3, 5, "x", "y", false) != "" {
		t.Error("too narrow = empty")
	}
}

func TestPanelBorderStylesAndNone(t *testing.T) {
	t.Setenv("MISSIONCTL_BORDERS", "sharp")
	if got := ansi.Strip(Panel(14, 3, "T", "a", false)); !strings.HasPrefix(got, "┌─ T ") {
		t.Errorf("sharp: %q", got)
	}
	t.Setenv("MISSIONCTL_BORDERS", "none")
	got := ansi.Strip(Panel(14, 3, "T", "a", false))
	lines := strings.Split(got, "\n")
	if len(lines) != 3 || strings.ContainsAny(got, "╭│┌") || !strings.Contains(lines[0], " T ") || strings.TrimSpace(lines[1]) != "a" {
		t.Errorf("none = title divider + content, no frame:\n%s", got)
	}
	t.Setenv("MISSIONCTL_BORDERS", "bogus")
	if Borders() != BorderRounded {
		t.Error("unknown value falls back to rounded")
	}
}

func TestTabsActivePillCountsAndOverflow(t *testing.T) {
	labels := []string{"DASHBOARD", "POSTS", "QUEUE", "HISTORY", "STATS", "SETTINGS", "LOGS"}
	full := Tabs(0, labels, 1, []int{0, 12})
	text := ansi.Strip(full)
	if !strings.Contains(text, " POSTS 12 ") || !strings.Contains(text, "DASHBOARD") {
		t.Errorf("tabs = %q", text)
	}
	for _, w := range []int{80, 40, 24} {
		got := Tabs(w, labels, 3, nil)
		if lipgloss.Width(got) > w {
			t.Errorf("Tabs(%d) is %d wide: %q", w, lipgloss.Width(got), ansi.Strip(got))
		}
		if !strings.Contains(ansi.Strip(got), "HISTORY") {
			t.Errorf("the active tab must always stay visible at width %d: %q", w, ansi.Strip(got))
		}
	}
	if !strings.Contains(ansi.Strip(Tabs(24, labels, 3, nil)), "…") {
		t.Error("overflow shows an ellipsis")
	}
	if Tabs(10, nil, 0, nil) != "" {
		t.Error("no labels = empty")
	}
}

func TestDuration(t *testing.T) {
	for in, want := range map[time.Duration]string{
		0: "0s", 45 * time.Second: "45s", 12 * time.Minute: "12m", 2 * time.Hour: "2h",
		65 * time.Minute: "1h 05m", 3*time.Hour + 59*time.Minute: "3h 59m",
		27 * time.Hour: "1d 3h", 48 * time.Hour: "2d",
	} {
		if got := Duration(in); got != want {
			t.Errorf("Duration(%v) = %q, want %q", in, got, want)
		}
	}
	if got := Duration(59*time.Second + 600*time.Millisecond); got != "1m" {
		t.Errorf("rounds to the second first: %q", got)
	}
}

func TestFrameAlwaysExactHeight(t *testing.T) {
	for _, h := range []int{5, 10, 30} {
		for _, body := range []string{"a", "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no\np\nq\nr\ns\nt\nu\nv\nw\nx\ny\nz\n1\n2\n3\n4\n5\n6"} {
			got := Frame(h, "HEADER\n──────", body, "footer")
			if n := strings.Count(got, "\n") + 1; n != h {
				t.Errorf("Frame(%d) = %d lines", h, n)
			}
			if !strings.HasPrefix(got, "HEADER") || !strings.HasSuffix(got, "footer") {
				t.Errorf("header/footer must stay pinned:\n%s", got)
			}
		}
	}
	if got := Frame(0, "h", "b", "f"); got != "h\nb\nf" {
		t.Errorf("height 0 just joins: %q", got)
	}
	if got := Frame(4, "", "b", ""); got != "b\n\n\n" {
		t.Errorf("no header/footer: %q", got)
	}
}

func TestTabsLayoutSpansMatchTheDrawnLine(t *testing.T) {
	labels := []string{"All", "Baby", "Change-Management", "Linux", "Notes"}
	line, spans := TabsLayout(0, labels, 2, []int{90, 5, 23, 1, 30})
	plainLine := ansi.Strip(line)
	if len(spans) != len(labels) {
		t.Fatalf("spans = %d, want %d", len(spans), len(labels))
	}
	for _, sp := range spans {
		cell := []rune(plainLine)[sp.X0:sp.X1]
		want := " " + labels[sp.Index]
		if !strings.HasPrefix(string(cell), want) {
			t.Errorf("tab %d span [%d,%d) shows %q, want it to start with %q", sp.Index, sp.X0, sp.X1, string(cell), want)
		}
	}
	for i := 1; i < len(spans); i++ {
		if spans[i].X0 < spans[i-1].X1 {
			t.Errorf("spans overlap: %+v %+v", spans[i-1], spans[i])
		}
	}
	// windowed: spans only for visible tabs, offsets still exact
	narrow, nsp := TabsLayout(30, labels, 3, nil)
	if lipgloss.Width(narrow) > 30 || len(nsp) == 0 || len(nsp) >= len(labels) {
		t.Fatalf("windowed layout: w=%d spans=%d", lipgloss.Width(narrow), len(nsp))
	}
	for _, sp := range nsp {
		if got := string([]rune(ansi.Strip(narrow))[sp.X0:sp.X1]); !strings.Contains(got, labels[sp.Index]) {
			t.Errorf("windowed tab %d span shows %q", sp.Index, got)
		}
	}
	if Tabs(30, labels, 3, nil) != narrow {
		t.Error("Tabs must equal TabsLayout's line")
	}
}

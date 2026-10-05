package statusbar

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestHintsDropsLowestPriorityFirst(t *testing.T) {
	pairs := [][2]string{{"enter", "open"}, {"x", "action"}, {"?", "help"}, {"q", "quit"}}
	full := ansi.Strip(Hints(0, pairs...))
	if full != "enter open  x action  ? help  q quit" {
		t.Fatalf("unconstrained = %q", full)
	}
	narrow := ansi.Strip(Hints(22, pairs...))
	if narrow != "enter open  x action" {
		t.Errorf("width 22 = %q, want the first two hints only", narrow)
	}
	if w := lipgloss.Width(Hints(22, pairs...)); w > 22 {
		t.Errorf("width %d exceeds 22", w)
	}
	if got := Hints(3, pairs...); got != "" {
		t.Errorf("nothing fits → empty, got %q", ansi.Strip(got))
	}
}

func TestLineLayout(t *testing.T) {
	got := Line(30, "left", "right")
	if lipgloss.Width(got) != 30 || !strings.HasPrefix(got, "left") || !strings.HasSuffix(got, "right") {
		t.Errorf("Line = %q (w=%d)", got, lipgloss.Width(got))
	}
	tr := Line(12, "a very long left side", "synced 2m")
	if lipgloss.Width(tr) != 12 || !strings.HasSuffix(tr, "synced 2m") || !strings.Contains(tr, "…") {
		t.Errorf("truncated Line = %q", tr)
	}
	if tiny := Line(5, "x", "synced 2m ago"); lipgloss.Width(tiny) > 5 {
		t.Errorf("right alone too wide must truncate, got %q", tiny)
	}
}

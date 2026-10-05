package emptystate

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderCentersInArea(t *testing.T) {
	out := Render(40, 9, "✓", "No tasks", "press n to add one")
	if lipgloss.Width(out) != 40 || lipgloss.Height(out) != 9 {
		t.Errorf("size = %dx%d, want 40x9", lipgloss.Width(out), lipgloss.Height(out))
	}
	plain := ansi.Strip(out)
	for _, want := range []string{"✓", "No tasks", "press n to add one"} {
		if !strings.Contains(plain, want) {
			t.Errorf("missing %q in:\n%s", want, plain)
		}
	}
}

func TestRenderUncenteredAndSkipsEmptyParts(t *testing.T) {
	if got := ansi.Strip(Render(0, 0, "", "Empty", "")); got != "Empty" {
		t.Errorf("= %q", got)
	}
}

func TestLoadingDefaultText(t *testing.T) {
	if got := ansi.Strip(Loading(0, 0, "⠋", "")); got != "⠋\nLoading…" {
		t.Errorf("= %q", got)
	}
}

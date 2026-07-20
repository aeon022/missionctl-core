package keymap

import (
	"strings"
	"testing"
)

func TestHelpBuildsSectionsAndRows(t *testing.T) {
	out := New("taskctl", "tasks from the terminal").
		Section("Navigation").
		Row("j / k", "move down / up").
		Section("Other").
		Row("q", "quit").
		String()

	for _, want := range []string{"taskctl", "Navigation", "move down / up", "Other", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing %q:\n%s", want, out)
		}
	}
}

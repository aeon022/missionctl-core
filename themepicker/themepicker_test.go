package themepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aeon022/missionctl-core/theme"
)

func TestPickerPreviewsLiveSavesOnEnterAndRestoresOnEsc(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	orig := theme.Current()
	t.Cleanup(func() { _ = theme.Apply(orig) })

	var p Picker
	if p.IsOpen() {
		t.Fatal("the zero value is closed")
	}
	p.Open()
	if !p.IsOpen() || !strings.Contains(p.View(), orig) {
		t.Fatalf("open picker lists the themes incl. the active one:\n%s", p.View())
	}

	// esc restores the theme that was active when the popup opened
	p.Key("j")
	if theme.Current() == orig {
		t.Error("moving the cursor applies the theme live")
	}
	if msg := p.Key("esc"); msg != "" || p.IsOpen() || theme.Current() != orig {
		t.Errorf("esc must close and restore %q, got open=%v current=%q msg=%q", orig, p.IsOpen(), theme.Current(), msg)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "missionctl", "theme.yaml")); err == nil {
		t.Error("cancelling must not write theme.yaml")
	}

	// enter keeps and saves it
	p.Open()
	p.Key("down")
	chosen := theme.Current()
	if msg := p.Key("enter"); msg != "Theme: "+chosen || p.IsOpen() {
		t.Errorf("enter closes with a confirmation, got %q open=%v", msg, p.IsOpen())
	}
	b, err := os.ReadFile(filepath.Join(home, ".config", "missionctl", "theme.yaml"))
	if err != nil || !strings.Contains(string(b), "preset: "+chosen) {
		t.Errorf("the choice is saved suite-wide: %v %q", err, b)
	}

	// wraps around
	p.Open()
	for range theme.Presets() {
		p.Key("j")
	}
	if theme.Current() != chosen {
		t.Errorf("a full lap returns to the start, got %q want %q", theme.Current(), chosen)
	}
}

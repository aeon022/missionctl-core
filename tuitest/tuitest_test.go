package tuitest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

type probe struct{ log []string }

func (p probe) Init() tea.Cmd { return nil }
func (p probe) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		p.log = append(p.log, "key:"+m.String())
		if m.String() == "enter" {
			return p, tea.Quit
		}
	case tea.WindowSizeMsg:
		p.log = append(p.log, "size")
	case tea.MouseClickMsg:
		p.log = append(p.log, "click")
	}
	return p, nil
}
func (p probe) View() tea.View { return tea.NewView("log: " + strings.Join(p.log, ",")) }

func TestKeyStrings(t *testing.T) {
	for in, want := range map[string]string{
		"space": "space", "enter": "enter", "esc": "esc", "tab": "tab",
		"shift+tab": "shift+tab", "ctrl+c": "ctrl+c", "alt+x": "alt+x",
		"?": "?", "V": "V", "a": "a", "up": "up", "pgdown": "pgdown", "f1": "f1", "f12": "f12",
	} {
		if got := Key(in).String(); got != want {
			t.Errorf("Key(%q).String() = %q, want %q", in, got, want)
		}
	}
	if k := Key("space"); k.Text != " " {
		t.Errorf("space must carry Text \" \" like a real terminal, got %q", k.Text)
	}
}

func TestSendCollectsCmdsAndDrivesModel(t *testing.T) {
	m, cmds := Keys(probe{}, "j", "enter")
	if len(cmds) != 1 {
		t.Errorf("want the Quit cmd collected, got %d cmds", len(cmds))
	}
	if got := Text(m); got != "log: key:j,key:enter" {
		t.Errorf("Text = %q", got)
	}
	m, _ = Send(m, Click(1, 1), Resize(80, 24))
	if got := Text(m); !strings.HasSuffix(got, "click,size") {
		t.Errorf("mouse/resize not delivered: %q", got)
	}
}

func TestSmokePasses(t *testing.T) { Smoke(t, probe{}, "a", "esc") }

func TestKeyUnknownPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Key(\"nope\") should panic")
		}
	}()
	Key("nope")
}

func TestSmokeSizeDeliversSize(t *testing.T) {
	SmokeSize(t, probe{}, 60, 15, "a")
}

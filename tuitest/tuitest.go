// Package tuitest drives Bubble Tea v2 models in tests without a terminal:
// build the exact key/mouse/resize messages a terminal would deliver, feed
// them through Update, and look at the model or the rendered text. Commands
// returned by Update are collected but never executed, so a test can't touch
// real files, apps or the clipboard by accident.
package tuitest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

var named = map[string]rune{
	"enter": tea.KeyEnter, "esc": tea.KeyEscape, "tab": tea.KeyTab,
	"backspace": tea.KeyBackspace, "delete": tea.KeyDelete, "space": tea.KeySpace,
	"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
	"pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown, "home": tea.KeyHome, "end": tea.KeyEnd,
	"f1": tea.KeyF1, "f2": tea.KeyF2, "f3": tea.KeyF3, "f4": tea.KeyF4, "f5": tea.KeyF5, "f6": tea.KeyF6,
	"f7": tea.KeyF7, "f8": tea.KeyF8, "f9": tea.KeyF9, "f10": tea.KeyF10, "f11": tea.KeyF11, "f12": tea.KeyF12,
}

// Key builds the key press a terminal delivers for name: "enter", "esc",
// "space", "tab", "shift+tab", "up", "f1".."f12", "ctrl+c", "alt+x", or a single character
// such as "?" or "V". Note v2 reports the space bar as "space", not " ".
func Key(name string) tea.KeyPressMsg {
	var mod tea.KeyMod
	for {
		switch {
		case strings.HasPrefix(name, "ctrl+"):
			mod |= tea.ModCtrl
			name = strings.TrimPrefix(name, "ctrl+")
		case strings.HasPrefix(name, "alt+"):
			mod |= tea.ModAlt
			name = strings.TrimPrefix(name, "alt+")
		case strings.HasPrefix(name, "shift+"):
			mod |= tea.ModShift
			name = strings.TrimPrefix(name, "shift+")
		default:
			goto done
		}
	}
done:
	if code, ok := named[name]; ok {
		k := tea.KeyPressMsg{Code: code, Mod: mod}
		if code == tea.KeySpace && mod == 0 {
			k.Text = " "
		}
		return k
	}
	r := []rune(name)
	if len(r) != 1 {
		panic("tuitest.Key: unknown key " + name)
	}
	k := tea.KeyPressMsg{Code: r[0], Mod: mod}
	if mod&(tea.ModCtrl|tea.ModAlt) == 0 {
		k.Text = name
	}
	return k
}

func Click(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{Button: tea.MouseLeft, X: x, Y: y}
}

func Motion(x, y int) tea.MouseMotionMsg { return tea.MouseMotionMsg{X: x, Y: y} }

func Wheel(up bool, x, y int) tea.MouseWheelMsg {
	b := tea.MouseWheelDown
	if up {
		b = tea.MouseWheelUp
	}
	return tea.MouseWheelMsg{Button: b, X: x, Y: y}
}

func Resize(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

// Send feeds msgs through m.Update in order and returns the final model plus
// every non-nil command produced along the way (not executed).
func Send(m tea.Model, msgs ...tea.Msg) (tea.Model, []tea.Cmd) {
	var cmds []tea.Cmd
	for _, msg := range msgs {
		var c tea.Cmd
		m, c = m.Update(msg)
		if c != nil {
			cmds = append(cmds, c)
		}
	}
	return m, cmds
}

// Keys is Send with Key(name) for each name.
func Keys(m tea.Model, names ...string) (tea.Model, []tea.Cmd) {
	msgs := make([]tea.Msg, len(names))
	for i, n := range names {
		msgs[i] = Key(n)
	}
	return Send(m, msgs...)
}

// Text is the model's rendered view with all ANSI styling removed.
func Text(m tea.Model) string { return ansi.Strip(m.View().Content) }

// Smoke is SmokeSize at 100x30.
func Smoke(t *testing.T, m tea.Model, keys ...string) {
	t.Helper()
	SmokeSize(t, m, 100, 30, keys...)
}

// SmokeSize sizes the model to w×h, renders it, then presses each key and
// renders again. It fails the test on a panic or an empty frame — the cheap
// "does it still work at all" check every TUI should have. It returns
// nothing; use Send/Keys when the final model matters.
func SmokeSize(t *testing.T, m tea.Model, w, h int, keys ...string) {
	t.Helper()
	step := "initial resize"
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic after %s: %v", step, r)
		}
	}()
	m, _ = Send(m, Resize(w, h))
	if strings.TrimSpace(Text(m)) == "" {
		t.Fatal("empty view after resize")
	}
	for _, k := range keys {
		step = "key " + k
		m, _ = Keys(m, k)
		if strings.TrimSpace(Text(m)) == "" {
			t.Fatalf("empty view after key %q", k)
		}
	}
}

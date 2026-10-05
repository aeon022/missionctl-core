package applescript

import (
	"os/exec"
	"testing"
)

func TestEscape(t *testing.T) {
	if got := Escape(`a\b"c`); got != `a\\b\"c` {
		t.Errorf("Escape = %q", got)
	}
	if got := EscapeLine("a\r\nb\"c"); got != `a\nb\"c` {
		t.Errorf("EscapeLine = %q", got)
	}
}

func TestRun(t *testing.T) {
	if _, err := exec.LookPath("osascript"); err != nil {
		t.Skip("osascript not available")
	}
	if got, err := Run(`return "  hi  "`); err != nil || got != "hi" {
		t.Errorf("Run = %q, %v; want trimmed hi", got, err)
	}
	if got, _ := RunRaw(`return "  hi  "`); got != "  hi  " {
		t.Errorf("RunRaw = %q; want spaces kept", got)
	}
	if _, err := Run(`error "boom"`); err == nil {
		t.Error("want error from failing script")
	}
}

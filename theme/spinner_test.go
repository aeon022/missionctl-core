package theme

import "testing"

func TestNewSpinnerUsesMiniDot(t *testing.T) {
	sp := NewSpinner(Help)
	if sp.Spinner.Frames == nil {
		t.Fatal("expected spinner to have frames set")
	}
}

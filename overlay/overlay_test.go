package overlay

import (
	"strings"
	"testing"
)

func TestCenter_NeverTouchesBackgroundBorderRing(t *testing.T) {
	// Regression test: an early version of this ignored the background's
	// own border ring entirely, so a centered popup collided with a
	// bordered background panel and produced visibly doubled-up
	// "╭──╭──╮──╮" corners. inset must keep the popup strictly inside it.
	const bgW = 30
	contentRow := func(label string) string {
		return "│ " + label + strings.Repeat(" ", bgW-4-len(label)) + " │"
	}
	bg := strings.Join([]string{
		"╭" + strings.Repeat("─", bgW-2) + "╮",
		contentRow("content row 1"),
		contentRow("content row 2"),
		contentRow("content row 3"),
		contentRow("content row 4"),
		"╰" + strings.Repeat("─", bgW-2) + "╯",
	}, "\n")
	popup := strings.Join([]string{
		"╭──────╮",
		"│ help │",
		"╰──────╯",
	}, "\n")

	out := Center(bg, popup, bgW, 6, 1)
	lines := strings.Split(out, "\n")

	if lines[0] != "╭────────────────────────────╮" {
		t.Errorf("top border row must be untouched, got %q", lines[0])
	}
	if lines[len(lines)-1] != "╰────────────────────────────╯" {
		t.Errorf("bottom border row must be untouched, got %q", lines[len(lines)-1])
	}
	for i, l := range lines {
		if i == 0 || i == len(lines)-1 || len(l) == 0 {
			continue
		}
		runes := []rune(l)
		if runes[0] != '│' || runes[len(runes)-1] != '│' {
			t.Errorf("row %d: left/right border column must be untouched, got %q", i, l)
		}
	}
}

func TestCenter_PopupContentAppears(t *testing.T) {
	bg := strings.Join([]string{
		strings.Repeat(" ", 20),
		strings.Repeat(" ", 20),
		strings.Repeat(" ", 20),
		strings.Repeat(" ", 20),
	}, "\n")
	out := Center(bg, "XY", 20, 4, 0)
	if !strings.Contains(out, "XY") {
		t.Errorf("expected popup content \"XY\" to appear in the output: %q", out)
	}
}

func TestCenter_BackgroundShorterThanHeightIsPadded(t *testing.T) {
	// A 2-line background composited into a 10-row canvas must not panic
	// or drop the popup — the short background gets padded with blank
	// lines first.
	out := Center("a\nb", "X", 10, 10, 0)
	if len(strings.Split(out, "\n")) != 10 {
		t.Errorf("expected output padded to 10 rows, got %d", len(strings.Split(out, "\n")))
	}
	if !strings.Contains(out, "X") {
		t.Errorf("expected popup content to still appear: %q", out)
	}
}

func TestCenter_PopupWiderThanSafeAreaClampsInsteadOfPanicking(t *testing.T) {
	// max(inset, width-inset-popW) can go negative when popup is wider
	// than the background; must clamp, not panic or produce a negative
	// slice bound.
	out := Center("short", "this popup line is way wider than the background", 5, 1, 1)
	if out == "" {
		t.Error("expected non-empty output even when popup exceeds background width")
	}
}

func TestCenter_ShortBackgroundLineDoesNotShiftPopupColumn(t *testing.T) {
	// Regression test: found while rolling this out to calctl. ansi.Cut on
	// a background line SHORTER than the requested right bound just
	// returns whatever's there — it doesn't pad. Without padding every
	// line to width first, a short line (a blank separator, a brief status
	// message like "No events yet") produces a narrower left-hand slice
	// than the popup's other rows, so the popup's left border lands one or
	// more columns further left on exactly that row than on every other
	// row — a visible zigzag in what should be a straight vertical edge.
	bg := strings.Join([]string{
		strings.Repeat("=", 40), // full-width row
		"short",                 // much shorter than width — the trigger
		strings.Repeat("=", 40), // full-width row
	}, "\n")
	popup := strings.Join([]string{
		"╭────╮",
		"│ hi │",
		"╰────╯",
	}, "\n")

	out := Center(bg, popup, 40, 3, 0)
	lines := strings.Split(out, "\n")

	col := -1
	for i, l := range lines {
		idx := strings.IndexAny(l, "╭│╰")
		if idx < 0 {
			t.Fatalf("line %d: expected a popup border character, found none in %q", i, l)
		}
		if col == -1 {
			col = idx
			continue
		}
		if idx != col {
			t.Errorf("line %d: popup border at column %d, want %d (same as other rows) — short background line threw it out of alignment", i, idx, col)
		}
	}
}

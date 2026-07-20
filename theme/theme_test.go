package theme

import "testing"

func TestStylesRender(t *testing.T) {
	if Header.Render("x") == "" {
		t.Fatal("Header.Render produced empty output")
	}
	if Key.Render("k") == "" {
		t.Fatal("Key.Render produced empty output")
	}
}

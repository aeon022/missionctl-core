package uistate

import (
	"path/filepath"
	"testing"
)

type testState struct {
	Tab    int    `json:"tab"`
	Filter string `json:"filter"`
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	var loaded testState
	if ok := Load(path, &loaded); ok {
		t.Fatal("expected Load of a nonexistent file to report false")
	}

	want := testState{Tab: 2, Filter: "unread"}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}

	var got testState
	if ok := Load(path, &got); !ok {
		t.Fatal("expected Load to succeed after Save")
	}
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := Save(path, "not an object"); err != nil {
		t.Fatal(err)
	}

	var v testState
	if ok := Load(path, &v); ok {
		t.Fatal("expected Load of mismatched JSON shape to report false")
	}
}

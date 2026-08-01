// Package uistate persists small per-tool UI preferences (last active tab,
// last active filter) across restarts, so a tool doesn't reset to its
// defaults every launch. Deliberately generic — each tool defines its own
// small state struct and just calls Save/Load on it, JSON-encoded to one
// file. Not for anything that needs to survive data changes meaningfully
// (e.g. a raw list-row cursor index) — a stale index into a list that's
// since changed shape points at the wrong thing, which is worse than
// resetting to the top.
package uistate

import (
	"encoding/json"
	"os"
)

// Save JSON-encodes v to path, overwriting any existing file.
func Save(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads path and JSON-decodes it into v. Returns false (v left
// untouched) if the file is missing or unparseable — first launch, or a
// format change, both just mean "start from defaults."
func Load(path string, v any) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, v) == nil
}

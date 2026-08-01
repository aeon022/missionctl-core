// Package lastsync persists a single "last synced at" timestamp per tool in
// a small marker file — rewriting a whole YAML config on every sync (which
// may run frequently, e.g. via missionctl's dashboard sync-all) is more
// machinery and corruption risk than one timestamp needs, and DB row
// columns like updated_at aren't reliable here since local edits between
// syncs would also bump them, making "last synced" lie.
package lastsync

import (
	"os"
	"strings"
	"time"
)

// Save writes t to path as RFC3339.
func Save(path string, t time.Time) error {
	return os.WriteFile(path, []byte(t.Format(time.RFC3339)), 0644)
}

// Load reads the timestamp at path. ok is false if the file is missing or
// unparseable — never synced yet, from the caller's point of view.
func Load(path string) (t time.Time, ok bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, false
	}
	t, err = time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

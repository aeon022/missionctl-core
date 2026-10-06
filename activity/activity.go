// Package activity is the suite's shared activity log: every tool appends one
// line per thing the user did ("completed 'Steuer' in taskctl"), and
// `missionctl log` / diaryctl read it back. Titles only — never note bodies,
// mail text or amounts — and logging is strictly best-effort: a failure here
// must never break the action it describes.
package activity

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aeon022/missionctl-core/config"
)

// Event is one logged user action.
type Event struct {
	Time   time.Time `json:"ts"`
	Tool   string    `json:"tool"`
	Action string    `json:"action"` // added, completed, deleted, checked, started, stopped, wrote, sent, published, imported
	Title  string    `json:"title"`
}

const maxTitle = 120

// Dir is where the log lives: $MISSIONCTL_DATA_DIR, else ~/.local/share/missionctl.
func Dir() string {
	if d := os.Getenv("MISSIONCTL_DATA_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "missionctl")
}

func Path() string { return filepath.Join(Dir(), "activity.jsonl") }

// DiaryMode says what diaryctl does with the day's activity.
type DiaryMode string

const (
	DiaryAsk  DiaryMode = "ask"  // offer to add it (default)
	DiaryAuto DiaryMode = "auto" // add it automatically when the daily entry is generated
	DiaryOff  DiaryMode = "off"  // never
)

// Settings are read from ~/.config/missionctl/activity.yaml:
//
//	enabled: true      # false stops all logging
//	diary: ask         # ask | auto | off
type Settings struct {
	Enabled bool
	Diary   DiaryMode
}

func settingsFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "missionctl", "activity.yaml")
}

func newStore() *config.Store {
	s := config.NewStore("activity")
	s.AddPath(filepath.Dir(settingsFile()))
	_ = s.Read() // a missing file just means the defaults
	return s
}

// Load returns the effective settings. MISSIONCTL_ACTIVITY=off always wins.
func Load() Settings {
	st := Settings{Enabled: true, Diary: DiaryAsk}
	s := newStore()
	if v, ok := s.Get("enabled").(bool); ok {
		st.Enabled = v
	}
	switch m := DiaryMode(strings.ToLower(s.GetString("diary"))); m {
	case DiaryAsk, DiaryAuto, DiaryOff:
		st.Diary = m
	}
	if strings.EqualFold(os.Getenv("MISSIONCTL_ACTIVITY"), "off") {
		st.Enabled = false
	}
	return st
}

// SetDiaryMode persists the diary mode, keeping other settings.
func SetDiaryMode(m DiaryMode) error {
	s := newStore()
	s.Set("diary", string(m))
	if err := os.MkdirAll(filepath.Dir(settingsFile()), 0o755); err != nil {
		return err
	}
	return s.Write(settingsFile())
}

var mu sync.Mutex

// Log appends one event. Best-effort: errors are swallowed, and it does
// nothing when logging is disabled. Each event is a single O_APPEND write, so
// concurrent tools don't interleave partial lines.
func Log(tool, action, title string) {
	if !Load().Enabled {
		return
	}
	title = strings.Join(strings.Fields(title), " ")
	if utf8.RuneCountInString(title) > maxTitle {
		title = string([]rune(title)[:maxTitle-1]) + "…"
	}
	b, err := json.Marshal(Event{Time: time.Now(), Tool: tool, Action: action, Title: title})
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if os.MkdirAll(Dir(), 0o755) != nil {
		return
	}
	f, err := os.OpenFile(Path(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}

// Read returns the events with from <= Time < to, oldest first. Unparsable
// lines are skipped; a missing log is not an error.
func Read(from, to time.Time) ([]Event, error) {
	f, err := os.Open(Path())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Time.IsZero() {
			continue
		}
		if !e.Time.Before(from) && e.Time.Before(to) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out, sc.Err()
}

// Day returns the local-time bounds [start, end) of t's calendar day.
func Day(t time.Time) (time.Time, time.Time) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 0, 1)
}

const (
	blockStart = "<!-- activity:start -->"
	blockEnd   = "<!-- activity:end -->"
)

// MarkdownBlock renders events as a diary section, wrapped in markers so
// ReplaceBlock can refresh it later without touching the rest of the entry.
// It returns "" for no events.
func MarkdownBlock(events []Event) string {
	if len(events) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(blockStart + "\n## Activity\n\n")
	for _, e := range events {
		b.WriteString("- " + e.Time.Format("15:04") + " " + e.Tool + " · " + e.Action + " — " + e.Title + "\n")
	}
	b.WriteString(blockEnd)
	return b.String()
}

// HasBlock reports whether body already contains an activity block.
func HasBlock(body string) bool { return strings.Contains(body, blockStart) }

// ReplaceBlock puts block into body: replacing an existing marked block in
// place, else appending it after a blank line. An empty block leaves body alone.
func ReplaceBlock(body, block string) string {
	if block == "" {
		return body
	}
	if i := strings.Index(body, blockStart); i >= 0 {
		if j := strings.Index(body[i:], blockEnd); j >= 0 {
			return body[:i] + block + body[i+j+len(blockEnd):]
		}
	}
	return strings.TrimRight(body, "\n") + "\n\n" + block + "\n"
}

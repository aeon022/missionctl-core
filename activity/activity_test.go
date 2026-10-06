package activity

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func sandbox(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MISSIONCTL_DATA_DIR", t.TempDir())
	t.Setenv("MISSIONCTL_ACTIVITY", "")
}

func TestLogReadRoundTripAndDayWindow(t *testing.T) {
	sandbox(t)
	Log("taskctl", "completed", "  Steuer \n abgeben ")
	Log("habctl", "checked", "Sport")
	now := time.Now()
	from, to := Day(now)
	evs, err := Read(from, to)
	if err != nil || len(evs) != 2 {
		t.Fatalf("Read = %v, %v", evs, err)
	}
	if evs[0].Tool != "taskctl" || evs[0].Title != "Steuer abgeben" {
		t.Errorf("title not normalized: %+v", evs[0])
	}
	if y, _ := Read(from.AddDate(0, 0, -2), from.AddDate(0, 0, -1)); len(y) != 0 {
		t.Errorf("events leaked into another day: %v", y)
	}
}

func TestLogTruncatesAndNeverFails(t *testing.T) {
	sandbox(t)
	Log("notectl", "wrote", strings.Repeat("ä", 500))
	evs, _ := Read(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if len(evs) != 1 || len([]rune(evs[0].Title)) != maxTitle || !strings.HasSuffix(evs[0].Title, "…") {
		t.Fatalf("title = %q", evs[0].Title)
	}
	// unwritable data dir: must not panic or error out
	t.Setenv("MISSIONCTL_DATA_DIR", "/proc/definitely/not/writable")
	Log("x", "y", "z")
}

func TestDisabledByEnvAndByFile(t *testing.T) {
	sandbox(t)
	t.Setenv("MISSIONCTL_ACTIVITY", "off")
	Log("a", "b", "c")
	if _, err := os.Stat(Path()); err == nil {
		t.Error("env off must not write")
	}
	t.Setenv("MISSIONCTL_ACTIVITY", "")
	home, _ := os.UserHomeDir()
	_ = os.MkdirAll(home+"/.config/missionctl", 0o755)
	_ = os.WriteFile(home+"/.config/missionctl/activity.yaml", []byte("enabled: false\n"), 0o644)
	Log("a", "b", "c")
	if _, err := os.Stat(Path()); err == nil {
		t.Error("enabled: false must not write")
	}
	if Load().Enabled {
		t.Error("Load().Enabled")
	}
}

func TestDiaryModeDefaultsAndPersistence(t *testing.T) {
	sandbox(t)
	if got := Load().Diary; got != DiaryAsk {
		t.Errorf("default = %q, want ask", got)
	}
	if err := SetDiaryMode(DiaryAuto); err != nil {
		t.Fatal(err)
	}
	if got := Load().Diary; got != DiaryAuto {
		t.Errorf("after set = %q", got)
	}
	_ = os.WriteFile(settingsFile(), []byte("diary: bogus\n"), 0o644)
	if got := Load().Diary; got != DiaryAsk {
		t.Errorf("invalid value must fall back to ask, got %q", got)
	}
}

func TestConcurrentLogsStayWholeLines(t *testing.T) {
	sandbox(t)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); Log("t", "added", "item") }()
	}
	wg.Wait()
	evs, _ := Read(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if len(evs) != 50 {
		t.Errorf("got %d of 50 intact events", len(evs))
	}
}

func TestMarkdownBlockAndReplace(t *testing.T) {
	ts := time.Date(2026, 10, 6, 14, 5, 0, 0, time.Local)
	block := MarkdownBlock([]Event{{Time: ts, Tool: "taskctl", Action: "completed", Title: "Steuer"}})
	if !strings.Contains(block, "- 14:05 taskctl · completed — Steuer") || MarkdownBlock(nil) != "" {
		t.Fatalf("block: %q", block)
	}
	body := "# Heute\n\nText"
	once := ReplaceBlock(body, block)
	if !HasBlock(once) || !strings.HasPrefix(once, "# Heute\n\nText\n\n") {
		t.Errorf("append: %q", once)
	}
	block2 := MarkdownBlock([]Event{{Time: ts, Tool: "habctl", Action: "checked", Title: "Sport"}})
	twice := ReplaceBlock(once+"\nNachtrag", block2)
	if strings.Count(twice, "<!-- activity:start -->") != 1 || strings.Contains(twice, "Steuer") ||
		!strings.Contains(twice, "Sport") || !strings.HasSuffix(strings.TrimSpace(twice), "Nachtrag") {
		t.Errorf("replace must be in place and keep surrounding text:\n%s", twice)
	}
	if ReplaceBlock(body, "") != body {
		t.Error("empty block leaves the body alone")
	}
}

func TestSetEnabledPersistsAndKeepsDiaryMode(t *testing.T) {
	sandbox(t)
	if err := SetDiaryMode(DiaryAuto); err != nil {
		t.Fatal(err)
	}
	if err := SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if st := Load(); st.Enabled || st.Diary != DiaryAuto {
		t.Errorf("after SetEnabled(false): %+v", st)
	}
	if err := SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !Load().Enabled {
		t.Error("re-enable")
	}
}

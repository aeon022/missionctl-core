// Package doctor provides the shared building blocks for each suite tool's
// `<tool> doctor` healthcheck subcommand: does the config file parse, does
// the database open and have the expected table, is the macOS app it talks
// to (Calendar/Reminders/Notes/Mail) actually reachable. Path-drift bugs
// (a tool silently reading the wrong DB/config path) have been the most
// common real bug class found across this suite — doctor exists to catch
// that class of problem directly instead of via a confusing symptom
// somewhere else.
package doctor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/aeon022/missionctl-core/syncdir"
	_ "modernc.org/sqlite"
)

// Check is one pass/fail healthcheck result.
type Check struct {
	Label  string
	OK     bool
	Detail string // path checked, error message, or other context
}

// CheckFileExists reports whether path exists and is readable.
func CheckFileExists(label, path string) Check {
	info, err := os.Stat(path)
	if err != nil {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s: %v", path, err)}
	}
	if info.IsDir() {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s: is a directory, not a file", path)}
	}
	return Check{Label: label, OK: true, Detail: path}
}

// CheckSQLite opens dbPath and confirms it isn't corrupt and has the
// expected table by running a trivial `SELECT count(*)` against it.
func CheckSQLite(label, dbPath, table string) Check {
	if _, err := os.Stat(dbPath); err != nil {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s: %v", dbPath, err)}
	}
	db, err := sql.Open("sqlite", dbPath+"?_timeout=5000")
	if err != nil {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s: %v", dbPath, err)}
	}
	defer db.Close()

	var count int
	query := fmt.Sprintf("SELECT count(*) FROM %s", table)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s: table %q: %v", dbPath, table, err)}
	}
	return Check{Label: label, OK: true, Detail: fmt.Sprintf("%s (%d rows in %s)", dbPath, count, table)}
}

// CheckDataDir reports on a tool's data directory: whether it's running in
// local (private default) or shared (user-configured, possibly folder-
// synced) mode and the resolved path; a specific, actionable failure if
// dbPath is currently an undownloaded iCloud Drive placeholder rather than
// a generic "file not found"; and, if the lock is currently held by
// another process, a note naming that — surfaced as informational context
// rather than a failure, since "another instance has it open right now"
// isn't itself a problem, just something worth knowing about.
func CheckDataDir(label, dbPath string, shared bool) Check {
	if isPlaceholder, placeholderPath := syncdir.ICloudPlaceholder(dbPath); isPlaceholder {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf(
			"%s hasn't finished downloading from iCloud yet (found %s) — open Finder and download it, or disable \"Optimize Mac Storage\" for this folder",
			dbPath, placeholderPath)}
	}

	mode := "local"
	if shared {
		mode = "shared"
	}
	detail := fmt.Sprintf("%s (%s)", dbPath, mode)

	if lock, err := syncdir.Acquire(dbPath); err != nil {
		if errors.Is(err, syncdir.ErrLocked) {
			detail += " — currently open in another process"
		}
	} else {
		lock.Release()
	}

	return Check{Label: label, OK: true, Detail: detail}
}

// CheckAppleApp confirms a macOS app is actually installed, via AppleScript's
// `id of application`, with a short timeout so a doctor run never hangs on a
// non-responsive app. Deliberately not `tell application X to get name` —
// that returns success for literally any string, real app or not, since
// AppleScript hands back the name you asked for without resolving the
// bundle first. `id of application` forces bundle resolution and errors
// (-1728) when nothing matches, confirmed by direct testing.
func CheckAppleApp(label, appName string) Check {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-e", fmt.Sprintf(`id of application %q`, appName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Check{Label: label, OK: false, Detail: fmt.Sprintf("%s not found: %v (%s)", appName, err, string(out))}
	}
	return Check{Label: label, OK: true, Detail: appName + " installed"}
}

// PrintReport renders checks as a ✓/✗ list to stdout and reports whether
// every check passed.
func PrintReport(checks []Check) bool {
	allOK := true
	for _, c := range checks {
		mark := "✓"
		if !c.OK {
			mark = "✗"
			allOK = false
		}
		fmt.Printf("%s %-28s %s\n", mark, c.Label, c.Detail)
	}
	if allOK {
		fmt.Println("\nAll checks passed.")
	} else {
		fmt.Println("\nSome checks failed — see above.")
	}
	return allOK
}

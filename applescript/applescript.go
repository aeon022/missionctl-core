// Package applescript is the one place the suite shells out to osascript
// (Calendar, Reminders, Notes, Mail), so error handling and string escaping
// are identical in every tool.
package applescript

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes script and returns its stdout with surrounding whitespace
// trimmed. On failure the error carries osascript's own stderr.
func Run(script string) (string, error) {
	out, err := run(script)
	return strings.TrimSpace(out), err
}

// RunRaw is Run without trimming leading/trailing spaces — only trailing
// newlines are dropped. For scripts whose output is user text (a note body)
// where leading whitespace is data.
func RunRaw(script string) (string, error) {
	out, err := run(script)
	return strings.TrimRight(out, "\n"), err
}

func run(script string) (string, error) {
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("osascript: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("osascript: %w", err)
	}
	return string(out), nil
}

// Escape makes s safe inside an AppleScript "double-quoted" literal.
func Escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
}

// EscapeLine is Escape for text that may contain line breaks: "\n" becomes
// the AppleScript escape \n and "\r" is dropped, keeping the literal on one
// source line.
func EscapeLine(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", "").Replace(s)
}

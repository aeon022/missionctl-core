// Package palette implements the matching logic behind the ":" command
// palette prototype (first built in habctl): type a command's name instead
// of memorizing its single-key shortcut, get a live-filtered, ranked list
// back. This package only does matching — each tool keeps its own Command
// list (the names/descriptions/keys are tool-specific) and its own key
// dispatch (a chosen Command's Key is replayed through the tool's existing
// key handler, so behavior is guaranteed identical to typing the key
// directly).
package palette

import "strings"

// Command is one palette entry: Name is what the user types to match,
// Desc is shown next to it, Key is the existing single-key shortcut this
// command replays.
type Command struct {
	Name string
	Desc string
	Key  string
}

// Match returns cmds whose Name contains query (case-insensitive),
// name-prefix matches first — same ranking fzf/k9s-style pickers use
// elsewhere in this suite. An empty query returns every command, in the
// order given.
func Match(cmds []Command, query string) []Command {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return cmds
	}
	var prefix, contains []Command
	for _, c := range cmds {
		name := strings.ToLower(c.Name)
		switch {
		case strings.HasPrefix(name, query):
			prefix = append(prefix, c)
		case strings.Contains(name, query):
			contains = append(contains, c)
		}
	}
	return append(prefix, contains...)
}

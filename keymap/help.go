// Package keymap provides a small builder for the "?" help overlays used
// across the missionctl suite TUIs. Every tool had reimplemented the same
// key/row/section rendering helpers by hand (see taskctl.renderHelp,
// diaryctl.viewHelp, calctl.renderHelp, ...); this package gives them one
// shared implementation instead.
package keymap

import (
	"fmt"
	"strings"

	"github.com/aeon022/missionctl-core/theme"
)

// Help builds a "?" overlay body: a title line followed by sections of
// key/description rows, rendered with the shared theme styles.
type Help struct {
	toolName string
	tagline  string
	b        strings.Builder
}

// New starts a help overlay titled "<toolName> — <tagline>".
func New(toolName, tagline string) *Help {
	h := &Help{toolName: toolName, tagline: tagline}
	h.b.WriteString("\n  " + theme.Header.Render(toolName) + theme.Help.Render(" — "+tagline) + "\n")
	return h
}

// Section starts a new labeled group of key rows (e.g. "Navigation").
func (h *Help) Section(title string) *Help {
	h.b.WriteString("\n  " + theme.Header.Render(title) + "\n")
	return h
}

// Row adds one "key   description" line to the current section.
func (h *Help) Row(key, desc string) *Help {
	h.b.WriteString("  " + theme.Key.Render(fmt.Sprintf("%-9s", key)) + theme.Help.Render(desc) + "\n")
	return h
}

// String renders the accumulated help body.
func (h *Help) String() string {
	return h.b.String()
}

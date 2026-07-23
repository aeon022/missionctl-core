// Package overlay composites a popup on top of already-rendered background
// content — for transient help screens, confirmation dialogs, or anything
// else that should keep the surrounding view visible instead of replacing
// the whole screen with a full view-state switch.
package overlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Center composites popup on top of background, centered. inset keeps the
// popup's own border clear of the background's own border ring (row 0 /
// last content row, column 0 / last column) when the background is itself
// a fully bordered panel — pass 0 if the background has no enclosing
// border. Callers are expected to size popup so it fits within the
// inset-shrunk area; Center only clamps placement, it doesn't resize or
// scroll popup itself.
//
// Uses ansi.Cut (github.com/charmbracelet/x/ansi) to slice the background
// at exact visible-column boundaries rather than raw byte/rune indexing —
// background lines carry their own ANSI styling, and naive slicing could
// land mid-escape-sequence and corrupt it.
func Center(background, popup string, width, height, inset int) string {
	bgLines := strings.Split(background, "\n")
	actualH := len(bgLines)
	for len(bgLines) < height {
		bgLines = append(bgLines, strings.Repeat(" ", width))
	}
	// Pad every existing line out to width too — ansi.Cut on a line shorter
	// than the requested right bound just returns what's there, so a short
	// line (a blank separator, a short status message) would produce a
	// narrower left-hand slice than xOff and throw the popup's left edge
	// out of alignment on exactly that row.
	for i, l := range bgLines {
		if lw := lipgloss.Width(l); lw < width {
			bgLines[i] = l + strings.Repeat(" ", width-lw)
		}
	}

	popLines := strings.Split(popup, "\n")
	popW := 0
	for _, l := range popLines {
		if w := lipgloss.Width(l); w > popW {
			popW = w
		}
	}
	popH := len(popLines)

	minX, maxX := inset, max(inset, width-inset-popW)
	minY, maxY := inset, max(inset, actualH-inset-popH)
	xOff := clampInt((width-popW)/2, minX, maxX)
	yOff := clampInt((actualH-popH)/2, minY, maxY)

	for i, pl := range popLines {
		row := yOff + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bg := bgLines[row]
		left := ansi.Cut(bg, 0, xOff)
		right := ansi.Cut(bg, xOff+popW, width)
		padded := pl + strings.Repeat(" ", max(0, popW-lipgloss.Width(pl)))
		bgLines[row] = left + padded + right
	}
	return strings.Join(bgLines, "\n")
}

func clampInt(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

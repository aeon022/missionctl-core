// Package humanize provides small text-formatting helpers shared across the
// missionctl suite's TUIs — currently just relative-time formatting for
// "last synced" indicators.
package humanize

import (
	"fmt"
	"time"
)

// TimeAgo formats t relative to now as a short "Xh ago" style string. Used
// for "last synced" indicators in tool headers/footers.
func TimeAgo(t time.Time) string {
	return timeAgoSince(t, time.Now())
}

// timeAgoSince is TimeAgo with an injectable "now" for deterministic tests.
func timeAgoSince(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < 0:
		return "just now" // clock skew or a timestamp from the same instant
	case d < 10*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

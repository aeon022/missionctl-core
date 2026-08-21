// Package dateutil holds day/week boundary and date-argument parsing
// helpers shared across the missionctl suite, so tools agree on what
// "today" and "this week" mean instead of each reimplementing the same
// arithmetic. Consolidated from taskctl's internal/dateutil, habctl's
// models.ParseDateArg, and calctl's mcpserver-private duplicates of all
// four functions — all four independently converged on the same Monday-
// start-week, YYYY-MM-DD convention.
package dateutil

import (
	"fmt"
	"time"
)

// StartOfDay returns t truncated to 00:00:00 in t's location.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns t set to 23:59:59 in t's location.
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, t.Location())
}

// WeekRange returns the Monday 00:00:00–Sunday 23:59:59 bounds (local time)
// of the week containing now.
func WeekRange(now time.Time) (time.Time, time.Time) {
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	mon := now.AddDate(0, 0, -(wd - 1))
	mon = time.Date(mon.Year(), mon.Month(), mon.Day(), 0, 0, 0, 0, time.Local)
	sun := mon.AddDate(0, 0, 6)
	sun = time.Date(sun.Year(), sun.Month(), sun.Day(), 23, 59, 59, 0, time.Local)
	return mon, sun
}

// ParseDateArg parses a YYYY-MM-DD date string, defaulting to time.Now()
// when s is empty. Callers that must reject an empty argument (rather than
// defaulting it) should check s == "" themselves before calling this.
func ParseDateArg(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q — expected YYYY-MM-DD", s)
	}
	return t, nil
}

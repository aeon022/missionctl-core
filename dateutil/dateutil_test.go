package dateutil

import (
	"testing"
	"time"
)

func TestStartEndOfDay(t *testing.T) {
	mid := time.Date(2026, 8, 21, 14, 30, 0, 0, time.Local)
	if got := StartOfDay(mid); got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
		t.Errorf("StartOfDay(%v) = %v, want 00:00:00", mid, got)
	}
	if got := EndOfDay(mid); got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Errorf("EndOfDay(%v) = %v, want 23:59:59", mid, got)
	}
}

func TestWeekRange(t *testing.T) {
	// 2026-08-21 is a Friday.
	fri := time.Date(2026, 8, 21, 10, 0, 0, 0, time.Local)
	mon, sun := WeekRange(fri)
	if mon.Weekday() != time.Monday || mon.Day() != 17 {
		t.Errorf("WeekRange(%v) start = %v, want Monday 2026-08-17", fri, mon)
	}
	if sun.Weekday() != time.Sunday || sun.Day() != 23 || sun.Hour() != 23 {
		t.Errorf("WeekRange(%v) end = %v, want Sunday 2026-08-23 23:59:59", fri, sun)
	}
}

func TestParseDateArg(t *testing.T) {
	if got, err := ParseDateArg(""); err != nil || got.IsZero() {
		t.Errorf("ParseDateArg(\"\") = %v, %v; want time.Now(), nil", got, err)
	}
	got, err := ParseDateArg("2026-08-21")
	if err != nil {
		t.Fatalf("ParseDateArg(\"2026-08-21\") error: %v", err)
	}
	if got.Year() != 2026 || got.Month() != 8 || got.Day() != 21 {
		t.Errorf("ParseDateArg(\"2026-08-21\") = %v, want 2026-08-21", got)
	}
	if _, err := ParseDateArg("not-a-date"); err == nil {
		t.Error("ParseDateArg(\"not-a-date\") should have errored")
	}
}

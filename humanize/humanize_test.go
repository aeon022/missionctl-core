package humanize

import (
	"testing"
	"time"
)

func TestTimeAgoSince(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		delta time.Duration
		want  string
	}{
		{0, "just now"},
		{5 * time.Second, "just now"},
		{30 * time.Second, "30s ago"},
		{90 * time.Second, "1m ago"},
		{45 * time.Minute, "45m ago"},
		{2 * time.Hour, "2h ago"},
		{23 * time.Hour, "23h ago"},
		{25 * time.Hour, "1d ago"},
		{72 * time.Hour, "3d ago"},
	}
	for _, c := range cases {
		got := timeAgoSince(now.Add(-c.delta), now)
		if got != c.want {
			t.Errorf("timeAgoSince(-%v) = %q, want %q", c.delta, got, c.want)
		}
	}
}

func TestTimeAgo_FutureIsJustNow(t *testing.T) {
	if got := TimeAgo(time.Now().Add(time.Minute)); got != "just now" {
		t.Errorf("TimeAgo(future) = %q, want %q", got, "just now")
	}
}

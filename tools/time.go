package tools

import (
	"fmt"
	"time"
)

// Relative formats a timestamp as a human-readable relative string.
func TimeRelative(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	d := time.Since(t)
	if d < 0 {
		d = -d
	}

	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return timePluralize(seconds, "second") + " ago"
	case minutes < 60:
		return timePluralize(minutes, "minute") + " ago"
	case hours < 24:
		return timePluralize(hours, "hour") + " ago"
	case days < 14:
		return timePluralize(days, "day") + " ago"
	case weeks < 9:
		return timePluralize(weeks, "week") + " ago"
	case months < 12:
		return timePluralize(months, "month") + " ago"
	default:
		return timePluralize(years, "year") + " ago"
	}
}

func timePluralize(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

package wtree

import (
	"fmt"
	"time"
)

// relativeTime formats a timestamp as a human-readable relative string.
func relativeTime(t time.Time) string {
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
		return pluralize(seconds, "second") + " ago"
	case minutes < 60:
		return pluralize(minutes, "minute") + " ago"
	case hours < 24:
		return pluralize(hours, "hour") + " ago"
	case days < 14:
		return pluralize(days, "day") + " ago"
	case weeks < 9:
		return pluralize(weeks, "week") + " ago"
	case months < 12:
		return pluralize(months, "month") + " ago"
	default:
		return pluralize(years, "year") + " ago"
	}
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

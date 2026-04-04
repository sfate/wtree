package wtree

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		expected string
	}{
		{"30 seconds ago", 30 * time.Second, "30 seconds ago"},
		{"1 minute ago", 1 * time.Minute, "1 minute ago"},
		{"45 minutes ago", 45 * time.Minute, "45 minutes ago"},
		{"3 hours ago", 3 * time.Hour, "3 hours ago"},
		{"5 days ago", 5 * 24 * time.Hour, "5 days ago"},
		{"3 weeks ago", 3 * 7 * 24 * time.Hour, "3 weeks ago"},
		{"5 months ago", 5 * 30 * 24 * time.Hour, "5 months ago"},
		{"2 years ago", 2 * 365 * 24 * time.Hour, "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := time.Now().Add(-tt.offset)
			require.Equal(t, tt.expected, relativeTime(past))
		})
	}

	t.Run("zero time", func(t *testing.T) {
		require.Equal(t, "unknown", relativeTime(time.Time{}))
	})
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		n        int
		unit     string
		expected string
	}{
		{1, "second", "1 second"},
		{2, "second", "2 seconds"},
		{0, "minute", "0 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, pluralize(tt.n, tt.unit))
		})
	}
}

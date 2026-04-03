package version

import (
	"os"
	"strings"
)

const baseValue = "v0.1.0"

// Current returns the active version string.
func Current() string {
	data, err := os.ReadFile("VERSION")
	if err == nil {
		if v := strings.TrimSpace(string(data)); v != "" {
			return v
		}
	}

	return baseValue
}

package version

import (
	_ "embed"
	"strings"

	"golang.org/x/mod/semver"
)

// Value is overridden at build time via ldflags for release/install builds.
var (
	Value = ""
)

//go:embed VERSION
var data string

// Current returns the active version string.
func Current() string {
	if v := strings.TrimSpace(string(data)); v != "" {
		return normalize(v, true)
	}

	return "dev"
}

func normalize(v string, dev bool) string {
	if !semver.IsValid(v) {
		if dev && !strings.HasSuffix(v, "+dev") {
			return v + "+dev"
		}
		return v
	}

	if !dev {
		return semver.Canonical(v)
	}

	return baseVersion(v) + "+dev"
}

func baseVersion(v string) string {
	if prerelease := semver.Prerelease(v); prerelease != "" {
		v = strings.TrimSuffix(v, prerelease)
	}
	if idx := strings.Index(v, "+"); idx >= 0 {
		v = v[:idx]
	}
	return semver.Canonical(v)
}

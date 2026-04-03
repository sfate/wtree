package version

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"
)

//go:embed VERSION
var sourceVersion string
var sourceVersionPath string

type semverImpl struct {
	currentVersion string
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	sourceVersionPath = filepath.Join(filepath.Dir(filename), "VERSION")
}

func validate(v string) (string, error) {
	v = strings.TrimSpace(v)
	if !semver.IsValid(v) {
		return "", fmt.Errorf("invalid semver: %q", v)
	}

	return semver.Canonical(v), nil
}

func NewSemver() (*semverImpl, error) {
	v, err := validate(sourceVersion)
	if err != nil {
		return nil, err
	}

	return &semverImpl{
		currentVersion: v,
	}, nil
}

func (s *semverImpl) Current() string {
	return s.currentVersion
}

func (s *semverImpl) Bump(bump string) (string, error) {
	core := strings.TrimPrefix(s.currentVersion, "v")

	var major, minor, patch int
	if _, err := fmt.Sscanf(core, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return "", fmt.Errorf("parse semver %q: %w", s.currentVersion, err)
	}

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "", "patch":
		patch++
	default:
		return "", fmt.Errorf("unsupported bump type '%q'", bump)
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

func (s *semverImpl) Set(v string) error {
	ver, err := validate(v)
	if err != nil {
		return err
	}

	return os.WriteFile(sourceVersionPath, []byte(ver+"\n"), 0o644)
}

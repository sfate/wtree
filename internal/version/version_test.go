package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentUsesInjectedValue(t *testing.T) {
	orig := Value
	Value = "v9.9.9"
	t.Cleanup(func() { Value = orig })

	if got := Current(); got != "v9.9.9" {
		t.Fatalf("Current() = %q, want %q", got, "v9.9.9")
	}
}

func TestCurrentUsesVersionFileWithDevSuffix(t *testing.T) {
	orig := Value
	Value = ""
	t.Cleanup(func() { Value = orig })

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("v1.2.3\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if got := Current(); got != "v1.2.3+dev" {
		t.Fatalf("Current() = %q, want %q", got, "v1.2.3+dev")
	}
}

func TestCurrentCanonicalizesInjectedSemver(t *testing.T) {
	orig := Value
	Value = "v1.2.3+buildmeta"
	t.Cleanup(func() { Value = orig })

	if got := Current(); got != "v1.2.3" {
		t.Fatalf("Current() = %q, want %q", got, "v1.2.3")
	}
}

func TestCurrentReplacesPrereleaseFromVersionFileWithBuildDev(t *testing.T) {
	orig := Value
	Value = ""
	t.Cleanup(func() { Value = orig })

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("v1.2.3-rc.1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if got := Current(); got != "v1.2.3+dev" {
		t.Fatalf("Current() = %q, want %q", got, "v1.2.3+dev")
	}
}

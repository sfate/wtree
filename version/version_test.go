package version

import (
	"os"
	"testing"
)

func TestNewSemverCurrent(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")

	v, err := NewSemver()
	if err != nil {
		t.Fatalf("NewSemver() returned error: %v", err)
	}
	if got := v.Current(); got != "v1.2.3" {
		t.Fatalf("Current() = %q, want %q", got, "v1.2.3")
	}
}

func TestNewSemverRejectsInvalidVersion(t *testing.T) {
	withSourceVersion(t, "not-a-version\n")

	if _, err := NewSemver(); err == nil {
		t.Fatal("expected error for invalid VERSION")
	}
}

func TestSetWritesCanonicalVersion(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")
	withVersionFileBackup(t)

	v, err := NewSemver()
	if err != nil {
		t.Fatalf("NewSemver() returned error: %v", err)
	}

	if err := v.Set("v2.3.4+buildmeta"); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	data, err := os.ReadFile(sourceVersionPath)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if got := string(data); got != "v2.3.4\n" {
		t.Fatalf("VERSION file = %q, want %q", got, "v2.3.4\n")
	}
}

func TestBump(t *testing.T) {
	tests := []struct {
		bump string
		want string
	}{
		{bump: "patch", want: "v1.2.4"},
		{bump: "minor", want: "v1.3.0"},
		{bump: "major", want: "v2.0.0"},
	}

	for _, tt := range tests {
		withSourceVersion(t, "v1.2.3\n")

		v, err := NewSemver()
		if err != nil {
			t.Fatalf("NewSemver() returned error: %v", err)
		}

		got, err := v.Bump(tt.bump)
		if err != nil {
			t.Fatalf("Bump(%q) returned error: %v", tt.bump, err)
		}
		if got != tt.want {
			t.Fatalf("Bump(%q) = %q, want %q", tt.bump, got, tt.want)
		}
	}
}

func TestBumpRejectsUnsupportedType(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")

	v, err := NewSemver()
	if err != nil {
		t.Fatalf("NewSemver() returned error: %v", err)
	}

	if _, err := v.Bump("weird"); err == nil {
		t.Fatal("expected error for unsupported bump type")
	}
}

func withSourceVersion(t *testing.T, contents string) {
	t.Helper()
	orig := sourceVersion
	sourceVersion = contents
	t.Cleanup(func() {
		sourceVersion = orig
	})
}

func withVersionFileBackup(t *testing.T) {
	t.Helper()
	orig, err := os.ReadFile(sourceVersionPath)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.WriteFile(sourceVersionPath, orig, 0o644)
	})
}

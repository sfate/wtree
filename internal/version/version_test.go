package version

import (
	"os"
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

func TestSourceReadsCanonicalVersion(t *testing.T) {
	withVersionFile(t, "v1.2.3\n")

	got, err := Source()
	if err != nil {
		t.Fatalf("Source() returned error: %v", err)
	}
	if got != "v1.2.3" {
		t.Fatalf("Source() = %q, want %q", got, "v1.2.3")
	}
}

func TestSetSourceWritesCanonicalVersion(t *testing.T) {
	withVersionFile(t, "v1.2.3\n")

	if err := SetSource("v2.3.4+buildmeta"); err != nil {
		t.Fatalf("SetSource() returned error: %v", err)
	}

	data, err := os.ReadFile(filePath())
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if got := string(data); got != "v2.3.4\n" {
		t.Fatalf("VERSION file = %q, want %q", got, "v2.3.4\n")
	}
}

func TestCurrentUsesVersionFileWithBuildDev(t *testing.T) {
	orig := Value
	Value = ""
	t.Cleanup(func() { Value = orig })
	withVersionFile(t, "v1.2.3\n")

	if got := Current(); got != "v1.2.3+dev" {
		t.Fatalf("Current() = %q, want %q", got, "v1.2.3+dev")
	}
}

func TestCurrentReplacesPrereleaseFromVersionFileWithBuildDev(t *testing.T) {
	orig := Value
	Value = ""
	t.Cleanup(func() { Value = orig })
	withVersionFile(t, "v1.2.3-rc.1\n")

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
		got, err := Bump("v1.2.3", tt.bump)
		if err != nil {
			t.Fatalf("Bump(%q) returned error: %v", tt.bump, err)
		}
		if got != tt.want {
			t.Fatalf("Bump(%q) = %q, want %q", tt.bump, got, tt.want)
		}
	}
}

func withVersionFile(t *testing.T, contents string) {
	t.Helper()

	orig, err := os.ReadFile(filePath())
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.WriteFile(filePath(), orig, 0o644)
	})

	if err := os.WriteFile(filePath(), []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
}

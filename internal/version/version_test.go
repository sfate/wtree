package version

import "testing"

func TestCurrentUsesInjectedValue(t *testing.T) {
	orig := Value
	Value = "v9.9.9"
	t.Cleanup(func() { Value = orig })

	if got := Current(); got != "v9.9.9" {
		t.Fatalf("Current() = %q, want %q", got, "v9.9.9")
	}
}

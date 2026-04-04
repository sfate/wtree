package version

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSemverCurrent(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")

	v, err := NewSemver()
	require.NoError(t, err)
	require.Equal(t, "v1.2.3", v.Current())
}

func TestNewSemverRejectsInvalidVersion(t *testing.T) {
	withSourceVersion(t, "not-a-version\n")

	_, err := NewSemver()
	require.Error(t, err)
}

func TestSetWritesCanonicalVersion(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")
	withVersionFileBackup(t)

	v, err := NewSemver()
	require.NoError(t, err)

	require.NoError(t, v.Set("v2.3.4+buildmeta"))

	data, err := os.ReadFile(sourceVersionPath)
	require.NoError(t, err)
	require.Equal(t, "v2.3.4\n", string(data))
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
		require.NoError(t, err)

		got, err := v.Bump(tt.bump)
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}
}

func TestBumpRejectsUnsupportedType(t *testing.T) {
	withSourceVersion(t, "v1.2.3\n")

	v, err := NewSemver()
	require.NoError(t, err)

	_, err = v.Bump("weird")
	require.Error(t, err)
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
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.WriteFile(sourceVersionPath, orig, 0o644)
	})
}

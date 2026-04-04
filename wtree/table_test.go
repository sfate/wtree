package wtree

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintTableEmpty(t *testing.T) {
	output := captureStdout(func() {
		PrintTable(nil)
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	// Empty table: top separator, header row, middle separator, bottom separator = 4 lines
	require.Len(t, lines, 4)

	// First and last lines should be separator lines starting with +-
	require.True(t, strings.HasPrefix(lines[0], "+-"))
	require.True(t, strings.HasPrefix(lines[len(lines)-1], "+-"))

	// Second line should be the header
	require.Contains(t, lines[1], "Ref")
	require.Contains(t, lines[1], "Branch")
	require.Contains(t, lines[1], "Last Activity")
}

func TestPrintTableWithEntries(t *testing.T) {
	entries := []WorktreeEntry{
		{
			Ref:          "ABC-1234",
			Branch:       "ob-abc-1234",
			LastCommitAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Ref:          "ABC-5678",
			Branch:       "ob-abc-5678",
			LastCommitAt: time.Now().Add(-5 * 24 * time.Hour),
		},
	}

	output := captureStdout(func() {
		PrintTable(entries)
	})

	require.Contains(t, output, "ABC-1234")
	require.Contains(t, output, "ob-abc-1234")
	require.Contains(t, output, "ABC-5678")
	require.Contains(t, output, "ob-abc-5678")
}

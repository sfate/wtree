package wtree

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
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
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (3 separators + 1 header), got %d: %q", len(lines), output)
	}

	// First and last lines should be separator lines starting with +-
	if !strings.HasPrefix(lines[0], "+-") {
		t.Errorf("first line should be a separator, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "+-") {
		t.Errorf("last line should be a separator, got %q", lines[len(lines)-1])
	}

	// Second line should be the header
	if !strings.Contains(lines[1], "Ref") || !strings.Contains(lines[1], "Branch") || !strings.Contains(lines[1], "Last Activity") {
		t.Errorf("header line missing expected columns, got %q", lines[1])
	}
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

	if !strings.Contains(output, "ABC-1234") {
		t.Errorf("output should contain ref ABC-1234, got %q", output)
	}
	if !strings.Contains(output, "ob-abc-1234") {
		t.Errorf("output should contain branch ob-abc-1234, got %q", output)
	}
	if !strings.Contains(output, "ABC-5678") {
		t.Errorf("output should contain ref ABC-5678, got %q", output)
	}
	if !strings.Contains(output, "ob-abc-5678") {
		t.Errorf("output should contain branch ob-abc-5678, got %q", output)
	}
}

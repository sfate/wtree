package wtree

import (
	"strings"
	"testing"
	"time"

	"github.com/sfate/wtree/config"
	gitpkg "github.com/sfate/wtree/git"
)

func TestSortEntriesDesc(t *testing.T) {
	now := time.Now()
	entries := []WorktreeEntry{
		{Ref: "old", LastCommitAt: now.Add(-3 * 24 * time.Hour)},
		{Ref: "newest", LastCommitAt: now.Add(-1 * time.Hour)},
		{Ref: "middle", LastCommitAt: now.Add(-1 * 24 * time.Hour)},
	}

	sortEntriesDesc(entries)

	if entries[0].Ref != "newest" {
		t.Errorf("expected first entry to be 'newest', got %q", entries[0].Ref)
	}
	if entries[1].Ref != "middle" {
		t.Errorf("expected second entry to be 'middle', got %q", entries[1].Ref)
	}
	if entries[2].Ref != "old" {
		t.Errorf("expected third entry to be 'old', got %q", entries[2].Ref)
	}
}

func TestWorktreeEntryRelativeAge(t *testing.T) {
	e := WorktreeEntry{
		Ref:          "test",
		LastCommitAt: time.Now().Add(-2 * time.Hour),
	}
	got := e.RelativeAge()
	if got != "2 hours ago" {
		t.Errorf("RelativeAge() = %q, want %q", got, "2 hours ago")
	}
}

func TestDefaultProjectConfig(t *testing.T) {
	cfg := config.DefaultProjectConfig("myproject", "/code/myproject")

	if cfg.Name != "myproject" {
		t.Errorf("Name = %q, want %q", cfg.Name, "myproject")
	}
	if cfg.Path != "/code/myproject" {
		t.Errorf("Path = %q, want %q", cfg.Path, "/code/myproject")
	}
	if cfg.BaseDir != "/code/myproject/.wtree" {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, "/code/myproject/.wtree")
	}
	if cfg.TicketPrefix != "ABC-" {
		t.Errorf("TicketPrefix = %q, want %q", cfg.TicketPrefix, "ABC-")
	}
	if cfg.BranchPrefix != "ob-abc-" {
		t.Errorf("BranchPrefix = %q, want %q", cfg.BranchPrefix, "ob-abc-")
	}
}

func TestDeriveBranchErrorPath(t *testing.T) {
	projectDir, err := gitpkg.FindRoot()
	if err != nil {
		t.Skip("not in a git repository, skipping")
	}

	m := &Manager{
		cfg:        config.DefaultProjectConfig("repo", projectDir),
		projectDir: projectDir,
	}

	// Ref without TicketPrefix should return an error.
	_, err = m.deriveBranch("no-prefix-123")
	if err == nil {
		t.Error("expected error for ref without TicketPrefix, got nil")
	}
	if !strings.Contains(err.Error(), "[branch] is required") {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestDeriveBranchHappyPath(t *testing.T) {
	projectDir, err := gitpkg.FindRoot()
	if err != nil {
		t.Skip("not in a git repository, skipping")
	}

	m := &Manager{
		cfg:        config.DefaultProjectConfig("repo", projectDir),
		projectDir: projectDir,
	}

	// Use a ticket ref that is unlikely to match an existing branch.
	branch, err := m.deriveBranch("ABC-99999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "ob-abc-99999" {
		t.Errorf("deriveBranch(ABC-99999) = %q, want %q", branch, "ob-abc-99999")
	}
}

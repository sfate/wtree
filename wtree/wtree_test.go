package wtree

import (
	"testing"
	"time"

	"github.com/sfate/wtree/config"
	gitpkg "github.com/sfate/wtree/git"
	"github.com/stretchr/testify/require"
)

func TestSortEntriesDesc(t *testing.T) {
	now := time.Now()
	entries := []WorktreeEntry{
		{Ref: "old", LastCommitAt: now.Add(-3 * 24 * time.Hour)},
		{Ref: "newest", LastCommitAt: now.Add(-1 * time.Hour)},
		{Ref: "middle", LastCommitAt: now.Add(-1 * 24 * time.Hour)},
	}

	sortEntriesDesc(entries)

	require.Equal(t, "newest", entries[0].Ref)
	require.Equal(t, "middle", entries[1].Ref)
	require.Equal(t, "old", entries[2].Ref)
}

func TestWorktreeEntryRelativeAge(t *testing.T) {
	e := WorktreeEntry{
		Ref:          "test",
		LastCommitAt: time.Now().Add(-2 * time.Hour),
	}
	require.Equal(t, "2 hours ago", e.RelativeAge())
}

func TestDefaultProjectConfig(t *testing.T) {
	cfg := config.DefaultProjectConfig("myproject", "/code/myproject")

	require.Equal(t, "myproject", cfg.Name)
	require.Equal(t, "/code/myproject", cfg.Path)
	require.Equal(t, "/code/myproject/.wtree", cfg.BaseDir)
	require.Equal(t, "ABC-", cfg.TicketPrefix)
	require.Equal(t, "ob-", cfg.BranchPrefix)
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
	require.Error(t, err)
	require.Contains(t, err.Error(), "[branch] is required")
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
	require.NoError(t, err)
	require.Equal(t, "ob-abc-99999", branch)
}

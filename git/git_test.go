package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// makeTestRepo creates a temporary git repository with one initial commit on
// branch "main" and returns its root directory.
func makeTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
	} {
		_, err := Run(dir, args...)
		require.NoError(t, err)
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("hello"), 0o644))
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		_, err := Run(dir, args...)
		require.NoError(t, err)
	}
	return dir
}

// gitSetup runs a git command in dir and fails the test on error.
func gitSetup(t *testing.T, dir string, args ...string) {
	t.Helper()
	_, err := Run(dir, args...)
	require.NoError(t, err)
}

func TestRunSuccess(t *testing.T) {
	dir := makeTestRepo(t)
	out, err := Run(dir, "rev-parse", "--is-inside-work-tree")
	require.NoError(t, err)
	require.Equal(t, "true", out)
}

func TestRunError(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := Run(dir, "not-a-real-command-xyz")
	require.Error(t, err)
}

func TestCurrentBranch(t *testing.T) {
	dir := makeTestRepo(t)
	branch, err := NewClient().CurrentBranch(dir)
	require.NoError(t, err)
	require.Equal(t, "main", branch)
}

func TestLastCommitTime(t *testing.T) {
	dir := makeTestRepo(t)
	before := time.Now().Add(-5 * time.Second)

	ts, err := NewClient().LastCommitTime(dir, "main")
	require.NoError(t, err)
	require.False(t, ts.Before(before))
	require.False(t, ts.After(time.Now().Add(5*time.Second)))
}

func TestLastCommitTimeInvalidBranch(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := NewClient().LastCommitTime(dir, "no-such-branch-xyz")
	require.Error(t, err)
}

func TestFindBranchFound(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "feature-abc")

	got, err := NewClient().FindBranch(dir, "feature-abc")
	require.NoError(t, err)
	require.Equal(t, "feature-abc", got)
}

func TestFindBranchNotFound(t *testing.T) {
	dir := makeTestRepo(t)
	got, err := NewClient().FindBranch(dir, "no-such-branch-xyz")
	require.NoError(t, err)
	require.Equal(t, "", got)
}

func TestBranchExistsTrue(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "exists-branch")

	exists, err := NewClient().BranchExists(dir, "exists-branch")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestBranchExistsFalse(t *testing.T) {
	dir := makeTestRepo(t)
	exists, err := NewClient().BranchExists(dir, "ghost-branch-xyz")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestBranchExistsExactMatchOnly(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "feature-abc")
	gitSetup(t, dir, "branch", "feature-abc-old")

	client := NewClient()
	exists, err := client.BranchExists(dir, "feature-ab")
	require.NoError(t, err)
	require.False(t, exists)

	found, err := client.FindBranch(dir, "feature-abc")
	require.NoError(t, err)
	require.Equal(t, "feature-abc", found)
}

func TestRemoteBranchExistsTrue(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "update-ref", "refs/remotes/origin/feature/abc", "main")

	exists, err := NewClient().RemoteBranchExists(dir, "origin", "feature/abc")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestRemoteBranchExistsExactMatchOnly(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "update-ref", "refs/remotes/origin/feature-abc-old", "main")

	client := NewClient()
	exists, err := client.RemoteBranchExists(dir, "origin", "feature-abc")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestEnsureBranchCreateNew(t *testing.T) {
	dir := makeTestRepo(t)
	client := NewClient()
	require.NoError(t, client.EnsureBranch(dir, "new-branch", "main"))
	exists, err := client.BranchExists(dir, "new-branch")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestEnsureBranchAlreadyExists(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "already-exists")

	require.NoError(t, NewClient().EnsureBranch(dir, "already-exists", "main"))
}

func TestWorktreeRegisteredFalse(t *testing.T) {
	dir := makeTestRepo(t)
	registered, err := NewClient().WorktreeRegistered(dir, "/some/nonexistent/path")
	require.NoError(t, err)
	require.False(t, registered)
}

func TestWorktreeRegisteredDoesNotUseSubstringMatch(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "wt-branch")

	base := t.TempDir()
	registeredDir := filepath.Join(base, "feature")
	similarDir := filepath.Join(base, "feature-copy")

	client := NewClient()
	require.NoError(t, client.WorktreeAdd(dir, registeredDir, "wt-branch"))
	t.Cleanup(func() {
		_ = client.WorktreeRemove(dir, registeredDir)
	})

	registered, err := client.WorktreeRegistered(dir, similarDir)
	require.NoError(t, err)
	require.False(t, registered)
}

func TestWorktreeAddAndRemove(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "wt-branch")

	wtDir := filepath.Join(t.TempDir(), "my-worktree")
	client := NewClient()

	require.NoError(t, client.WorktreeAdd(dir, wtDir, "wt-branch"))

	registered, err := client.WorktreeRegistered(dir, wtDir)
	require.NoError(t, err)
	require.True(t, registered)

	require.NoError(t, client.WorktreeRemove(dir, wtDir))

	registered, err = client.WorktreeRegistered(dir, wtDir)
	require.NoError(t, err)
	require.False(t, registered)
}

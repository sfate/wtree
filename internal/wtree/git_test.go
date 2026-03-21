package wtree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		if _, err := gitRun(dir, args...); err != nil {
			t.Fatalf("repo setup: git %s: %v", strings.Join(args, " "), err)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		if _, err := gitRun(dir, args...); err != nil {
			t.Fatalf("repo setup: git %s: %v", strings.Join(args, " "), err)
		}
	}
	return dir
}

// gitSetup runs a git command in dir and fails the test on error.
func gitSetup(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := gitRun(dir, args...); err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
}

func TestGitRun_Success(t *testing.T) {
	dir := makeTestRepo(t)
	out, err := gitRun(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("got %q, want %q", out, "true")
	}
}

func TestGitRun_Error(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := gitRun(dir, "not-a-real-command-xyz")
	if err == nil {
		t.Fatal("expected error for invalid git subcommand, got nil")
	}
}

func TestGitCurrentBranch(t *testing.T) {
	dir := makeTestRepo(t)
	branch, err := gitCurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("gitCurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestGitLastCommitTime(t *testing.T) {
	dir := makeTestRepo(t)
	before := time.Now().Add(-5 * time.Second)

	ts, err := gitLastCommitTime(dir, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Before(before) {
		t.Errorf("commit time %v predates lower bound %v", ts, before)
	}
	if ts.After(time.Now().Add(5 * time.Second)) {
		t.Errorf("commit time %v is unexpectedly in the future", ts)
	}
}

func TestGitLastCommitTime_InvalidBranch(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := gitLastCommitTime(dir, "no-such-branch-xyz")
	if err == nil {
		t.Fatal("expected error for non-existent branch, got nil")
	}
}

func TestGitFindBranch_Found(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "feature-abc")

	got, err := gitFindBranch(dir, "feature-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feature-abc" {
		t.Errorf("gitFindBranch() = %q, want %q", got, "feature-abc")
	}
}

func TestGitFindBranch_NotFound(t *testing.T) {
	dir := makeTestRepo(t)
	got, err := gitFindBranch(dir, "no-such-branch-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("gitFindBranch() = %q, want empty string", got)
	}
}

func TestGitBranchExists_True(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "exists-branch")

	if !gitBranchExists(dir, "exists-branch") {
		t.Error("gitBranchExists() = false, want true")
	}
}

func TestGitBranchExists_False(t *testing.T) {
	dir := makeTestRepo(t)
	if gitBranchExists(dir, "ghost-branch-xyz") {
		t.Error("gitBranchExists() = true, want false")
	}
}

func TestGitEnsureBranch_CreateNew(t *testing.T) {
	dir := makeTestRepo(t)
	if err := gitEnsureBranch(dir, "new-branch", "main"); err != nil {
		t.Fatalf("gitEnsureBranch() error: %v", err)
	}
	if !gitBranchExists(dir, "new-branch") {
		t.Error("branch was not created")
	}
}

func TestGitEnsureBranch_AlreadyExists(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "already-exists")

	if err := gitEnsureBranch(dir, "already-exists", "main"); err != nil {
		t.Fatalf("gitEnsureBranch() error on existing branch: %v", err)
	}
}

func TestGitWorktreeRegistered_False(t *testing.T) {
	dir := makeTestRepo(t)
	registered, err := gitWorktreeRegistered(dir, "/some/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registered {
		t.Error("expected false for unregistered path, got true")
	}
}

func TestGitWorktreeAddAndRemove(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "wt-branch")

	wtDir := filepath.Join(t.TempDir(), "my-worktree")

	if err := gitWorktreeAdd(dir, wtDir, "wt-branch"); err != nil {
		t.Fatalf("gitWorktreeAdd() error: %v", err)
	}

	registered, err := gitWorktreeRegistered(dir, wtDir)
	if err != nil {
		t.Fatalf("gitWorktreeRegistered() error: %v", err)
	}
	if !registered {
		t.Error("worktree should be registered after add")
	}

	if err := gitWorktreeRemove(dir, wtDir); err != nil {
		t.Fatalf("gitWorktreeRemove() error: %v", err)
	}

	registered, err = gitWorktreeRegistered(dir, wtDir)
	if err != nil {
		t.Fatalf("gitWorktreeRegistered() after remove: %v", err)
	}
	if registered {
		t.Error("worktree should not be registered after remove")
	}
}

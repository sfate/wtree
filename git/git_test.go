package git

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
		if _, err := Run(dir, args...); err != nil {
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
		if _, err := Run(dir, args...); err != nil {
			t.Fatalf("repo setup: git %s: %v", strings.Join(args, " "), err)
		}
	}
	return dir
}

// gitSetup runs a git command in dir and fails the test on error.
func gitSetup(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := Run(dir, args...); err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
}

func TestRunSuccess(t *testing.T) {
	dir := makeTestRepo(t)
	out, err := Run(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("got %q, want %q", out, "true")
	}
}

func TestRunError(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := Run(dir, "not-a-real-command-xyz")
	if err == nil {
		t.Fatal("expected error for invalid git subcommand, got nil")
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := makeTestRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestLastCommitTime(t *testing.T) {
	dir := makeTestRepo(t)
	before := time.Now().Add(-5 * time.Second)

	ts, err := LastCommitTime(dir, "main")
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

func TestLastCommitTimeInvalidBranch(t *testing.T) {
	dir := makeTestRepo(t)
	_, err := LastCommitTime(dir, "no-such-branch-xyz")
	if err == nil {
		t.Fatal("expected error for non-existent branch, got nil")
	}
}

func TestFindBranchFound(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "feature-abc")

	got, err := FindBranch(dir, "feature-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feature-abc" {
		t.Errorf("FindBranch() = %q, want %q", got, "feature-abc")
	}
}

func TestFindBranchNotFound(t *testing.T) {
	dir := makeTestRepo(t)
	got, err := FindBranch(dir, "no-such-branch-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("FindBranch() = %q, want empty string", got)
	}
}

func TestBranchExistsTrue(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "exists-branch")

	if !BranchExists(dir, "exists-branch") {
		t.Error("BranchExists() = false, want true")
	}
}

func TestBranchExistsFalse(t *testing.T) {
	dir := makeTestRepo(t)
	if BranchExists(dir, "ghost-branch-xyz") {
		t.Error("BranchExists() = true, want false")
	}
}

func TestEnsureBranchCreateNew(t *testing.T) {
	dir := makeTestRepo(t)
	if err := EnsureBranch(dir, "new-branch", "main"); err != nil {
		t.Fatalf("EnsureBranch() error: %v", err)
	}
	if !BranchExists(dir, "new-branch") {
		t.Error("branch was not created")
	}
}

func TestEnsureBranchAlreadyExists(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "already-exists")

	if err := EnsureBranch(dir, "already-exists", "main"); err != nil {
		t.Fatalf("EnsureBranch() error on existing branch: %v", err)
	}
}

func TestWorktreeRegisteredFalse(t *testing.T) {
	dir := makeTestRepo(t)
	registered, err := WorktreeRegistered(dir, "/some/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registered {
		t.Error("expected false for unregistered path, got true")
	}
}

func TestWorktreeAddAndRemove(t *testing.T) {
	dir := makeTestRepo(t)
	gitSetup(t, dir, "branch", "wt-branch")

	wtDir := filepath.Join(t.TempDir(), "my-worktree")

	if err := WorktreeAdd(dir, wtDir, "wt-branch"); err != nil {
		t.Fatalf("WorktreeAdd() error: %v", err)
	}

	registered, err := WorktreeRegistered(dir, wtDir)
	if err != nil {
		t.Fatalf("WorktreeRegistered() error: %v", err)
	}
	if !registered {
		t.Error("worktree should be registered after add")
	}

	if err := WorktreeRemove(dir, wtDir); err != nil {
		t.Fatalf("WorktreeRemove() error: %v", err)
	}

	registered, err = WorktreeRegistered(dir, wtDir)
	if err != nil {
		t.Fatalf("WorktreeRegistered() after remove: %v", err)
	}
	if registered {
		t.Error("worktree should not be registered after remove")
	}
}
